package models

// Bulletin Model
type BulletinModel struct {
	BaseModel
}

// Returns all Bulletins
func (m *BulletinModel) GetAllBulletins(b *[]Bulletin) (err error) {
	err = m.DB.Find(b).Error
	return
}

// Returns all Bulletins of a type TODO
func (m *BulletinModel) GetAllBulletinsByType(b *[]Bulletin, bulletinType string) (err error) {
	err = m.DB.Where("type = ?", bulletinType).Find(b).Error
	return
}

func (m *BulletinModel) GetBulletinByID(id uint, b *Bulletin) (err error) {
	err = m.DB.Where(NewBulletinWithID(id)).First(b).Error
	return
}

func (m *BulletinModel) DeleteBulletinByID(id uint, b *Bulletin) (err error) {
	err = m.DB.Delete(NewBulletinWithID(id)).Error
	return
}

// Update the bulletin. Only allow specific keys to be passed
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
