package models

import "github.com/jinzhu/gorm"

// Tag Model
type TagModel struct {
	*BaseModel
}

func NewTagModel(db *gorm.DB) *TagModel {
	return &TagModel{
		BaseModel: NewBaseModel(db),
	}
}