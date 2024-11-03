package models

import (
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// Bearbones Club Model
// Final version final version may not import UserModel
type ClubModel struct {
	*BaseModel
}

// Inserts new club into the database
func (m *ClubModel) CreateClub(p *ClubTrak) (err error) {
	err = m.DB.Create(p).Error
	if err != nil {
		return err
	}

	err = m.DB.
		Preload("DormRoom").
		Preload("DormRoom.Dorm").
		Preload("DormRoom.Dorm.Neighborhood").
		First(p).Error
	return
}

func NewClubModel(db *gorm.DB, log *zap.SugaredLogger) *ClubModel {
	return &ClubModel{
		BaseModel: NewBaseModel(db, log),
	}
}
