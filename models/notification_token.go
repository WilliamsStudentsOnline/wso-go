package models

import (
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// NotificationToken Model
type NotificationTokenModel struct {
	*BaseModel
}

func NewNotificationTokenModel(db *gorm.DB, log *zap.SugaredLogger) *NotificationTokenModel {
	return &NotificationTokenModel{
		BaseModel: NewBaseModel(db, log),
	}
}

func (m *NotificationTokenModel) GetTokens(userID uint, t *[]*NotificationToken) (err error) {
	// Get tokens
	err = m.DB.Model(&NotificationToken{}).
		Where(NotificationToken{UserID: userID}).
		Find(t).Error
	return
}

func (m *NotificationTokenModel) CreateToken(t *NotificationToken) (err error) {
	err = m.DB.Where(NotificationToken{UserID: t.UserID, Type: t.Type, Token: t.Token}).
		FirstOrCreate(t).Error
	if err != nil {
		return err
	}
	return
}

func (m *NotificationTokenModel) GetTokensWhereSalmonNotif(t *[]*NotificationToken) (err error) {
	err = m.DB.Model(&NotificationToken{}).
		Joins("INNER JOIN notification_settings ON notification_tokens.user_id = notification_settings.user_id").
		Where("notification_settings.enable_notifications = ?", true).
		Where("notification_settings.salmon_notify = ?", true).
		Find(t).Error
	return
}

func (m *NotificationTokenModel) GetNotifTokensForUser(user *User, t *[]*NotificationToken) (err error) {
	err = m.DB.Model(&NotificationToken{}).
		Joins("INNER JOIN notification_settings ON notification_tokens.user_id = notification_settings.user_id").
		Where("notification_settings.enable_notifications = ?", true).
		Where("notification_settings.food_notify = ?", true).
		Where("notification_settings.user_id = ?", user.ID).
		Find(t).Error
	return
}

func (m *NotificationTokenModel) DeleteToken(t *NotificationToken) (err error) {
	return m.DB.Delete(t).Error
}
