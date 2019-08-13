package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

// BulletinRide Model
type BulletinRideModel struct {
	*BaseModel
}

func NewBulletinRideModel(db *gorm.DB) *BulletinRideModel {
	return &BulletinRideModel{
		BaseModel: NewBaseModel(db),
	}
}

type GetAllBulletinRidesOptions struct {
	// Pagination
	Offset *time.Time `json:"offset" form:"offset"`
	Limit  *uint      `json:"limit" form:"limit"`

	// What to preload
	Preload []string `json:"preload" form:"preload"`

	// Pass a specific type (request, offer) to get just either requests or offers
	Type *string `json:"type" form:"type"`

	// Get all rides, rather than just rides that the date hasn't passed yet.
	// By default, this is false.
	All bool `json:"all" form:"all"`
}

func (p *GetAllBulletinRidesOptions) Order(db *gorm.DB) *gorm.DB {
	return db.Order("bulletin_rides.date desc", true)
}

// Pagination starts at most recent and goes down from there
func (p *GetAllBulletinRidesOptions) Paginate(db *gorm.DB) *gorm.DB {
	db = p.Order(db)
	if p.Offset != nil {
		db = db.Where("bulletin_rides.date < ?", *p.Offset)
	}
	if p.Limit != nil {
		db = db.Limit(*p.Limit)
	}
	return db
}

// Preload specifically allowed parts if requested
func (p *GetAllBulletinRidesOptions) Preloader(db *gorm.DB) *gorm.DB {
	if p.Preload != nil {
		if stringsContains(p.Preload, "user") {
			db = db.Preload("User")
		}
	}

	return db
}

func (p *GetAllBulletinRidesOptions) Run(db *gorm.DB) *gorm.DB {
	db = p.Paginate(db)
	db = p.Preloader(db)

	if !p.All {
		now := time.Now()
		db = db.Where("bulletin_rides.date >= ?", time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()))
	}

	if p.Type != nil {
		if *p.Type == "request" {
			db = db.Where("bulletin_rides.offer = ?", false)
		} else if *p.Type == "offer" {
			db = db.Where("bulletin_rides.offer = ?", true)
		}
	}

	return db
}

func (m *BulletinRideModel) GetAllRides(r *[]*BulletinRide, opts Options) (err error) {
	db := m.DB
	if opts != nil {
		db = opts.Run(db)
	}
	err = db.Find(r).Error
	return
}

func (m *BulletinRideModel) GetRideByID(id uint, r *BulletinRide) (err error) {
	err = m.DB.Preload("User").First(r, id).Error
	return
}

func (m *BulletinRideModel) CreateRide(r *BulletinRide) (err error) {
	err = m.DB.Create(r).Error
	if err != nil {
		return err
	}

	err = m.DB.Find(r).Error
	return
}

func (m *BulletinRideModel) UpdateRide(r *BulletinRide) (err error) {
	err = m.DB.Save(r).Error
	return
}

func (m *BulletinRideModel) DeleteRide(r *BulletinRide) (err error) {
	err = m.DB.Delete(r).Error
	return
}
