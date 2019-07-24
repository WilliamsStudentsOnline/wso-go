package models

import "github.com/jinzhu/gorm"

// AreaOfStudy Model
type AreaOfStudyModel struct {
	*BaseModel
}

func NewAreaOfStudyModel(db *gorm.DB) *AreaOfStudyModel {
	return &AreaOfStudyModel{
		BaseModel: NewBaseModel(db),
	}
}
