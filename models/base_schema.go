package models

import (
	"time"

	"github.com/jinzhu/gorm"
	funk "github.com/thoas/go-funk"
)

type BaseModel struct {
	DB *gorm.DB
}

func NewBaseModel(db *gorm.DB) *BaseModel {
	return &BaseModel{db}
}

type BaseSchema struct {
	ID        uint       `gorm:"primary_key" json:"id"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `sql:"index" json:"-"`
}

// Given a map, delete all keys not in the passed permit list
func MapPermit(m map[string]interface{}, permits ...string) {
	for key := range m {
		if !funk.ContainsString(permits, key) {
			delete(m, key)
		}
	}
}

func DeleteNilFields(m map[string]interface{}) {
	for key, value := range m {
		if value == nil {
			delete(m, key)
		}
	}
}
