package models

import "time"

// BoardBaseSchema is like BaseSchema but exposes createdAt/updatedAt in JSON
// so clients can show "last edited".
type BoardBaseSchema struct {
	ID        uint       `gorm:"primary_key" json:"id"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `sql:"index" json:"-"`
}
