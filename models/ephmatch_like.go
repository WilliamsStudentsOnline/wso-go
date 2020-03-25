package models

import (
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// EphmatchLike Model
type EphmatchLikeModel struct {
	*BaseModel
}

func NewEphmatchLikeModel(db *gorm.DB, log *zap.SugaredLogger) *EphmatchLikeModel {
	return &EphmatchLikeModel{
		BaseModel: NewBaseModel(db, log),
	}
}

// Check if userID has liked otherID
func (m *EphmatchLikeModel) DoesLikeExist(userID uint, likedID uint) (exists bool, err error) {
	var count int
	err = m.DB.Model(&EphmatchLike{}).Where(&EphmatchLike{
		UserID:  userID,
		LikedID: likedID,
	}).Count(&count).Error
	exists = count > 0
	return
}

// Gets an ephmatch for every user a specific user likes.
// Gives no info about if the other user likes the specified user.
func (m *EphmatchLikeModel) GetUserLikes(userID uint, p *[]*EphmatchLike) (err error) {
	err = m.DB.Model(&EphmatchLike{}).
		Where(&EphmatchLike{
			UserID: userID,
		}).
		Find(p).Error
	return
}
