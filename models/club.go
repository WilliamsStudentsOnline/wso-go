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

type GetAllClubsOptions struct {
	// Offset is ignored unless limit is supplied
	Offset *uint `json:"offset" form:"offset"`
	Limit  *uint `json:"limit" form:"limit"`

	// Unsure what other data we may want to preload in future
	//Preload []string `json:"preload" form:"preload[]"`
}

// Inserts new club into the database
func (m *ClubModel) CreateClub(p *Club) (err error) {
	err = m.DB.Create(p).Error
	if err != nil {
		return err
	}
	return
}
