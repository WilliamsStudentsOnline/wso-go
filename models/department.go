package models

import "github.com/jinzhu/gorm"

// Department Model
type DepartmentModel struct {
	*BaseModel
}

func NewDepartmentModel(db *gorm.DB) *DepartmentModel {
	return &DepartmentModel{
		BaseModel: NewBaseModel(db),
	}
}
