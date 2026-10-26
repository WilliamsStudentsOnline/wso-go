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

func (m *TagModel) GetTagByID(id uint, tag *Tag) error {
	return m.DB.First(tag, id).Error
}

func (m *TagModel) GetTagByName(name string, tag *Tag) error {
	return m.DB.Where("name = ?", name).First(tag).Error
}

func (m *TagModel) FirstOrCreateByName(name string) (*Tag, error) {
	tag := &Tag{Name: name}
	if err := m.DB.Where(Tag{Name: name}).FirstOrCreate(tag).Error; err != nil {
		return nil, err
	}
	return tag, nil
}
