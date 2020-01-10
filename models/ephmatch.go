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

// Check if two users are matching by seeing if there is already an ephmatch from that user and with the other user.
func (m *EphmatchModel) IsMatching(userID uint, otherID uint) (matching bool, err error) {
	var count int
	err = m.DB.Model(&Ephmatch{}).Where(&Ephmatch{
		UserID:  userID,
		OtherID: otherID,
	}).Count(&count).Error
	matching = count > 0
	return
}

// Check if ephmatch already exists by seeing if there is already an ephmatch from that user and with the other user.
func (m *EphmatchModel) CheckDuplicateEphmatch(userID uint, otherID uint) (duplicate bool, err error) {
	return m.IsMatching(userID, otherID)
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
// TODO: Do we want matches between profiles or users?
func (m *EphmatchModel) GetMatches(userID uint, p *[]*Ephmatch) (err error) {
	// Get matches
	err = m.DB.Model(&Ephmatch{}).
		// Join on itself to get ephmatches that actually match
		Joins("INNER JOIN ephmatches b ON b.other_id = ephmatches.user_id").
		Where("ephmatches.user_id = ? AND ephmatches.other_id = b.user_id", userID).
		// Join on users to ensure student type and visibility type
		Joins("INNER JOIN users u ON u.id = ephmatches.other_id").
		Where("u.type = ? AND u.visible = ?", UserTypeStudent, true).
		// Join on profiles for other user to ensure each
		Joins("INNER JOIN ephmatch_profiles p ON p.user_id = u.id").
		Where("p.deleted_at IS NULL").
		// Preload other column and other's ephmatch profile
		Preload("Other").
		Preload("Other.EphmatchProfile").
		Find(p).Error
	return
}

// Gets an ephmatch for every user a specific user likes.
// Gives no info about if the other user likes the specified user.
func (m *EphmatchModel) GetUserLikes(userID uint, p *[]*Ephmatch) (err error) {
	err = m.DB.Model(&Ephmatch{}).
		Where("user_id = ?", userID).
		Find(p).Error
	return
}
