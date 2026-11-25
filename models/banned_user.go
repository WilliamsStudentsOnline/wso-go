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

func (b *BannedUser) AnyBanned() bool {
	return b.Factrak || b.Dormtrak || b.Ephcatch || b.BulletinRead || b.BulletinWrite || b.Ephmatch
}

func (m *BannedUserModel) GetBannedUsersByUserIDs(userIDs []uint) (map[uint]*BannedUser, error) {
	bans := map[uint]*BannedUser{}
	if len(userIDs) == 0 {
		return bans, nil
	}
	var rows []*BannedUser
	if err := m.DB.Where("user_id IN (?)", userIDs).Order("id asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, r := range rows {
		if _, ok := bans[r.UserID]; !ok {
			bans[r.UserID] = r
		}
	}
	return bans, nil
}

func (m *BannedUserModel) SetBan(userID uint, ban *BannedUser) error {
	if !ban.AnyBanned() {
		return m.DB.Where("user_id = ?", userID).Delete(&BannedUser{}).Error
	}

	var existing BannedUser
	missing, err := m.GetBannedUserByID(userID, &existing)
	if err != nil {
		return err
	}
	ban.UserID = userID
	if !missing {
		ban.ID = existing.ID
		ban.CreatedAt = existing.CreatedAt
	}
	return m.DB.Save(ban).Error
}
