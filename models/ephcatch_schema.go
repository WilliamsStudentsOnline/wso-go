package models

import "time"

// Ephcatch Schema
type Ephcatch struct {
	BaseSchema

	Seen *bool `gorm:"default:false;not null" json:"seen"`

	// Belongs to user
	User   *User `json:"user"`
	UserID uint  `json:"userID"`

	// Belongs to other
	Other   *User `json:"other"`
	OtherID uint  `json:"otherID"`

	// Pass the created time: not looked at by GORM
	CreatedTime time.Time `gorm:"-" json:"createdTime"`
}

func (*Ephcatch) TableName() string {
	return "ephcatches"
}

// Again, I hate hooks but this is the best way.
// This populates the createdTime field: please don't use this field for database updates.
func (m *Ephcatch) AfterFind() (err error) {
	m.CreatedTime = m.CreatedAt
	return
}
