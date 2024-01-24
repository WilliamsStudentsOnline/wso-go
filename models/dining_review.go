package models

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// DiningReview Model
type DiningHallReviewModel struct {
	*BaseModel
}

func NewDiningHallReviewModel(db *gorm.DB, log *zap.SugaredLogger) *DiningHallReviewModel {
	return &DiningHallReviewModel{
		BaseModel: NewBaseModel(db, log),
	}
}

type GetAllDiningHallReviewsOptions struct {
	// Scopes for specific surveys
	UserID   *uint `json:"userID" form:"userID"`
	DiningID *uint `json:"diningID" form:"diningID"`

	// Preloading: review
	Preload []string `json:"preload" form:"preload[]"`

	// Populate survey agreements
	PopulateAllAgreements bool `json:"populateAllAgreements" form:"populateAllAgreements"`
	// Populate client agreed with survey
	PopulateUserAgreement bool `json:"populateUserAgreement" form:"populateUserAgreement"`
	// Pass an ignored user ID for the PopulateDidAgree function to look up
	UserAgreementUserID uint `json:"-" form:"-"`

	// Scope to only get flagged surveys. True means only get flagged; false/empty means ignore this scope.
	Flagged bool `json:"-" form:"-"`
}

func (o *GetAllDiningHallReviewsOptions) Order(db *gorm.DB) *gorm.DB {
	return db.Order("dining_reviews.created_at desc", true)
}

// Preload specifically allowed parts if requested
func (o *GetAllDiningHallReviewsOptions) Preloader(db *gorm.DB) *gorm.DB {
	if o.Preload == nil {
		return db
	}

	dhrM := NewDiningHallReviewModel(nil, nil)

	if stringsContains(o.Preload, "diningHall") {
		db = dhrM.preloadDiningHall(db)
	}

	return db
}

func (o *GetAllDiningHallReviewsOptions) Run(db *gorm.DB) *gorm.DB {
	db = o.Preloader(db)

	dhrM := NewDiningHallReviewModel(nil, nil)

	if o.UserID != nil {
		db = dhrM.withAuthorID(*o.UserID)(db)
	}
	if o.Flagged {
		db = dhrM.scopeFlagged(db)
	}

	return db
}

func (o *GetAllDiningHallReviewsOptions) Post(db *gorm.DB, reviews *[]*DiningHallReview) (err error) {
	if o.PopulateAllAgreements {
		dhrM := NewDiningHallReviewModel(db, nil)
		err = dhrM.PopulateUserAgreementCountsSlice(*reviews)
		if err != nil {
			return err
		}
	}
	if o.PopulateUserAgreement && o.UserAgreementUserID > 0 {
		dhrM := NewDiningHallReviewModel(db, nil)
		err = dhrM.PopulateUserAgreementsSlice(o.UserAgreementUserID, *reviews)
		if err != nil {
			return err
		}
	}
	return
}

// Gets all reviews with options
func (m *DiningHallReviewModel) GetAllReviews(r *[]*DiningHallReview, allOpts *GetAllDiningHallReviewsOptions) (err error) {
	db := m.DB
	db = m.scopeDefault(db)

	if allOpts != nil {
		db = allOpts.Run(db)
	}

	// Do db query
	err = db.Find(r).Error
	if err != nil {
		return
	}

	// Run post-query options
	if allOpts != nil {
		err = allOpts.Post(m.DB, r)
		if err != nil {
			return
		}
	}

	return
}

// Gets review by its id.
func (m *DiningHallReviewModel) GetReviewByID(id uint, r *DiningHallReview) (err error) {
	err = m.DB.Scopes(m.scopePreloadDefault).First(r, id).Error
	return
}

// Checks if review exists
func (m *DiningHallReviewModel) DoesReviewExist(id uint) (exists bool, err error) {
	var count int
	err = m.DB.Model(&FactrakSurvey{}).Scopes(m.scopeDefault).Where("dining_reviews.id = ?", id).Count(&count).Error
	exists = count > 0
	return
}

// Check if duplicate review already exists by seeing if there is already a review with that user and dining id.
func (m *DiningHallReviewModel) CheckDuplicateReview(userID uint) (duplicate bool, err error) {
	var count int
	err = m.DB.Model(&DiningHallReview{}).Where(&DiningHallReview{
		UserID: userID,
	}).Count(&count).Error
	duplicate = count > 0
	return
}

// Create a new review
func (m *DiningHallReviewModel) CreateReview(r *DiningHallReview) (err error) {
	err = m.DB.Create(r).Error
	if err != nil {
		return err
	}

	err = m.DB.
		Preload("Dining Hall").
		First(r).Error
	return
}

// update review
func (m *DiningHallReviewModel) UpdateReview(r *DiningHallReview) (err error) {
	err = m.DB.Save(&r).Error
	if err != nil {
		return
	}

	err = m.DB.First(r, r.ID).Error
	return
}

// delete review
func (m *DiningHallReviewModel) DeleteReview(r *DiningHallReview) (err error) {
	// Delete agreements
	err = m.DB.Unscoped().Where(DiningReviewAgreement{DiningReviewID: r.ID}).Delete(&DiningReviewAgreement{}).Error
	if err != nil {
		return
	}

	err = m.DB.Delete(r).Error
	return
}

func (m *DiningHallReviewModel) SetReviewFlag(id uint, flag bool) (err error) {
	err = m.DB.Model(NewDiningHallReview(id)).Update("flagged", flag).Error
	return
}

var reviewFields = []string{
	"would_recommend_food",
	"food_quality",
	"wait_time",
}

// Gets average wait time by dining hall id
func (m *DiningHallReviewModel) GetReviewRatingsByDiningHall(diningID *uint, metric *string, ratings *DiningHallReviewAvgRatings) (err error) {
	scopes := []func(db *gorm.DB) *gorm.DB{
		m.scopeCurrent,
	}
	if diningID == nil {
		return errors.New("must have diningID")
	}

	if metric != nil && *metric != "" {
		return m.getSingleReviewRating(ratings, metric, scopes...)
	}

	return m.getReviewRatings(ratings, scopes...)
}

// This will populate the agreement count fields for a slice of surveys. This does an extra 2*(num surveys) SQL
// requests, which could be slow, so disable and make a new endpoint with this data if that is the case.
func (m *DiningHallReviewModel) PopulateUserAgreementCountsSlice(reviews []*DiningHallReview) (err error) {
	for _, survey := range reviews {
		err = m.PopulateUserAgreementCounts(survey)
		if err != nil {
			return err
		}
	}
	return
}

// This will populate the agreement count fields. This does an extra 2 SQL requests, which could be slow,
// so disable and make a new endpoint with this data if that is the case.
func (m *DiningHallReviewModel) PopulateUserAgreementCounts(review *DiningHallReview) (err error) {
	var posAgree int
	err = m.DB.Model(&DiningReviewAgreement{}).Where(
		"dining_review_id = ?",
		review.ID,
	).Where("agrees = ?", true).Count(&posAgree).Error
	if err != nil {
		return err
	}

	var negAgree int
	err = m.DB.Model(&DiningReviewAgreement{}).Where(
		"dining_review_id = ?",
		review.ID,
	).Where("agrees = ?", false).Count(&negAgree).Error
	if err != nil {
		return err
	}

	review.TotalAgree = posAgree
	review.TotalDisagree = negAgree
	return
}

// This will populate the didAgree field for a slice of reviews. This does an extra 1*(num surveys) SQL
// requests, which could be slow, so disable and make a new endpoint with this data if that is the case.
func (m *DiningHallReviewModel) PopulateUserAgreementsSlice(userID uint, reviews []*DiningHallReview) (err error) {
	for _, review := range reviews {
		err = m.PopulateUserAgreements(userID, review)
		if err != nil {
			return err
		}
	}
	return
}

// This will populate the didAgree field. This does an extra 1 SQL requests, which could be slow,
// so disable and make a new endpoint with this data if that is the case.
func (m *DiningHallReviewModel) PopulateUserAgreements(userID uint, review *DiningHallReview) (err error) {
	agreement := DiningReviewAgreement{}
	err = m.DB.Where(&DiningReviewAgreement{
		DiningReviewID: review.ID,
		UserID:         userID,
	}).Find(&agreement).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			review.UserAgreement = nil
			return nil
		} else {
			return err
		}
	}

	review.UserAgreement = &agreement.Agrees
	return
}

func (m *DiningHallReviewModel) getReviewRatings(ratings *DiningHallReviewAvgRatings, scopes ...func(*gorm.DB) *gorm.DB) (err error) {
	queries := make([]string, 2*len(reviewFields))
	for i, field := range reviewFields {
		queries[2*i] = fmt.Sprintf("avg(%s) AS avg_%s", field, field)
		queries[2*i+1] = fmt.Sprintf("count(%s) AS num_%s", field, field)
	}

	q := strings.Join(queries, ", ")

	err = m.DB.Model(&DiningHallReview{}).Select(q).Scopes(scopes...).Scan(&ratings).Error
	return
}

func (m *DiningHallReviewModel) getSingleReviewRating(ratings *DiningHallReviewAvgRatings, metric *string, scopes ...func(*gorm.DB) *gorm.DB) (err error) {
	avg := fmt.Sprintf("avg(%s) AS avg_%s", *metric, *metric)
	count := fmt.Sprintf("count(%s) AS num_%s", *metric, *metric)

	q := avg + ", " + count

	err = m.DB.Model(&DiningHallReview{}).Select(q).Scopes(scopes...).Scan(&ratings).Error
	return
}

func (*DiningHallReviewModel) registrationStart(now time.Time) time.Time {
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

func (*DiningHallReviewModel) withAuthorID(authorID uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(&DiningHallReview{UserID: authorID})
	}
}

// Default for scopes dining hall reviews to preload other objects.
func (m *DiningHallReviewModel) scopePreloadDefault(db *gorm.DB) *gorm.DB {
	db = m.preloadDiningHall(db)
	return m.scopeDefault(db)
}

// Preloads dining hall
func (m *DiningHallReviewModel) preloadDiningHall(db *gorm.DB) *gorm.DB {
	return db.Preload("diningHall")
}

func (m *DiningHallReviewModel) scopeDefault(db *gorm.DB) *gorm.DB {
	db = m.scopeOrderDefault(db)
	return m.scopeCurrent(db)
}

func (*DiningHallReviewModel) scopeOrderDefault(db *gorm.DB) *gorm.DB {
	return db.Order("dining_reviews.created_at desc")
}

func (*DiningHallReviewModel) scopeFlagged(db *gorm.DB) *gorm.DB {
	return db.Where("dining_reviews.flagged = ?", true)
}

func (*DiningHallReviewModel) scopeCurrent(db *gorm.DB) *gorm.DB {
	return db.Where("dining_reviews.created_at >= ?", time.Now().AddDate(-5, 0, 0))
}

type DiningHallReviewAvgRatings struct {
	AvgWouldRecommendFood *bool `json:"avgWouldRecommendFood"`
	AvgFoodQuality        *int  `json:"avgFoodQuality"`
	AvgWaitTime           *int  `json:"avgWaitTime"`
}
