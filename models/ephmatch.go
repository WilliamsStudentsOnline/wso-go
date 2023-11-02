package models

import (
	"errors"

	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

var (
	EphmatchModelErrorBadRelation           = errors.New("bad relation")
	EphmatchModelErrorRelationAlreadyExists = errors.New("relation with same value already exists")
)

// Ephmatch Model
type EphmatchModel struct {
	*BaseModel
}

func NewEphmatchModel(db *gorm.DB, log *zap.SugaredLogger) *EphmatchModel {
	return &EphmatchModel{
		BaseModel: NewBaseModel(db, log),
	}
}

func (m *EphmatchModel) SetRelationWithMatchHooks(userID uint, otherID uint, relation string) (matched bool, err error) {
	if !ValidateEphmatchRelation(relation) {
		return false, EphmatchModelErrorBadRelation
	}

	// Transaction setup
	tx := m.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// More transaction setup
	if err = tx.Error; err != nil {
		return false, err
	}

	// Check to see if out-relation exists.
	// If out-relation does not exists, create out-relation with value <relation>
	// If out-relation exists and is NOT <relation>, update relation to <relation>
	// If out-relation exists and is <relation>, error
	foundRelation := EphmatchRelation{}
	foundRelationErr := tx.Model(&EphmatchRelation{}).Where(&EphmatchRelation{
		UserID:  userID,
		OtherID: otherID,
	}).First(&foundRelation).Error
	if foundRelationErr != nil {
		// We got a bad error
		if !gorm.IsRecordNotFoundError(foundRelationErr) {
			tx.Rollback()
			return false, foundRelationErr
		} else {
			// Out relation doesn't exist so let's make it
			err = tx.Create(&EphmatchRelation{
				UserID:   userID,
				OtherID:  otherID,
				Relation: relation,
			}).Error
			if err != nil {
				tx.Rollback()
				return
			}
		}
	} else if foundRelation.Relation != relation {
		// If relation is found and is NOT what we want to set it to, update it to what we want
		err = tx.Model(&EphmatchRelation{}).Where(&EphmatchRelation{
			UserID:  userID,
			OtherID: otherID,
		}).Update("relation", relation).Error
		if err != nil {
			tx.Rollback()
			return
		}
	} else {
		// Relation is found and is <relation>, so let's error out
		tx.Rollback()
		return false, EphmatchModelErrorRelationAlreadyExists
	}

	// Now for the hooks:
	if relation == EphmatchRelationLike {
		matched, err = m.matchHookLikedRelation(tx, userID, otherID)
		if err != nil {
			tx.Rollback()
			return
		}
		return matched, tx.Commit().Error
	} else if relation == EphmatchRelationDislike {
		err = m.matchHookDislikedOrDeletedRelation(tx, userID, otherID)
		if err != nil {
			tx.Rollback()
			return
		}
		return false, tx.Commit().Error
	}

	return
}

func (m *EphmatchModel) DeleteRelationWithMatchHooks(userID uint, otherID uint) (err error) {
	// Transaction setup
	tx := m.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// More transaction setup
	if err = tx.Error; err != nil {
		return err
	}

	// Delete the like
	err = tx.Unscoped().Where(&EphmatchRelation{UserID: userID, OtherID: otherID}).Delete(EphmatchRelation{}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// Hook
	err = m.matchHookDislikedOrDeletedRelation(tx, userID, otherID)
	if err != nil {
		tx.Rollback()
		return
	}
	return tx.Commit().Error
}

func (m *EphmatchModel) Reset() (err error) {
	// Transaction setup
	tx := m.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// More transaction setup
	if err = tx.Error; err != nil {
		return err
	}

	// Soft delete all likes
	err = tx.Delete(EphmatchLike{}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// Soft delete all relations
	err = tx.Delete(EphmatchRelation{}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// Soft delete all matches
	err = tx.Delete(EphmatchMatch{}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// Soft delete all profiles
	err = tx.Delete(EphmatchProfile{}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (m *EphmatchModel) matchHookLikedRelation(tx *gorm.DB, userID uint, otherID uint) (matched bool, err error) {
	// We know we like them. Now check they like us. (Check like other way)
	lr := NewEphmatchRelationModel(tx, m.log)
	otherLike, err := lr.DoesLikeExist(otherID, userID)
	if err != nil {
		return
	}

	// If other user liked us, create a match!
	if otherLike {
		smallestID, largestID := orderUIntPair(userID, otherID)

		// Get old match if it exists
		var match EphmatchMatch
		foundErr := tx.Unscoped().Model(&EphmatchMatch{}).Where(&EphmatchMatch{
			UserAID: smallestID,
			UserBID: largestID,
		}).First(&match).Error

		// Then either create a new match or update the old one
		if foundErr != nil && !gorm.IsRecordNotFoundError(foundErr) {
			// If we have an error, error!
			return false, foundErr
		} else if foundErr != nil && gorm.IsRecordNotFoundError(foundErr) {
			// If we couldn't find the match, create a new one
			err = tx.Create(&EphmatchMatch{
				UserAID: smallestID,
				UserBID: largestID,
			}).Error
			if err != nil {
				return
			}
		} else {
			// If the match exists, set deleted to false and update it!
			err = tx.Unscoped().
				Model(&EphmatchMatch{}).
				Where("id = ?", match.ID).
				Update("deleted_at", nil).Error
			if err != nil {
				return
			}
		}
	}

	return otherLike, nil
}

func (m *EphmatchModel) matchHookDislikedOrDeletedRelation(tx *gorm.DB, userID uint, otherID uint) (err error) {
	// We know we dont like them (either disliked or deleted). Now delete the match if it exists.
	smallestID, largestID := orderUIntPair(userID, otherID)

	// Delete the match if it exists
	err = tx.Where(EphmatchMatch{
		UserAID: smallestID,
		UserBID: largestID,
	}).Delete(EphmatchMatch{}).Error
	if err != nil {
		return err
	}

	return
}

/*func (m *EphmatchModel) CreateLikeAndMatch(userID uint, likedID uint) (matched bool, err error) {
	// Transaction setup
	tx := m.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// More transaction setup
	if err := tx.Error; err != nil {
		return false, err
	}

	// Create like
	err = tx.Create(&EphmatchLike{
		UserID:  userID,
		LikedID: likedID,
	}).Error
	if err != nil {
		tx.Rollback()
		return
	}

	// Check like other way
	lm := NewEphmatchLikeModel(tx, m.log)
	otherLike, err := lm.DoesLikeExist(likedID, userID)
	if err != nil {
		tx.Rollback()
		return
	}

	// If other user liked us, create a match!
	if otherLike {
		smallestID, largestID := orderUIntPair(userID, likedID)

		// Get old match if it exists
		var match EphmatchMatch
		err = tx.Unscoped().Model(&EphmatchMatch{}).Where(&EphmatchMatch{
			UserAID: smallestID,
			UserBID: largestID,
		}).First(&match).Error

		// Then either create a new match or update the old one
		if err != nil && !gorm.IsRecordNotFoundError(err) {
			// If we have an error, error!
			tx.Rollback()
			return
		} else if err != nil && gorm.IsRecordNotFoundError(err) {
			// If we couldn't find the match, create a new one
			err = tx.Create(&EphmatchMatch{
				UserAID: smallestID,
				UserBID: largestID,
			}).Error
			if err != nil {
				tx.Rollback()
				return
			}
		} else {
			// If the match exists, set deleted to false and update it!
			err = tx.Unscoped().
				Model(&EphmatchMatch{}).
				Where("id = ?", match.ID).
				Update("deleted_at", nil).Error
			if err != nil {
				tx.Rollback()
				return
			}
		}

	}

	return otherLike, tx.Commit().Error
}

func (m *EphmatchModel) DeleteLikeAndMatch(userID uint, likedID uint) (err error) {
	// Transaction setup
	tx := m.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// More transaction setup
	if err := tx.Error; err != nil {
		return err
	}

	// Delete the like
	err = tx.Unscoped().Where(&EphmatchLike{UserID: userID, LikedID: likedID}).Delete(EphmatchLike{}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	smallestID, largestID := orderUIntPair(userID, likedID)

	// Delete the match if it exists
	err = tx.Where(EphmatchMatch{
		UserAID: smallestID,
		UserBID: largestID,
	}).Delete(EphmatchMatch{}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}*/
