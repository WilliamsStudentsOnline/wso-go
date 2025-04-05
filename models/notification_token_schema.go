package models

const (
	NotificationTokenTypeIOS      = "ios"
	NotificationTokenTypeFirebase = "firebase"
)

func ValidateNotificationTokenType(str string) bool {
	switch str {
	case NotificationTokenTypeIOS, NotificationTokenTypeFirebase:
		return true
	default:
		return false
	}
}

// NotificationToken Schema
type NotificationToken struct {
	BaseSchema

	// Belongs to user
	UserID uint  `gorm:"index:index_notification_tokens_on_user_id;not null" json:"userID"`
	User   *User `json:"user,omitempty"`

	Type  string `json:"type"`
	Token string `json:"token"`
}

func (*NotificationToken) TableName() string {
	return "notification_tokens"
}
