package models

import "github.com/jinzhu/gorm"

// Neighborhood Model
type NeighborhoodModel struct {
	*BaseModel
}

func NewNeighborhoodModel(db *gorm.DB) *NeighborhoodModel {
	return &NeighborhoodModel{
		BaseModel: NewBaseModel(db),
	}
}

// Gets all neighborhoods.
func (m *NeighborhoodModel) GetAllNeighborhoods(p *[]*Neighborhood) (err error) {
	err = m.DB.Find(p).Error
	return
}

// Gets neighborhood by its id with dorms preloaded.
func (m *NeighborhoodModel) GetNeighborhoodByID(id uint, p *Neighborhood) (err error) {
	err = m.DB.Preload("Dorms").First(p, id).Error
	return
}

func (*NeighborhoodModel) scopeTrakked(db *gorm.DB) *gorm.DB {
	return db.Where("neighborhoods.trakked = ?", true)
}
