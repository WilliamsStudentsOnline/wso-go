package models

import (
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type ClubModel struct {
	*BaseModel
}

func NewClubModel(db *gorm.DB, log *zap.SugaredLogger) *ClubModel {
	return &ClubModel{
		BaseModel: NewBaseModel(db, log),
	}
}

//Database Functions

// Inserts new club into the database
func (m *ClubModel) CreateClub(p *Club) (err error) {
	err = m.DB.Create(p).Error
	if err != nil {
		return err
	}
	return
}
