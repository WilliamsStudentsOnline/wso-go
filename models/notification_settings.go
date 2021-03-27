package models

import (
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// NotificationSettings Model
type NotificationSettingsModel struct {
	*BaseModel
}

func NewNotificationSettingsModel(db *gorm.DB, log *zap.SugaredLogger) *NotificationSettingsModel {
	return &NotificationSettingsModel{
		BaseModel: NewBaseModel(db, log),
	}
}

func (m *NotificationSettingsModel) GetSettings(userID uint, s *NotificationSettings) (err error) {
	err = m.DB.Model(&NotificationSettings{}).
		Where("notification_settings.user_id = ?", userID).
		First(s).Error
	return
}

func (m *NotificationSettingsModel) GetOrCreateSettings(userID uint, s *NotificationSettings) (err error) {
	err = m.DB.Model(&NotificationSettings{}).
		Where(NotificationSettings{UserID: userID}).
		FirstOrCreate(s).Error
	return
}

func (m *NotificationSettingsModel) UpdateSettings(s *NotificationSettings) (err error) {
	err = m.DB.Save(&s).Error
	if err != nil {
		return
	}

	err = m.DB.First(s, s.ID).Error
	return
}
