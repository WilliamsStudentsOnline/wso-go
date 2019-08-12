package models

import "github.com/jinzhu/gorm"

// Ephcatch Model
type EphcatchModel struct {
	*BaseModel
}

func NewEphcatchModel(db *gorm.DB) *EphcatchModel {
	return &EphcatchModel{
		BaseModel: NewBaseModel(db),
	}
}

// Check if ephcatch already exists by seeing if there is already an ephcatch from that user and with the other user.
func (m *EphcatchModel) CheckDuplicateEphcatch(userID uint, otherID uint) (duplicate bool, err error) {
	var count int
	err = m.DB.Model(&Ephcatch{}).Where(&Ephcatch{
		UserID:  userID,
		OtherID: otherID,
	}).Count(&count).Error
	duplicate = count > 0
	return
}

func (m *EphcatchModel) CreateEphcatchWithUserOther(userID uint, otherID uint) (err error) {
	err = m.DB.Create(&Ephcatch{
		UserID:  userID,
		OtherID: otherID,
	}).Error
	return
}

func (m *EphcatchModel) DeleteEphcatchWithUserOther(userID uint, otherID uint) (err error) {
	err = m.DB.Unscoped().Delete(&Ephcatch{
		UserID:  userID,
		OtherID: otherID,
	}).Error
	return
}
