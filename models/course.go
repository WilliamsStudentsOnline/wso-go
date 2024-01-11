package models

import (
	"fmt"
	"strings"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// Course Model
type CourseModel struct {
	*BaseModel
}

func NewCourseModel(db *gorm.DB, log *zap.SugaredLogger) *CourseModel {
	return &CourseModel{
		BaseModel: NewBaseModel(db, log),
	}
}

func (m *CourseModel) GetAllCourses(c *[]*Course, opts *GetAllCoursesOptions) (err error) {
	db := m.DB
	if opts != nil {
		db = opts.Run(db)
	}

	// Do db query
	err = db.Find(c).Error
	if err != nil {
		return
	}

	// Run post-query options
	if opts != nil {
		opts.Post(*c)
	}
	return
}

// Return courses, ranked by one of the factrak surveys' fields
func (m *CourseModel) GetCoursesRanked(sort string, c *[]*Course, opts *GetAllCoursesOptions) (err error) {
	// Check if the metric is valid
	if !isCourseMetric(sort) {
		return lib.ErrorInvalidRankingMetric
	}
	return m.GetAllCourses(c, opts)
}

type GetAllCoursesOptions struct {
	// Offset is ignored unless limit is supplied
	Offset *uint `json:"offset" form:"offset"`
	Limit  *uint `json:"limit" form:"limit"`

	// You can preload: areaOfStudy, professors, and surveys
	Preload []string `json:"preload" form:"preload[]"`

	// Filters
	AreaOfStudyID *uint `json:"areaOfStudyID" form:"areaOfStudyID"`
	DepartmentID  *uint `json:"departmentID" form:"departmentID"`
	ProfessorID   *uint `json:"professorID" form:"professorID"`

	//Rank by a course metric, in either sort direction
	Metric    *string `json:"metric" form:"metric"`
	Ascending *bool   `json:"ascending" form:"ascending"`
}

// Preload specifically allowed parts if requested
func (o *GetAllCoursesOptions) Preloader(db *gorm.DB) *gorm.DB {
	if o.Preload == nil {
		return db
	}

	if lib.StringsContains(o.Preload, "areaOfStudy") {
		db = db.Preload("AreaOfStudy")
	}
	if lib.StringsContains(o.Preload, "surveys") {
		fsm := NewFactrakSurveyModel(nil, nil)
		db = db.Preload("FactrakSurveys", fsm.scopeCurrent, fsm.scopeProfAtWilliams)
	}
	// We get profs in factrak surveys here, but in Post() we convert those professors into course professors
	if lib.StringsContains(o.Preload, "professors") {
		// Ensure to preload only recent professors
		fsm := NewFactrakSurveyModel(nil, nil)
		pm := NewProfessorModel(nil, nil)
		db = db.Preload("FactrakSurveys", fsm.scopeCurrent).Preload("FactrakSurveys.Professor", pm.scopeAtWilliams)
	}

	if lib.StringsContains(o.Preload, "professorWithAreaOfStudy") {
		fsm := NewFactrakSurveyModel(nil, nil)
		pm := NewProfessorModel(nil, nil)
		db = db.Preload("FactrakSurveys", fsm.scopeCurrent).Preload("FactrakSurveys.Professor", pm.scopeAtWilliams).Preload("FactrakSurveys.Professor.AreasOfStudy")
	}

	return db
}

func (o *GetAllCoursesOptions) Order(db *gorm.DB) *gorm.DB {
	return db.Order("courses.id ASC", true)
}

func (o *GetAllCoursesOptions) Paginate(db *gorm.DB) *gorm.DB {
	//db = o.Order(db)
	if o.Limit != nil {
		db = db.Limit(*o.Limit)
		if o.Offset != nil {
			db = db.Offset(*o.Offset)
		}
	}

	return db
}

func (o *GetAllCoursesOptions) Run(db *gorm.DB) *gorm.DB {
	db = o.Preloader(db)

	m := NewCourseModel(db.New(), nil)

	if o.Metric != nil {
		if o.Ascending != nil {
			db = m.withRanking(*o.Metric, *o.Ascending)(db)
		} else {
			db = m.withRanking(*o.Metric, false)(db)
		}
	} else {
		// No reason to order twice, only order by ID if no metric is given
		db = o.Order(db)
	}

	if o.AreaOfStudyID != nil {
		db = m.withAreaOfStudy(*o.AreaOfStudyID)(db)
		db = db.Preload("AreasOfStudy")
	}
	if o.DepartmentID != nil {
		db = m.withDepartment(*o.DepartmentID)(db)
	}
	if o.ProfessorID != nil {
		db = m.withProfessor(*o.ProfessorID)(db)
	}
	db = o.Paginate(db)

	return db
}

func (o *GetAllCoursesOptions) Post(courses []*Course) {
	if lib.StringsContains(o.Preload, "professors") || lib.StringsContains(o.Preload, "professorWithAreaOfStudy") {
		// Go through all preloaded survey professors and make a unique list of profs
		for _, course := range courses {
			// Unique set of professors of all surveys in this course
			profsSet := make(map[uint]*User)
			for _, survey := range course.FactrakSurveys {
				if survey.Professor != nil {
					profsSet[survey.Professor.ID] = survey.Professor
				}
			}
			// Now make that set a slice
			profs := make([]*User, len(profsSet))
			i := 0
			for _, prof := range profsSet {
				profs[i] = prof
				i++
			}
			// Now populate the (usually) hidden professors field
			course.Professors = profs

			// Now delete the attached factrak surveys because that is bloated if we didn't preload surveys
			if !lib.StringsContains(o.Preload, "surveys") {
				course.FactrakSurveys = nil
			} else {
				// If we are preloading surveys, delete the attached professors
				for i := range course.FactrakSurveys {
					course.FactrakSurveys[i].Professor = nil
				}
			}
		}
	}
}

// Get course by ID. NOTE: does not preload. To preload (like in factrak/courses), call GetCourseByIDWithProfessor() and
// set profID to nil.
func (m *CourseModel) GetCourseByID(id uint, c *Course) (err error) {
	return m.DB.First(c, id).Error
}

func (m *CourseModel) FindOrCreate(c *Course) (err error) {
	return m.DB.FirstOrCreate(c, Course{
		Number:        c.Number,
		AreaOfStudyID: c.AreaOfStudyID,
	}).Error
}

// Finds the relevant course by the title, like CSCI 136. Automatically capitalizes the area.
func (m *CourseModel) FindByAbbrevAndNumber(areaAbbreviation string, number string, c *Course) (err error) {
	areaAbbreviation = strings.ToUpper(areaAbbreviation)

	err = m.DB.
		Preload("AreaOfStudy").
		Preload("AreaOfStudy.Department").
		Joins("JOIN areas_of_study ON areas_of_study.id = courses.area_of_study_id").
		Where("areas_of_study.abbrev = ?", areaAbbreviation).
		Where("courses.number = ?", number).
		First(c).Error
	return
}

// When preloading, must adhere to preloading rules defined in FactrakSurveyModel.scopePreloadDefault()
func (m *CourseModel) GetCourseByIDWithProfessor(id uint, c *Course, profID *uint) (err error) {
	fsM := &FactrakSurveyModel{}

	// Make sure we only return with surveys from profs at Williams
	preloadScopes := []interface{}{
		fsM.preloadDefault,
		fsM.scopeProfAtWilliams,
		fsM.preloadProfessor,
	}
	if profID != nil {
		preloadScopes = append(preloadScopes, fsM.withProfessorID(*profID))
	}

	err = m.DB.Preload("FactrakSurveys", preloadScopes...).Preload("AreaOfStudy").First(c, id).Error
	return
}

func (m *CourseModel) DoesCourseExist(id uint) (exists bool, err error) {
	var count int
	err = m.DB.Model(&Course{}).Where("courses.id = ?", id).Count(&count).Error
	exists = count > 0
	return
}

func (m *CourseModel) GetCoursesByProfessor(profID uint, courses *[]*Course) (err error) {
	return m.GetAllCourses(courses, &GetAllCoursesOptions{
		ProfessorID: &profID,
	})
}

func (m *CourseModel) GetCoursesByDepartment(deptID uint, courses *[]*Course) (err error) {
	return m.GetAllCourses(courses, &GetAllCoursesOptions{
		DepartmentID: &deptID,
	})
}

func (m *CourseModel) GetCoursesByAreaOfStudy(areaID uint, courses *[]*Course) (err error) {
	return m.GetAllCourses(courses, &GetAllCoursesOptions{
		AreaOfStudyID: &areaID,
	})
}

func (m *CourseModel) GetCoursesByAreaOfStudyAndProfessors(areaID uint, courses *[]*Course) (err error) {
	return m.GetAllCourses(courses, &GetAllCoursesOptions{
		AreaOfStudyID: &areaID,
		Preload:       []string{"professors"},
	})
}

func (m *CourseModel) withAreaOfStudy(areaID uint) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("area_of_study_id = ?", areaID)
	}
}

func (m *CourseModel) withDepartment(departmentID uint) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(
			"area_of_study_id in (?)",
			m.DB.Model(&AreaOfStudy{}).Select("id").Where(
				"department_id = ?", departmentID,
			).QueryExpr(),
		)
	}
}

func (m *CourseModel) withProfessor(professorID uint) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(
			"courses.id in (?)",
			m.DB.Model(&FactrakSurvey{}).Select("course_id").Where(
				"professor_id = ?", professorID,
			).QueryExpr(),
		)
	}
}

func isCourseMetric(metric string) bool {
	switch metric {
	case
		"would_recommend_course",
		"course_workload",
		"course_stimulating":
		return true
	}
	return false
}

func (m *CourseModel) withRanking(ranking string, ascending bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		var order string

		if ascending {
			order = "ASC"
		} else {
			order = "DESC"
		}

		avgRating := fmt.Sprintf("avg(factrak_surveys.%s)", ranking)
		havingCount := fmt.Sprintf("count(factrak_surveys.%s) >= 5", ranking)
		notNull := fmt.Sprintf("factrak_surveys.%s IS NOT NULL", ranking)

		subQuery := db.Table("factrak_surveys").Select("course_id, " + avgRating + " as factrak_score").Group("factrak_surveys.course_id").Where(notNull).Having(havingCount).SubQuery()

		db = db.Table("courses").Select("courses.*, scores.factrak_score").Joins("left join (?) as scores on courses.id = scores.course_id", subQuery).Where("factrak_score IS NOT NULL").Order("factrak_score "+order, true)

		/*db = db.Joins("left join factrak_surveys on courses.id = factrak_surveys.course_id")
		db = db.Where(notNull)
		db = db.Where("factrak_surveys.created_at >= ?", time.Now().AddDate(-5, 0, 0))
		db = db.Group("factrak_surveys.course_id")
		db = db.Having(havingCount)
		db = db.Order(avgRating, true)*/
		return db
	}
}
