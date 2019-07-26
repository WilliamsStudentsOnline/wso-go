package models

import "github.com/jinzhu/gorm"

// BulletinModel models a bulletin
type BulletinModel struct {
	*BaseModel
}

func NewBulletinModel(db *gorm.DB) *BulletinModel {
	return &BulletinModel{
		BaseModel: NewBaseModel(db),
	}
}

// CreateBulletin creates a survey
func (m *BulletinModel) CreateBulletin(b *Bulletin) (err error) {
	err = m.DB.Create(b).Error
	if err != nil {
		return err
	}

	err = m.DB.Find(b).Error
	return
}

// GetAllBulletins Returns all Bulletins
func (m *BulletinModel) GetAllBulletins(b *[]Bulletin) (err error) {
	err = m.DB.Find(b).Error
	return
}

// GetAllBulletinsByType Returns all Bulletins of a type
func (m *BulletinModel) GetAllBulletinsByType(b *[]Bulletin, bulletinType string) (err error) {
	err = m.DB.Where("type = ?", bulletinType).Find(b).Error
	return
}

// GetBulletinByID retrieves a bulletin by ID
func (m *BulletinModel) GetBulletinByID(id uint, b *Bulletin) (err error) {
	err = m.DB.Where(NewBulletinWithID(id)).First(b).Error
	return
}

// DeleteBulletinByID deletes a bulletin by ID
func (m *BulletinModel) DeleteBulletinByID(id uint, b *Bulletin) (err error) {
	err = m.DB.Delete(NewBulletinWithID(id)).Error
	return
}

// UpdateBulletin Updates the bulletin, only allowing specific keys to be passed
func (m *BulletinModel) UpdateBulletin(id uint, update map[string]interface{}) (err error) {
	dbUpdate := map[string]interface{}{
		"title":     update["title"],
		"body":      update["body"],
		"startDate": update["startDate"],
		"endDate":   update["endDate"],
	}
	DeleteNilFields(dbUpdate)

	err = m.DB.Model(NewBulletinWithID(id)).Updates(dbUpdate).Error
	return
}
