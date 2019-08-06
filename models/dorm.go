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

// Gets all dorms.
func (m *DormModel) GetAllDorms(p *[]*Dorm) (err error) {
	err = m.DB.Find(p).Error
	return
}

// Gets dorm by its id with neighborhood and dorm rooms preloaded.
func (m *DormModel) GetDormByID(id uint, p *Dorm) (err error) {
	err = m.DB.Preload("Neighborhood").Preload("DormRooms").First(p, id).Error
	return
}

func (m *DormModel) DoesDormExist(id uint) (exists bool, err error) {
	var count int
	err = m.DB.Model(&Dorm{}).Where("dorms.id = ?", id).Count(&count).Error
	exists = count > 0
	return
}
