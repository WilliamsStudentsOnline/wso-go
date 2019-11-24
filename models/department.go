package models

import (
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// Department Model
type DepartmentModel struct {
	*BaseModel
}

func NewDepartmentModel(db *gorm.DB, log *zap.SugaredLogger) *DepartmentModel {
	return &DepartmentModel{
		BaseModel: NewBaseModel(db, log),
	}
}

// Gets all departments.
func (m *DepartmentModel) GetAllDepartments(p *[]Department) (err error) {
	err = m.DB.Find(p).Error
	return
}

// Gets department by its id with areas of study preloaded.
func (m *DepartmentModel) GetDepartmentByID(id uint, p *Department) (err error) {
	err = m.DB.Preload("AreasOfStudy").First(p, id).Error
	return
}

func (m *DepartmentModel) DoesDepartmentExist(id uint) (exists bool, err error) {
	var count int
	err = m.DB.Model(&Department{}).Where("departments.id = ?", id).Count(&count).Error
	exists = count > 0
	return
}
