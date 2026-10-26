package models

import (
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type ModeratorTaskModel struct {
	*BaseModel
}

func NewModeratorTaskModel(db *gorm.DB, log *zap.SugaredLogger) *ModeratorTaskModel {
	return &ModeratorTaskModel{
		BaseModel: NewBaseModel(db, log),
	}
}
