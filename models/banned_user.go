package models

import (
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type BannedUserModel struct {
	*BaseModel
}

func NewBannedUserModel(db *gorm.DB, log *zap.SugaredLogger) *BannedUserModel {
	return &BannedUserModel{
		BaseModel: NewBaseModel(db, log),
	}
}

// GetBannedUserByID Gets banned user by the userid.
func (m *BannedUserModel) GetBannedUserByID(userid uint, u *BannedUser) (missing bool, err error) {
	err = m.DB.Where(&BannedUser{UserID: userid}).First(u).Error
	if gorm.IsRecordNotFoundError(err) {
		return true, nil
	}
	return
}
