package models

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

// FactrakSurvey Model
type FactrakSurveyModel struct {
	*BaseModel
}

func NewFactrakSurveyModel(db *gorm.DB) *FactrakSurveyModel {
	return &FactrakSurveyModel{
		BaseModel: NewBaseModel(db),
	}
}

// Gets all surveys.
func (m *FactrakSurveyModel) GetAllSurveys(p *[]*FactrakSurvey, opts *GetAllFactrakSurveysOptions) (err error) {
	err = m.GetAllSurveysWithOptions(p, &GetAllFactrakSurveysOptions{
		ProfessorID:    opts.ProfessorID,
		CourseID:       opts.CourseID,
		UserID:         opts.UserID,
		Offset:         opts.Offset,
		Limit:          opts.Limit,
		Preload:        opts.Preload,
		ProfAtWilliams: true,
	})
	return
}

type GetAllFactrakSurveysOptions struct {
	// Scopes for specific surveys
	ProfessorID *uint `json:"professorID" form:"professorID"`
	CourseID    *uint `json:"courseID" form:"courseID"`
	UserID      *uint `json:"userID" form:"userID"`

	// Pagination
	Offset *time.Time `json:"offset" form:"offset"`
	Limit  *uint      `json:"limit" form:"limit"`

	// Preloading
	Preload []string `json:"preload" form:"preload[]"`

	// Scope to only get flagged surveys. True means only get flagged; false/empty means ignore this scope.
	Flagged bool `json:"-" form:"-"`

	// Scope to only get surveys where the professor is at williams. True means only get profs at Williams;
	// false/empty means ignore this scope.
	ProfAtWilliams bool `json:"-" form:"-"`
}

func (o *GetAllFactrakSurveysOptions) Order(db *gorm.DB) *gorm.DB {
	return db.Order("factrak_surveys.created_at desc", true)
}

// Pagination starts at most recent and goes down from there
func (o *GetAllFactrakSurveysOptions) Paginate(db *gorm.DB) *gorm.DB {
	db = o.Order(db)
	if o.Offset != nil {
		db = db.Where("factrak_surveys.created_at < ?", *o.Offset)
	}
	if o.Limit != nil {
		db = db.Limit(*o.Limit)
	}
	return db
}

// Preload specifically allowed parts if requested
func (o *GetAllFactrakSurveysOptions) Preloader(db *gorm.DB) *gorm.DB {
	if o.Preload == nil {
		return db
	}

	fsM := NewFactrakSurveyModel(nil)

	if stringsContains(o.Preload, "professor") {
		db = fsM.preloadProfessor(db)
	}
	if stringsContains(o.Preload, "course") {
		db = fsM.preloadCourse(db)
	}

	return db
}

func (o *GetAllFactrakSurveysOptions) Run(db *gorm.DB) *gorm.DB {
	db = o.Paginate(db)
	db = o.Preloader(db)

	m := NewFactrakSurveyModel(nil)

	if o.ProfessorID != nil {
		db = m.withProfessorID(*o.ProfessorID)(db)
	}
	if o.CourseID != nil {
		db = m.withCourseID(*o.CourseID)(db)
	}
	if o.UserID != nil {
		db = m.withAuthorID(*o.UserID)(db)
	}
	if o.Flagged {
		db = m.scopeFlagged(db)
	}
	if o.ProfAtWilliams {
		db = m.scopeProfAtWilliams(db)
	}

	return db
}

// Gets all surveys with options
func (m *FactrakSurveyModel) GetAllSurveysWithOptions(p *[]*FactrakSurvey, opts Options) (err error) {
	db := m.DB
	db = m.scopeDefault(db)

	if opts != nil {
		db = opts.Run(db)
	}

	err = db.Find(p).Error
	return
}

// Gets all flagged surveys.
func (m *FactrakSurveyModel) GetAllFlaggedSurveys(p *[]*FactrakSurvey, opts *GetAllFactrakSurveysOptions) (err error) {
	err = m.GetAllSurveysWithOptions(p, &GetAllFactrakSurveysOptions{
		Offset:         opts.Offset,
		Limit:          opts.Limit,
		Preload:        opts.Preload,
		ProfAtWilliams: true,
		Flagged:        true,
	})
	return
}

// Gets survey by its id.
func (m *FactrakSurveyModel) GetSurveyByID(id uint, p *FactrakSurvey) (err error) {
	err = m.DB.Scopes(m.scopeProfAtWilliams, m.scopePreloadDefault).First(p, id).Error
	return
}

func (m *FactrakSurveyModel) DoesSurveyExist(id uint) (exists bool, err error) {
	var count int
	err = m.DB.Model(&FactrakSurvey{}).Scopes(m.scopeProfAtWilliams, m.scopeDefault).Where("factrak_surveys.id = ?", id).Count(&count).Error
	exists = count > 0
	return
}

func (m *FactrakSurveyModel) CountSurveysByUser(userID uint) (count int, err error) {
	err = m.DB.Model(&FactrakSurvey{}).Where("factrak_surveys.user_id = ?", userID).Count(&count).Error
	return
}

// Check if survey already exists by seeing if there is already a survey with that user, course, and professor id.
func (m *FactrakSurveyModel) CheckDuplicateSurvey(userID uint, profID uint, courseID uint) (duplicate bool, err error) {
	var count int
	err = m.DB.Model(&FactrakSurvey{}).Where(&FactrakSurvey{
		UserID:      userID,
		ProfessorID: profID,
		CourseID:    courseID,
	}).Count(&count).Error
	duplicate = count > 0
	return
}

// Create a new survey.
func (m *FactrakSurveyModel) CreateSurvey(p *FactrakSurvey) (err error) {
	err = m.DB.Create(p).Error
	if err != nil {
		return err
	}

	err = m.DB.
		Preload("Professor").
		Preload("Course").
		Preload("Course.AreaOfStudy").
		Preload("Course.AreaOfStudy.Department").
		First(p).Error
	return
}

func (m *FactrakSurveyModel) UpdateSurvey(p *FactrakSurvey) (err error) {
	err = m.DB.Save(&p).Error
	if err != nil {
		return
	}

	err = m.DB.First(p, p.ID).Error
	return
}

func (m *FactrakSurveyModel) DeleteSurvey(p *FactrakSurvey) (err error) {
	// Delete agreements
	err = m.DB.Unscoped().Where(FactrakAgreement{FactrakSurveyID: p.ID}).Delete(&FactrakAgreement{}).Error
	if err != nil {
		return
	}

	err = m.DB.Delete(p).Error
	return
}

func (m *FactrakSurveyModel) SetSurveyFlag(id uint, flag bool) (err error) {
	err = m.DB.Model(NewFactrakSurvey(id)).Update("flagged", flag).Error
	return
}

func (m *FactrakSurveyModel) GetSurveysByProfessor(profID uint, fs *[]*FactrakSurvey, opts *GetAllFactrakSurveysOptions) (err error) {
	err = m.GetAllSurveysWithOptions(fs, &GetAllFactrakSurveysOptions{
		ProfessorID:    &profID,
		CourseID:       opts.CourseID,
		Offset:         opts.Offset,
		Limit:          opts.Limit,
		Preload:        opts.Preload,
		ProfAtWilliams: false,
	})
	return
}

func (m *FactrakSurveyModel) GetSurveysByAuthor(authorUserID uint, fs *[]*FactrakSurvey, opts *GetAllFactrakSurveysOptions) (err error) {
	err = m.GetAllSurveysWithOptions(fs, &GetAllFactrakSurveysOptions{
		ProfessorID:    opts.ProfessorID,
		CourseID:       opts.CourseID,
		UserID:         &authorUserID,
		Offset:         opts.Offset,
		Limit:          opts.Limit,
		Preload:        opts.Preload,
		ProfAtWilliams: true,
	})
	return
}

func (m *FactrakSurveyModel) GetSurveysByCourse(courseID uint, fs *[]*FactrakSurvey, opts *GetAllFactrakSurveysOptions) (err error) {
	err = m.GetAllSurveysWithOptions(fs, &GetAllFactrakSurveysOptions{
		ProfessorID:    opts.ProfessorID,
		CourseID:       &courseID,
		Offset:         opts.Offset,
		Limit:          opts.Limit,
		Preload:        opts.Preload,
		ProfAtWilliams: true,
	})
	return
}

var surveyFields = []string{
	"would_recommend_course",
	"course_workload",
	"course_stimulating",
	"would_take_another",
	"approachability",
	"lead_lecture",
	"promote_discussion",
	"outside_helpfulness",
}

// Gets average survey ratings by professor id, course id, or both.
func (m *FactrakSurveyModel) GetSurveyRatingsByProfessorOrCourse(profID *uint, courseID *uint, ratings *FactrakSurveyAvgRatings) (err error) {
	scopes := []func(db *gorm.DB) *gorm.DB{
		m.scopeCurrent,
	}
	if profID == nil && courseID == nil {
		return errors.New("must have at least one of profID and courseID")
	}
	if profID != nil {
		scopes = append(scopes, m.withProfessorID(*profID))
	}
	if courseID != nil {
		scopes = append(scopes, m.withCourseID(*courseID))
	}

	return m.getSurveyRatings(ratings, scopes...)
}

// This will populate the agreement count fields for a slice of surveys. This does an extra 2*(num surveys) SQL
// requests, which could be slow, so disable and make a new endpoint with this data if that is the case.
func (m *FactrakSurveyModel) PopulateAgreementCountsSlice(surveys []*FactrakSurvey) (err error) {
	for _, survey := range surveys {
		err = m.PopulateAgreementCounts(survey)
		if err != nil {
			return err
		}
	}
	return
}

// This will populate the agreement count fields. This does an extra 2 SQL requests, which could be slow,
// so disable and make a new endpoint with this data if that is the case.
func (m *FactrakSurveyModel) PopulateAgreementCounts(survey *FactrakSurvey) (err error) {
	var posAgree int
	m.DB.Model(&FactrakAgreement{}).Where(
		"factrak_agreements.factrak_survey_id = ?",
		survey.ID,
	).Where("factrak_agreements.agrees = ?", true).Count(&posAgree)

	var negAgree int
	m.DB.Model(&FactrakAgreement{}).Where(
		"factrak_agreements.factrak_survey_id = ?",
		survey.ID,
	).Where("factrak_agreements.agrees = ?", false).Count(&negAgree)

	survey.TotalAgree = posAgree
	survey.TotalDisagree = negAgree
	return
}

func (m *FactrakSurveyModel) getSurveyRatings(ratings *FactrakSurveyAvgRatings, scopes ...func(*gorm.DB) *gorm.DB) (err error) {
	queries := make([]string, 2*len(surveyFields))
	for i, field := range surveyFields {
		queries[2*i] = fmt.Sprintf("avg(%s) AS avg_%s", field, field)
		queries[2*i+1] = fmt.Sprintf("count(%s) AS num_%s", field, field)
	}

	q := strings.Join(queries, ", ")

	err = m.DB.Model(&FactrakSurvey{}).Select(q).Scopes(scopes...).Scan(&ratings).Error
	return
}

func (*FactrakSurveyModel) registrationStart() time.Time {
	// if changed, also change scheduled update user stuff
	springReg := time.October
	fallReg := time.March

	// Between October/X and February/X+1, want October/X
	now := time.Now()
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

// Default for scopes factrak surveys to preload other objects.
func (m *FactrakSurveyModel) scopePreloadDefault(db *gorm.DB) *gorm.DB {
	db = m.preloadProfessor(db)
	db = m.preloadCourse(db)
	return m.scopeDefault(db)
}

// Preloads professor
func (m *FactrakSurveyModel) preloadProfessor(db *gorm.DB) *gorm.DB {
	return db.Preload("Professor")
}

// Preloads course
func (m *FactrakSurveyModel) preloadCourse(db *gorm.DB) *gorm.DB {
	return db.Preload("Course").
		Preload("Course.AreaOfStudy")
}

// Default for scopes preloading factrak surveys via another object. YOU MUST USE THESE UNLESS YOU HAVE EXPLICIT REASONS NOT TO.
func (m *FactrakSurveyModel) preloadDefault(db *gorm.DB) *gorm.DB {
	return m.scopeDefault(db)
}

func (m *FactrakSurveyModel) scopeDefault(db *gorm.DB) *gorm.DB {
	db = m.scopeOrderDefault(db)
	return m.scopeCurrent(db)
}

func (*FactrakSurveyModel) scopeOrderDefault(db *gorm.DB) *gorm.DB {
	return db.Order("factrak_surveys.created_at desc")
}

func (*FactrakSurveyModel) withProfessorID(profID uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(&FactrakSurvey{ProfessorID: profID})
	}
}

func (*FactrakSurveyModel) withAuthorID(authorID uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(&FactrakSurvey{UserID: authorID})
	}
}

func (*FactrakSurveyModel) withCourseID(courseID uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(&FactrakSurvey{CourseID: courseID})
	}
}

func (*FactrakSurveyModel) scopeFlagged(db *gorm.DB) *gorm.DB {
	return db.Where("factrak_surveys.flagged = ?", true)
}

func (*FactrakSurveyModel) scopeProfAtWilliams(db *gorm.DB) *gorm.DB {
	return db.Joins("JOIN users AS professors ON professors.id = factrak_surveys.professor_id").Where("professors.at_williams = ?", true)
}

func (m *FactrakSurveyModel) scopeThisSemester(db *gorm.DB) *gorm.DB {
	return db.Where("factrak_surveys.created_at >= ?", m.registrationStart())
}

func (*FactrakSurveyModel) scopeCurrent(db *gorm.DB) *gorm.DB {
	return db.Where("factrak_surveys.created_at >= ?", time.Now().AddDate(-5, 0, 0))
}

type FactrakSurveyPaginator struct {
	Offset time.Time
	Limit  uint
}

func (p *FactrakSurveyPaginator) Order(db *gorm.DB) *gorm.DB {
	return db.Order("factrak_surveys.created_at desc")
}

// Pagination starts at most recent and goes down from there
func (p *FactrakSurveyPaginator) Paginate(db *gorm.DB) *gorm.DB {
	db = p.Order(db).Limit(p.Limit).Where("factrak_surveys.created_at < ?", p.Offset)
	return db
}

func (m *FactrakSurveyModel) NewSurveyPaginate(offset time.Time, limit int) Paginator {
	l := uint(limit)

	if l == 0 {
		return &NoPaginator{}
	}

	return &FactrakSurveyPaginator{
		Offset: offset,
		Limit:  l,
	}
}

type FactrakSurveyAvgRatings struct {
	AvgWouldRecommendCourse float64 `json:"avgWouldRecommendCourse"`
	NumWouldRecommendCourse int     `json:"numWouldRecommendCourse"`
	AvgCourseWorkload       float64 `json:"avgCourseWorkload"`
	NumCourseWorkload       int     `json:"numCourseWorkload"`
	AvgCourseStimulating    float64 `json:"avgCourseStimulating"`
	NumCourseStimulating    int     `json:"numCourseStimulating"`
	AvgWouldTakeAnother     float64 `json:"avgWouldTakeAnother"`
	NumWouldTakeAnother     int     `json:"numWouldTakeAnother"`
	AvgApproachability      float64 `json:"avgApproachability"`
	NumApproachability      int     `json:"numApproachability"`
	AvgLeadLecture          float64 `json:"avgLeadLecture"`
	NumLeadLecture          int     `json:"numLeadLecture"`
	AvgPromoteDiscussion    float64 `json:"avgPromoteDiscussion"`
	NumPromoteDiscussion    int     `json:"numPromoteDiscussion"`
	AvgOutsideHelpfulness   float64 `json:"avgOutsideHelpfulness"`
	NumOutsideHelpfulness   int     `json:"numOutsideHelpfulness"`
}
