package models

import (
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type ClubModel struct {
	*BaseModel
}

func NewClubModel(db *gorm.DB, log *zap.SugaredLogger) *ClubModel {
	return &ClubModel{
		BaseModel: NewBaseModel(db, log),
	}
}

// Gets all clubs.
func (m *ClubModel) GetAllClubs(p *[]*Club, opts Options) (err error) {
	db := m.DB
	if opts != nil {
		db = opts.Run(db)
	}
	err = db.Find(p).Error
	return
}

type GetAllClubsOptions struct {
	// Offset is ignored unless limit is supplied
	Offset *uint `json:"offset" form:"offset"`
	Limit  *uint `json:"limit" form:"limit"`

	// Unsure what other data we may want to preload in future
	Preload []string `json:"preload" form:"preload[]"`
}

// Preload specifically allowed parts if requested
func (o *GetAllClubsOptions) Preloader(db *gorm.DB) *gorm.DB {
	// Leaving this function blank until I figure out what or if we should preload
	//any data
	return db
}

func (o *GetAllClubsOptions) Order(db *gorm.DB) *gorm.DB {
	return db.Order("clubs.id ASC", true)
}

func (o *GetAllClubsOptions) Paginate(db *gorm.DB) *gorm.DB {
	db = o.Order(db)
	if o.Limit != nil {
		db = db.Limit(*o.Limit)
		if o.Offset != nil {
			db = db.Offset(*o.Offset)
		}
	}

	return db
}

func (o *GetAllClubsOptions) Run(db *gorm.DB) *gorm.DB {
	return o.Paginate(o.Preloader(db))
}

// Inserts new club into the database
func (m *ClubModel) CreateClub(p *Club) (err error) {
	err = m.DB.Create(p).Error
	if err != nil {
		return err
	}
	return
}

// Returns specfied Club
func (m *ClubModel) GetClubByID(id uint, c *Club, opts Options) (err error) {
	db := m.DB
	if opts != nil {
		db = opts.Run(db)
	}
	err = db.First(c, id).Error
	return
}

// Deletes club
func (m *ClubModel) DeleteClub(c *Club) (err error) {
	tx := m.DB.Begin()

	if err = tx.Error; err != nil {
		return
	}

	err = tx.Where("clubs.id = ?", c.ID).Delete(&Club{}).Error
	if err != nil {
		tx.Rollback()
		return
	}

	err = tx.Delete(c).Error
	if err != nil {
		tx.Rollback()
		return
	}

	err = tx.Commit().Error
	return
}
