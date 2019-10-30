package models

type Ephmatcher struct {
	BaseSchema

	// Belongs to user
	UserID uint  `gorm:"not null;" json:"userID"`
	User   *User `json:"user,omitempty"`

	Gender      string `json:"gender"`
	Description string `json:"description"`

	Liked bool `gorm:"-" json:"liked"` // If me (user) has an ephmatch entry where ephmatch.other_id=users.id and ephmatch.user_id=myID
}
