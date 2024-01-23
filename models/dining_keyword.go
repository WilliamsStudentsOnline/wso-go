package models

import (
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// Dining keyword Model
type DiningKeywordModel struct {
	*BaseModel
}

func NewDiningKeywordModel(db *gorm.DB, log *zap.SugaredLogger) *DiningKeywordModel {
	return &DiningKeywordModel{
		BaseModel: NewBaseModel(db, log),
	}
}

func (m *DiningKeywordModel) GetAllKeywords(t *[]*DiningKeyword) (err error) {
	// Get all keywords
	err = m.DB.Find(t).Error
	return
}

func (m *DiningKeywordModel) CreateKeyword(t *DiningKeyword) (err error) {
	err = m.DB.Where(DiningKeyword{Keyword: t.Keyword}).
		FirstOrCreate(t).Error
	if err != nil {
		return err
	}
	return
}

func (m *DiningKeywordModel) NewKeywordAssociation(t *DiningKeyword, u *User) (err error) {
	err = m.DB.Model(t).Association("Users").Append(u).Error
	if err != nil {
		return err
	}
	return
}

func (m *DiningKeywordModel) DeleteKeywordAssociation(t *DiningKeyword, u *User) (err error) {
	err = m.DB.Model(t).Association("Users").Delete(u).Error
	if err != nil {
		return err
	}
	return
}

func (m *DiningKeywordModel) DeleteKeyword(t *DiningKeyword) (err error) {
	return m.DB.Delete(t).Error
}
