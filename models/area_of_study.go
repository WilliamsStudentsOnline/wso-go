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

// Gets all departments.
func (m *AreaOfStudyModel) GetAllAreasOfStudy(p *[]AreaOfStudy) (err error) {
	err = m.DB.Find(p).Error
	return
}

// Gets department by its id with areas of study preloaded.
func (m *AreaOfStudyModel) GetAreaOfStudyByID(id uint, p *AreaOfStudy) (err error) {
	err = m.DB.Preload("Department").First(p, id).Error
	return
}

func (m *AreaOfStudyModel) DoesAreaOfStudyExist(id uint) (exists bool, err error) {
	var count int
	err = m.DB.Model(&AreaOfStudy{}).Where("areas_of_study.id = ?", id).Count(&count).Error
	exists = count > 0
	return
}
