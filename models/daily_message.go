package models

import (
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type DailyMessageModel struct {
	*BaseModel
}

func NewDailyMessageModel(db *gorm.DB, log *zap.SugaredLogger) *DailyMessageModel {
	return &DailyMessageModel{
		BaseModel: NewBaseModel(db, log),
	}
}

// TODO get by day
// TODO get by week
// TODO get by month
// TODO get by category
