package models

import "github.com/jinzhu/gorm"

// Dorm Model
type DormModel struct {
	*BaseModel
}

func NewDormModel(db *gorm.DB) *DormModel {
	return &DormModel{
		BaseModel: NewBaseModel(db),
	}
}