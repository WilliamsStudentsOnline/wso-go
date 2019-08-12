package models

import (
	"strings"

	"github.com/jinzhu/gorm"
)

// AreaOfStudy Model
type AreaOfStudyModel struct {
	*BaseModel
}

func NewAreaOfStudyModel(db *gorm.DB) *AreaOfStudyModel {
	return &AreaOfStudyModel{
		BaseModel: NewBaseModel(db),
	}
}

// Gets all areas of study.
func (m *AreaOfStudyModel) GetAllAreasOfStudy(p *[]AreaOfStudy) (err error) {
	err = m.DB.Find(p).Error
	return
}

// Gets area of study by its id with department preloaded.
func (m *AreaOfStudyModel) GetAreaOfStudyByID(id uint, p *AreaOfStudy) (err error) {
	err = m.DB.Preload("Department").First(p, id).Error
	return
}

// Gets area of study by its abbreviation. NOTE: all abbreviations are uppercase.
func (m *AreaOfStudyModel) GetAreaOfStudyByAbbreviation(abbreviation string, p *AreaOfStudy) (err error) {
	err = m.DB.Where(&AreaOfStudy{Abbreviation: strings.ToUpper(abbreviation)}).First(p).Error
	return
}

func (m *AreaOfStudyModel) DoesAreaOfStudyExist(id uint) (exists bool, err error) {
	var count int
	err = m.DB.Model(&AreaOfStudy{}).Where("areas_of_study.id = ?", id).Count(&count).Error
	exists = count > 0
	return
}
