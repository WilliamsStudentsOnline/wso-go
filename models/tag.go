package models

import (
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// Tag Model
type TagModel struct {
	*BaseModel
}

func NewTagModel(db *gorm.DB, log *zap.SugaredLogger) *TagModel {
	return &TagModel{
		BaseModel: NewBaseModel(db, log),
	}
}
