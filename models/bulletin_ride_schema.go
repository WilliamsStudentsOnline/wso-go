package models

import "time"

// BulletinRide Schema
type BulletinRide struct {
	BaseSchema

	Body        string    `gorm:"size:65535" json:"body"`
	Date        time.Time `json:"date"`
	Offer       *bool     `gorm:"not null;" json:"offer"`
	Source      string    `gorm:"not null;" json:"source"`
	Destination string    `gorm:"not null;" json:"destination"`

	// Belongs to user
	UserID uint  `json:"userID"`
	User   *User `json:"user,omitempty"`
}

func (*BulletinRide) TableName() string {
	return "bulletin_rides"
}
