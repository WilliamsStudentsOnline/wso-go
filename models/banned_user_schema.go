package models

// BannedUser Schema
type BannedUser struct {
	BaseSchema

	// Belongs to user
	User   *User `json:"user"`
	UserID uint  `gorm:"index:index_banned_users_on_user_id;not null;" json:"userID"`

	// Reason why user was banned
	Reason string `json:"reason"`

	// Service access: true means allowed to access if would normally get access, false means scope will not be granted
	Factrak       bool `gorm:"default:true" json:"factrak"`
	Dormtrak      bool `gorm:"default:true" json:"dormtrak"`
	Ephcatch      bool `gorm:"default:true" json:"ephcatch"`
	BulletinRead  bool `gorm:"default:true" json:"bulletinRead"`
	BulletinWrite bool `gorm:"default:true" json:"bulletinWrite"`
	Ephmatch      bool `gorm:"default:true" json:"ephmatch"`
}

func (*BannedUser) TableName() string {
	return "banned_users"
}
