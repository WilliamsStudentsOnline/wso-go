package models

// BannedUser Schema
type BannedUser struct {
	BaseSchema

	// Belongs to user
	User   *User `json:"user"`
	UserID uint  `gorm:"index:index_banned_users_on_user_id;not null;" json:"userID"`

	// Reason why user was banned
	Reason string `json:"reason"`

	// If banned from this service. True means user is banned; false means user is not banned
	Factrak       bool `gorm:"default:false" json:"factrak"`
	Dormtrak      bool `gorm:"default:false" json:"dormtrak"`
	Ephcatch      bool `gorm:"default:false" json:"ephcatch"`
	BulletinRead  bool `gorm:"default:false" json:"bulletinRead"`
	BulletinWrite bool `gorm:"default:false" json:"bulletinWrite"`
	Ephmatch      bool `gorm:"default:false" json:"ephmatch"`
}

func (*BannedUser) TableName() string {
	return "banned_users"
}
