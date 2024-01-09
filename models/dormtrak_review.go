package models

import (
	"time"

	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// DormtrakReview Model
type DormtrakReviewModel struct {
	*BaseModel
}

// New Function (from factrak_survey.go) - Brenda
func (m *DormtrakReviewModel) CountReviewsbyUser(userID uint) (count int, err error) {
	err = m.DB.Model(&DormtrakReview{}).Where("dormtrak_reviews.user_id = ?", userID).Count(&count).Error
	return
}

// New Function - Brenda
func (m *DormtrakReviewModel) CountReviewsThisSemesterbyUser(userID uint, now time.Time) (count int, err error) {
	err = m.DB.Model(&DormtrakReview{}).
		Scopes(m.withScopeThisSemester(now), m.withAuthorID(userID)).
		Count(&count).Error
	return
}

func NewDormtrakReviewModel(db *gorm.DB, log *zap.SugaredLogger) *DormtrakReviewModel {
	return &DormtrakReviewModel{
		BaseModel: NewBaseModel(db, log),
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
	Start *time.Time `json:"start" form:"start"`
	// Offset is ignored unless limit is supplied
	Offset *uint `json:"offset" form:"offset"`
	Limit  *uint `json:"limit" form:"limit"`

	// What to preload
	Preload []string `json:"preload" form:"preload[]"`

	// Scope to only get commented reviews. True means only get commented; false/empty means ignore this scope.
	Commented bool `json:"commented" form:"commented"`
}

func (o *GetAllDormtrakReviewsOptions) Order(db *gorm.DB) *gorm.DB {
	return db.Order("dormtrak_reviews.created_at desc", true)
}

// Pagination starts at most recent and goes down from there
func (o *GetAllDormtrakReviewsOptions) Paginate(db *gorm.DB) *gorm.DB {
	db = o.Order(db)
	if o.Start != nil {
		db = db.Where("dormtrak_reviews.created_at < ?", *o.Start)
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
func (o *GetAllDormtrakReviewsOptions) Preloader(db *gorm.DB) *gorm.DB {
	if o.Preload == nil {
		return db
	}

	if stringsContains(o.Preload, "dormRoom") {
		db = db.Preload("DormRoom")
	}
	if stringsContains(o.Preload, "dorm") {
		db = db.Preload("DormRoom.Dorm")
	}
	if stringsContains(o.Preload, "neighborhood") {
		db = db.Preload("DormRoom.Dorm.Neighborhood")
	}

	return db
}

func (o *GetAllDormtrakReviewsOptions) Run(db *gorm.DB) *gorm.DB {
	db = o.Preloader(db)
	db = o.Paginate(db)

	m := NewDormtrakReviewModel(nil, nil)

	if o.DormID != nil {
		db = m.withDormID(*o.DormID)(db)
	}
	if o.DormRoomID != nil {
		db = m.withDormRoomID(*o.DormRoomID)(db)
	}
	if o.UserID != nil {
		db = m.withUserID(*o.UserID)(db)
	}
	if o.Commented {
		db = m.scopeCommented(db)
	}

	return db
}

// Gets all reviews.
func (m *DormtrakReviewModel) GetAllReviewsWithOptions(p *[]*DormtrakReview, opts Options) (err error) {
	db := m.DB
	db = m.scopeDefault(db)
	if opts != nil {
		db = opts.Run(db)
	}

	err = db.Find(p).Error
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

// New Function - Brenda
func (*DormtrakReviewModel) registrationStart(now time.Time) time.Time {
	// if changed, also change scheduled update user stuff
	springReg := time.October
	fallReg := time.March

	// Between October/X and February/X+1, want October/X
	month := now.Month()

	// TODO: Ensure time.Local is EST/EDT on server
	if month <= time.February {
		return time.Date(now.Year()-1, springReg, 1, 1, 0, 0, 0, time.Local)
	} else if month >= time.October {
		return time.Date(now.Year(), springReg, 1, 1, 0, 0, 0, time.Local)
	} else {
		return time.Date(now.Year(), fallReg, 1, 1, 0, 0, 0, time.Local)
	}
}

// New Function - Brenda
func (m *DormtrakReviewModel) withScopeThisSemester(now time.Time) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("dormtrak_reviews.created_at >= ?", m.registrationStart(now))
	}
}

// New Function - Brenda
func (*DormtrakReviewModel) scopeCurrent(db *gorm.DB) *gorm.DB {
	return db.Where("dormtrak_reviews.created_at >= ?", time.Now().AddDate(-5, 0, 0))
}

// New Function - Brenda
func (*DormtrakReviewModel) withAuthorID(authorID uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(&FactrakSurvey{UserID: authorID})
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
