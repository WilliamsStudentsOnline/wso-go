package models

// Ephmatch Match Schema
type EphmatchMatch struct {
	BaseSchema

	// Belongs to user A (only in DB, not JSON)
	UserA   *User `gorm:"foreignkey:UserAID" json:"-"`
	UserAID uint  `gorm:"column:user_a_id" json:"-"`

	// Belongs to user B (only in DB, not JSON)
	UserB   *User `gorm:"foreignkey:UserBID" json:"-"`
	UserBID uint  `gorm:"column:user_b_id" json:"-"`

	// The user that self matched with. We dont report self, as that should be obvious.
	MatchedUser   *User `gorm:"-" json:"matchedUser"`
	MatchedUserID uint  `gorm:"-" json:"matchedUserID"`
}

func (*EphmatchMatch) TableName() string {
	return "ephmatch_matches"
}
