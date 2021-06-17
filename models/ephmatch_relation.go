package models

import (
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

/*
Terminology:

Out Relation - a relation FROM user-self TO another user (eg. I like Maud)
In Relation - a relation TO user-self FROM another user (eg. Maud likes me)
*/

// EphmatchRelation Model
type EphmatchRelationModel struct {
	*BaseModel
}

func NewEphmatchRelationModel(db *gorm.DB, log *zap.SugaredLogger) *EphmatchRelationModel {
	return &EphmatchRelationModel{
		BaseModel: NewBaseModel(db, log),
	}
}

// Check if userID has liked otherID
func (m *EphmatchRelationModel) DoesLikeExist(userID uint, likedID uint) (exists bool, err error) {
	var count int
	err = m.DB.Model(&EphmatchRelation{}).Where(&EphmatchRelation{
		UserID:   userID,
		OtherID:  likedID,
		Relation: EphmatchRelationLike,
	}).Count(&count).Error
	exists = count > 0
	return
}

// Check if userID has a out-relation with otherID
func (m *EphmatchRelationModel) DoesOutRelationExist(userID uint, otherID uint) (exists bool, err error) {
	var count int
	err = m.DB.Model(&EphmatchRelation{}).Where(&EphmatchRelation{
		UserID:  userID,
		OtherID: otherID,
	}).Count(&count).Error
	exists = count > 0
	return
}

// Check if userID has a out-relation with otherID
func (m *EphmatchRelationModel) GetOutRelation(userID uint, otherID uint) (exists bool, relation string, err error) {
	rel := EphmatchRelation{}
	err = m.DB.Model(&EphmatchRelation{}).Where(&EphmatchRelation{
		UserID:  userID,
		OtherID: otherID,
	}).First(&rel).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return false, "", nil
		} else {
			return false, "", err
		}
	}

	return true, rel.Relation, nil
}

// Gets an ephmatch like for every user a specific user likes.
// Gives no info about if the other user likes the specified user.
func (m *EphmatchRelationModel) GetUserLikes(userID uint, p *[]*EphmatchRelation) (err error) {
	err = m.DB.Model(&EphmatchRelation{}).
		Where(&EphmatchRelation{
			UserID:   userID,
			Relation: EphmatchRelationLike,
		}).
		Find(p).Error
	return
}

// Gets an ephmatch relation for every user a specific user sets a relation about.
// Gives no info about if the other user has a relation with the specified user.
func (m *EphmatchRelationModel) GetUserOutRelations(userID uint, p *[]*EphmatchRelation) (err error) {
	err = m.DB.Model(&EphmatchRelation{}).
		Where(&EphmatchRelation{
			UserID: userID,
		}).
		Find(p).Error
	return
}

// Gets an ephmatch relation for every user that set a relation for the specific user. (eg. In-Relations)
// Gives no info about if the user set relation to the relation-setters.
func (m *EphmatchRelationModel) GetUserInRelations(userID uint, p *[]*EphmatchRelation) (err error) {
	err = m.DB.Model(&EphmatchRelation{}).
		Where(&EphmatchRelation{
			OtherID: userID,
		}).
		Find(p).Error
	return
}

// Gets an ephmatch like-relation for every user that liked a specific user. (eg. In-Like-Relations)
// Gives no info about if the user likes the admirers.
func (m *EphmatchRelationModel) GetUserAdmirers(userID uint, p *[]*EphmatchRelation) (err error) {
	err = m.DB.Model(&EphmatchRelation{}).
		Where(&EphmatchRelation{
			OtherID:  userID,
			Relation: EphmatchRelationLike,
		}).
		Find(p).Error
	return
}
