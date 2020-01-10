package models

// Ephmatch Schema
type Ephmatch struct {
	BaseSchema

	Seen *bool `gorm:"default:false;not null" json:"seen"`

	// Belongs to user
	User   *User `json:"user"`
	UserID uint  `json:"userID"`

	// Belongs to other
	Other   *User `json:"other"`
	OtherID uint  `json:"otherID"`
}

func (*Ephmatch) TableName() string {
	return "ephmatches"
}
