package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

// BulletinModel models a bulletin
type BulletinModel struct {
	*BaseModel
}

func NewBulletinModel(db *gorm.DB) *BulletinModel {
	return &BulletinModel{
		BaseModel: NewBaseModel(db),
	}
}

// GetAllBulletins Returns all Bulletins
func (m *BulletinModel) GetAllBulletins(b *[]*Bulletin) (err error) {
	err = m.GetAllBulletinsWithOptions(b, &GetAllBulletinsOptions{})
	return
}

type GetAllBulletinsOptions struct {
	// Pagination
	Offset *time.Time `json:"offset" form:"offset"`
	Limit  *uint      `json:"limit" form:"limit"`

	// What to preload
	Preload []string `json:"preload" form:"preload"`

	// Pass a specific bulletin type to get just those bulletins
	Type *string `json:"type" form:"type"`

	// Get all bulletins, rather than just bulletins that have already started, and have not ended yet.
	// By default, this is false.
	All bool `json:"all" form:"all"`
}

func (p *GetAllBulletinsOptions) Order(db *gorm.DB) *gorm.DB {
	return db.Order("bulletins.start_date desc", true)
}

// Pagination starts at most recent and goes down from there
func (p *GetAllBulletinsOptions) Paginate(db *gorm.DB) *gorm.DB {
	db = p.Order(db)
	if p.Offset != nil {
		db = db.Where("bulletins.start_date < ?", *p.Offset)
	}
	if p.Limit != nil {
		db = db.Limit(*p.Limit)
	}
	return db
}

// Preload specifically allowed parts if requested
func (p *GetAllBulletinsOptions) Preloader(db *gorm.DB) *gorm.DB {
	if p.Preload != nil {
		if stringsContains(p.Preload, "user") {
			db = db.Preload("User")
		}
	}

	return db
}

func (p *GetAllBulletinsOptions) Run(db *gorm.DB) *gorm.DB {
	db = p.Paginate(db)
	db = p.Preloader(db)

	if !p.All {
		db = db.Where("bulletins.start_date <= ?", time.Now())
		db = db.Where("bulletins.end_date IS NULL OR bulletins.end_date > ?", time.Now())
	}

	if p.Type != nil {
		if *p.Type == BulletinTypeAnnouncement || *p.Type == BulletinTypeExchange || *p.Type == BulletinTypeJob ||
			*p.Type == BulletinTypeLostAndFound {
			db = db.Where("bulletins.type = ?", p.Type)
		}
	}

	return db
}

func (m *BulletinModel) GetAllBulletinsWithOptions(b *[]*Bulletin, opts Options) (err error) {
	db := m.DB
	if opts != nil {
		db = opts.Run(db)
	}
	err = db.Find(b).Error
	return
}

// GetAllBulletinsByType Returns all Bulletins of a type
func (m *BulletinModel) GetAllBulletinsByType(b *[]*Bulletin, bulletinType string) (err error) {
	err = m.GetAllBulletinsWithOptions(b, &GetAllBulletinsOptions{
		Type: &bulletinType,
	})
	return
}

// GetBulletinByID retrieves a bulletin by ID
func (m *BulletinModel) GetBulletinByID(id uint, b *Bulletin) (err error) {
	err = m.DB.Preload("User").Where(NewBulletinWithID(id)).First(b).Error
	return
}

// CreateBulletin creates a bulletin
func (m *BulletinModel) CreateBulletin(b *Bulletin) (err error) {
	err = m.DB.Create(b).Error
	if err != nil {
		return err
	}

	err = m.DB.Find(b).Error
	return
}

// UpdateBulletin Updates the bulletin, only allowing specific keys to be passed
func (m *BulletinModel) UpdateBulletin(b *Bulletin) (err error) {
	err = m.DB.Save(b).Error
	return
}

// DeleteBulletin deletes a bulletin
func (m *BulletinModel) DeleteBulletin(b *Bulletin) (err error) {
	err = m.DB.Delete(b).Error
	return
}
