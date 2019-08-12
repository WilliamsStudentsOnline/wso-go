package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

// DormtrakReview Model
type DormtrakReviewModel struct {
	*BaseModel
}

func NewDormtrakReviewModel(db *gorm.DB) *DormtrakReviewModel {
	return &DormtrakReviewModel{
		BaseModel: NewBaseModel(db),
	}
}

// Gets all reviews.
func (m *DormtrakReviewModel) GetAllReviews(p *[]*DormtrakReview) (err error) {
	err = m.GetAllReviewsWithOptions(p, nil)
	return
}

// Get a review by the id. Preload the dorm room, dorm, and neighborhood
func (m *DormtrakReviewModel) GetReviewByID(id uint, p *DormtrakReview) (err error) {
	err = m.DB.Preload("DormRoom").
		Preload("DormRoom.Dorm").
		Preload("DormRoom.Dorm.Neighborhood").
		First(p, id).Error
	return
}

// Check if review already exists by seeing if there is already a review with that user and dorm room.
func (m *DormtrakReviewModel) CheckDuplicateReview(userID uint, dormRoomID uint) (duplicate bool, err error) {
	var count int
	err = m.DB.Model(&DormtrakReview{}).Where(&DormtrakReview{
		UserID:     userID,
		DormRoomID: dormRoomID,
	}).Count(&count).Error
	duplicate = count > 0
	return
}

// Create a new review.
func (m *DormtrakReviewModel) CreateReview(p *DormtrakReview) (err error) {
	err = m.DB.Create(p).Error
	if err != nil {
		return err
	}

	err = m.DB.
		Preload("DormRoom").
		Preload("DormRoom.Dorm").
		Preload("DormRoom.Dorm.Neighborhood").
		First(p).Error
	return
}

func (m *DormtrakReviewModel) UpdateReview(p *DormtrakReview) (err error) {
	err = m.DB.Save(&p).Error
	if err != nil {
		return
	}

	err = m.DB.
		Preload("DormRoom").
		Preload("DormRoom.Dorm").
		Preload("DormRoom.Dorm.Neighborhood").
		First(p, p.ID).Error
	return
}

func (m *DormtrakReviewModel) DeleteReview(p *DormtrakReview) (err error) {
	err = m.DB.Delete(p).Error
	return
}

type GetAllDormtrakReviewsOptions struct {
	// Scopes for specific areas
	DormID     *uint `json:"dormID" form:"dormID"`
	DormRoomID *uint `json:"dormRoomID" form:"dormRoomID"`
	UserID     *uint `json:"userID" form:"userID"`

	// Pagination
	Offset *time.Time `json:"offset" form:"offset"`
	Limit  *uint      `json:"limit" form:"limit"`

	// What to preload
	Preload []string `json:"preload" form:"preload"`

	// Scope to only get commented reviews. True means only get commented; false/empty means ignore this scope.
	Commented bool `json:"commented" form:"commented"`
}

func (p *GetAllDormtrakReviewsOptions) Order(db *gorm.DB) *gorm.DB {
	return db.Order("dormtrak_reviews.created_at desc")
}

// Pagination starts at most recent and goes down from there
func (p *GetAllDormtrakReviewsOptions) Paginate(db *gorm.DB) *gorm.DB {
	db = p.Order(db)
	if p.Offset != nil {
		db = db.Where("dormtrak_reviews.created_at < ?", *p.Offset)
	}
	if p.Limit != nil {
		db = db.Limit(*p.Limit)
	}
	return db
}

// Preload specifically allowed parts if requested
func (p *GetAllDormtrakReviewsOptions) Preloader(db *gorm.DB) *gorm.DB {
	if p.Preload != nil {
		if stringsContains(p.Preload, "dormRoom") {
			db = db.Preload("DormRoom")
		}
		if stringsContains(p.Preload, "dorm") {
			db = db.Preload("DormRoom.Dorm")
		}
		if stringsContains(p.Preload, "neighborhood") {
			db = db.Preload("DormRoom.Dorm.Neighborhood")
		}
	}

	return db
}

// Gets all reviews.
func (m *DormtrakReviewModel) GetAllReviewsWithOptions(p *[]*DormtrakReview, opts *GetAllDormtrakReviewsOptions) (err error) {
	scopes := []func(*gorm.DB) *gorm.DB{
		m.scopeDefault,
	}
	db := m.DB

	if opts != nil {
		db = opts.Paginate(db)
		db = opts.Preloader(db)
		if opts.DormID != nil {
			scopes = append(scopes, m.withDormID(*opts.DormID))
		}
		if opts.DormRoomID != nil {
			scopes = append(scopes, m.withDormRoomID(*opts.DormRoomID))
		}
		if opts.UserID != nil {
			scopes = append(scopes, m.withUserID(*opts.UserID))
		}
		if opts.Commented {
			scopes = append(scopes, m.scopeCommented)
		}
	}

	err = db.Scopes(scopes...).Find(p).Error
	return
}

func (*DormtrakReviewModel) withDormID(dormID uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Joins("JOIN dorm_rooms ON dorm_rooms.id = dormtrak_reviews.dorm_room_id").Where("dorm_rooms.dorm_id = ?", dormID)
	}
}

func (*DormtrakReviewModel) withDormRoomID(dormRoomID uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(&DormtrakReview{DormRoomID: dormRoomID})
	}
}

func (*DormtrakReviewModel) withUserID(userID uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(&DormtrakReview{UserID: userID})
	}
}

func (m *DormtrakReviewModel) scopeDefault(db *gorm.DB) *gorm.DB {
	return m.scopeOrderDefault(db)
}

func (*DormtrakReviewModel) scopeOrderDefault(db *gorm.DB) *gorm.DB {
	return db.Order("dormtrak_reviews.created_at desc")
}

func (*DormtrakReviewModel) scopeCommented(db *gorm.DB) *gorm.DB {
	return db.Where("dormtrak_reviews.comment IS NOT NULL AND dormtrak_reviews.comment <> ''")
}
