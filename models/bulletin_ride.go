package models

import (
	"time"

	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// BulletinRide Model
type BulletinRideModel struct {
	*BaseModel
}

func NewBulletinRideModel(db *gorm.DB, log *zap.SugaredLogger) *BulletinRideModel {
	return &BulletinRideModel{
		BaseModel: NewBaseModel(db, log),
	}
}

type GetAllBulletinRidesOptions struct {
	// Pagination
	Start *time.Time `json:"start" form:"start"`
	// Offset is ignored unless limit is supplied
	Offset *uint `json:"offset" form:"offset"`
	Limit  *uint `json:"limit" form:"limit"`

	// What to preload
	Preload []string `json:"preload" form:"preload[]"`

	// Pass a specific type (request, offer) to get just either requests or offers
	Type *string `json:"type" form:"type"`

	// Get all rides, rather than just rides that the date hasn't passed yet.
	// By default, this is false.
	All bool `json:"all" form:"all"`
}

func (o *GetAllBulletinRidesOptions) Order(db *gorm.DB) *gorm.DB {
	return db.Order("bulletin_rides.date desc", true)
}

// Pagination starts at most recent and goes down from there
func (o *GetAllBulletinRidesOptions) Paginate(db *gorm.DB) *gorm.DB {
	db = o.Order(db)
	if o.Start != nil {
		db = db.Where("bulletin_rides.date < ?", *o.Start)
	}
	if o.Limit != nil {
		db = db.Limit(*o.Limit)
		if o.Offset != nil {
			db = db.Offset(*o.Offset)
		}
	}
	return db
}

// Preload specifically allowed parts if requested
func (o *GetAllBulletinRidesOptions) Preloader(db *gorm.DB) *gorm.DB {
	if o.Preload == nil {
		return db
	}

	if stringsContains(o.Preload, "user") {
		db = db.Preload("User")
	}

	return db
}

func (o *GetAllBulletinRidesOptions) Run(db *gorm.DB) *gorm.DB {
	db = o.Paginate(db)
	db = o.Preloader(db)
	db = o.filter(db)

	return db
}

func (o *GetAllBulletinRidesOptions) filter(db *gorm.DB) *gorm.DB {
	if !o.All {
		now := time.Now()
		db = db.Where("bulletin_rides.date >= ?", time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()))
	}

	if o.Type != nil {
		if *o.Type == "request" {
			db = db.Where("bulletin_rides.offer = ?", false)
		} else if *o.Type == "offer" {
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

func (m *BulletinRideModel) CountAllRides(opts *GetAllBulletinRidesOptions) (count int, err error) {
	db := m.DB.Model(&BulletinRide{})
	if opts != nil {
		db = opts.filter(db)
	}
	err = db.Count(&count).Error
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
