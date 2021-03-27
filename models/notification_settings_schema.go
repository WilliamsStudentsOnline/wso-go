package models

// NotificationSettings Schema
type NotificationSettings struct {
	BaseSchema

	// Belongs to user
	UserID uint  `gorm:"index:index_notification_settings_on_user_id;not null" json:"userID"`
	User   *User `json:"user,omitempty"`

	// Settings:
	EnableNotifications bool `json:"enableNotifications"`
	SalmonNotify        bool `json:"salmonNotify"`
}

func (*NotificationSettings) TableName() string {
	return "notification_settings"
}
