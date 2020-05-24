package models

type EphmatchProfile struct {
	BaseSchema

	// Belongs to user
	UserID uint  `gorm:"index:index_ephmatch_profiles_on_user_id;not null;" json:"userID"`
	User   *User `json:"user,omitempty"`

	// These are optional
	Description  *string `json:"description"`
	MatchMessage *string `json:"matchMessage"`

	Liked   *bool `gorm:"-" json:"liked,omitempty"`   // If self has liked this profile (user)
	Matched *bool `gorm:"-" json:"matched,omitempty"` // If user and self are matched

	// Current location columns
	LocationVisible *bool   `gorm:"DEFAULT:true;not null" json:"locationVisible"`
	LocationTown    *string `json:"locationTown"`
	LocationState   *string `json:"locationState"`
	LocationCountry *string `json:"locationCountry"`

	// Non db entry that acts as a flag for deleted_at column
	Deleted bool `gorm:"-" json:"deleted"`
}

func (*EphmatchProfile) TableName() string {
	return "ephmatch_profiles"
}

// Again, I hate hooks but this is the best way.
// This populates the deleted flag on the profile
func (p *EphmatchProfile) AfterFind() (err error) {
	p.Deleted = p.DeletedAt != nil
	return
}
