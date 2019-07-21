package models

import "github.com/jinzhu/gorm"

type Neighborhood struct {
	BaseSchema
	Name string `json:"name"`
	Dorms []Dorm `json:"dorms"`
}

func (*Neighborhood) TableName() string {
	return "neighborhoods"
}

func (*Neighborhood) ScopeTrakked(db *gorm.DB) *gorm.DB {
	return db.Where("amount > ?", 1000)
}
