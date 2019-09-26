package models

import "github.com/jinzhu/gorm"

// Ephmatch Model
type EphmatchModel struct {
	*BaseModel
}

func NewEphmatchModel(db *gorm.DB) *EphmatchModel {
	return &EphmatchModel{
		BaseModel: NewBaseModel(db),
	}
}

// Check if ephmatch already exists by seeing if there is already an ephmatch from that user and with the other user.
func (m *EphmatchModel) CheckDuplicateEphmatch(userID uint, otherID uint) (duplicate bool, err error) {
	var count int
	err = m.DB.Model(&Ephmatch{}).Where(&Ephmatch{
		UserID:  userID,
		OtherID: otherID,
	}).Count(&count).Error
	duplicate = count > 0
	return
}

func (m *EphmatchModel) CreateEphmatchWithUserOther(userID uint, otherID uint) (err error) {
	err = m.DB.Create(&Ephmatch{
		UserID:  userID,
		OtherID: otherID,
	}).Error
	return
}

func (m *EphmatchModel) DeleteEphmatchWithUserOther(userID uint, otherID uint) (err error) {
	err = m.DB.Unscoped().Where("user_id = ? AND other_id = ?", userID, otherID).Delete(Ephmatch{}).Error
	return
}

// Get ephmatch matches of user.
func (m *EphmatchModel) GetMatches(userID uint, p *[]*Ephmatch) (err error) {
	// Get matches
	err = m.DB.Model(&Ephmatch{}).
		Joins("INNER JOIN ephmatches b ON b.other_id = ephmatches.user_id").
		Where("ephmatches.user_id = ? AND ephmatches.other_id = b.user_id", userID).
		Joins("INNER JOIN users u ON u.id = ephmatches.other_id").
		Where("u.type = ? AND u.visible = ? AND u.opt_out_ephcatch = ?", UserTypeStudent, true, false).
		Preload("Other").
		Find(p).Error
	//"type = ? AND visible = ? AND opt_out_ephmatch = ?", UserTypeStudent, true, false
	return
}
