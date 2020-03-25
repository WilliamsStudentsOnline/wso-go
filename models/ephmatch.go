package models

import (
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
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

func (m *EphmatchModel) CreateLikeAndMatch(userID uint, likedID uint) (matched bool, err error) {
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
		// Get old match if it exists
		var match EphmatchMatch
		err = tx.Unscoped().Model(&EphmatchMatch{}).Where(&EphmatchMatch{
			UserAID: userID,
			UserBID: likedID,
		}).Or(&EphmatchMatch{
			UserAID: likedID,
			UserBID: userID,
		}).First(&match).Error

		// Then either create a new match or update the old one
		if err != nil && !gorm.IsRecordNotFoundError(err) {
			// If we have an error, error!
			tx.Rollback()
			return
		} else if err != nil && gorm.IsRecordNotFoundError(err) {
			// If we couldn't find the match, create a new one
			err = tx.Create(&EphmatchMatch{
				UserAID: userID,
				UserBID: likedID,
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

	// Delete the match if it exists
	err = tx.Where(EphmatchMatch{
		UserAID: userID,
		UserBID: likedID,
	}).Or(EphmatchMatch{
		UserAID: likedID,
		UserBID: userID,
	}).Delete(EphmatchMatch{}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
