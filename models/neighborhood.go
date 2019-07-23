package models

import "github.com/jinzhu/gorm"

const (
	NeighborhoodFirstYear = "First-year"
	NeighborhoodCoop      = "Co-op"
)

// Neighborhood Model
type NeighborhoodModel struct {
	BaseModel
}

func (*NeighborhoodModel) scopeTrakked(db *gorm.DB) *gorm.DB {
	return db.Not("name = ?", NeighborhoodFirstYear).Not("name = ?", NeighborhoodCoop)
}
