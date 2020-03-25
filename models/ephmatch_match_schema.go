package models

// Ephmatch Match Schema
type EphmatchMatch struct {
	BaseSchema

	// In order to have efficient search, UserA and UserB are sorted by id such that
	// UserA will always have an ID smaller than UserB
	// That is:
	// UserAID = min{userA.ID, userB.ID}
	// UserBID = max{userA.ID, userB.ID}

	// Belongs to user A, whoever has the smallest ID (only in DB, not JSON)
	UserA   *User `gorm:"foreignkey:UserAID" json:"-"`
	UserAID uint  `gorm:"column:user_a_id;index:index_ephmatch_matches_on_user_a_id;not null;" json:"-"`

	// Belongs to user B, whoever has the largest ID (only in DB, not JSON)
	UserB   *User `gorm:"foreignkey:UserBID" json:"-"`
	UserBID uint  `gorm:"column:user_b_id;index:index_ephmatch_matches_on_user_b_id;not null;" json:"-"`

	// The user that self matched with. We dont report self, as that should be obvious.
	MatchedUser   *User `gorm:"-" json:"matchedUser"`
	MatchedUserID uint  `gorm:"-" json:"matchedUserID"`
}

func (*EphmatchMatch) TableName() string {
	return "ephmatch_matches"
}
