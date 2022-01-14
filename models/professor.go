package models

import (
	"fmt"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// Professor Model
type ProfessorModel struct {
	*UserModel
}

func NewProfessorModel(db *gorm.DB, log *zap.SugaredLogger) *ProfessorModel {
	return &ProfessorModel{
		UserModel: NewUserModel(db, log),
	}
}

func (m *ProfessorModel) GetAllProfessors(u *[]*User, opts Options) (err error) {
	db := m.DB.Scopes(m.scopeDefault)
	if opts != nil {
		db = opts.Run(db)
	}
	err = db.Find(u).Error
	return
}

// Get all professors, ranked by one of the factrak surveys' fields
func (m *ProfessorModel) GetProfessorsRanked(sort string, u *[]*User, opts Options) (err error) {
	if !isProfessorMetric(sort) {
		return lib.ErrorInvalidRankingMetric
	}

	return m.GetAllProfessors(u, opts)
}

type GetAllProfessorsOptions struct {
	// Offset is ignored unless limit is supplied
	Offset *uint `json:"offset" form:"offset"`
	Limit  *uint `json:"limit" form:"limit"`

	// You can preload: department, office, and surveys
	Preload []string `json:"preload" form:"preload[]"`

	// Filtering
	CourseID      *uint `json:"courseID" form:"courseID"`
	DepartmentID  *uint `json:"departmentID" form:"departmentID"`
	AreaOfStudyID *uint `json:"areaOfStudyID" form:"areaOfStudyID"`

	// Rank by a professor metric, in either sort direction
	Metric    *string `json:"metric" form:"metric"`
	Ascending *bool   `json:"ascending" form:"ascending"`
}

// Preload specifically allowed parts if requested
func (o *GetAllProfessorsOptions) Preloader(db *gorm.DB) *gorm.DB {
	if o.Preload == nil {
		return db
	}

	if lib.StringsContains(o.Preload, "department") {
		db = db.Preload("Department")
	}
	if lib.StringsContains(o.Preload, "office") {
		db = db.Preload("Office")
	}
	if lib.StringsContains(o.Preload, "surveys") {
		db = db.Preload("ProfessorFactrakSurveys")
	}

	return db
}

func (o *GetAllProfessorsOptions) Order(db *gorm.DB) *gorm.DB {
	return db.Order("users.id ASC", true)
}

func (o *GetAllProfessorsOptions) Paginate(db *gorm.DB) *gorm.DB {
	db = o.Order(db)
	if o.Limit != nil {
		db = db.Limit(*o.Limit)
		if o.Offset != nil {
			db = db.Offset(*o.Offset)
		}
	}

	return db
}

func (o *GetAllProfessorsOptions) Run(db *gorm.DB) *gorm.DB {
	db = o.Preloader(db)
	db = o.Paginate(db)

	m := NewProfessorModel(db.New(), nil)

	if o.CourseID != nil {
		db = m.withCourse(*o.CourseID)(db)
	}
	if o.DepartmentID != nil {
		db = m.withDepartment(*o.DepartmentID)(db)
	}
	if o.AreaOfStudyID != nil {
		db = m.withAreaOfStudy(*o.AreaOfStudyID)(db)
	}
	if o.Metric != nil {
		if o.Ascending != nil {
			db = m.withRanking(*o.Metric, *o.Ascending)(db)
		} else {
			db = m.withRanking(*o.Metric, false)(db)
		}
	}

	return db
}

func (m *ProfessorModel) DoesProfessorExist(id uint) (exists bool, err error) {
	var count int
	err = m.DB.Model(&User{}).Scopes(m.scopeDefault).Where("users.id = ?", id).Count(&count).Error
	exists = count > 0
	return
}

func (m *ProfessorModel) GetProfessorByID(id uint, u *User) (err error) {
	return m.GetProfessorByIDWithCourse(id, u, nil)
}

// When preloading, must adhere to preloading rules defined in FactrakSurveyModel.scopePreloadDefault()
func (m *ProfessorModel) GetProfessorByIDWithCourse(id uint, u *User, courseID *uint) (err error) {
	fsM := &FactrakSurveyModel{}

	preloadScopes := []interface{}{
		fsM.preloadDefault,
		fsM.preloadCourse,
	}
	if courseID != nil {
		preloadScopes = append(preloadScopes, fsM.withCourseID(*courseID))
	}

	err = m.DB.Scopes(m.scopeDefault).Preload(
		"ProfessorFactrakSurveys", preloadScopes...).Where(NewUserWithID(id)).First(u).Error
	return
}

func (m *ProfessorModel) GetProfessorsByCourse(courseID uint, professors *[]*User) (err error) {
	return m.GetAllProfessors(professors, &GetAllProfessorsOptions{
		CourseID: &courseID,
	})
}

func (m *ProfessorModel) GetProfessorsByDepartment(deptID uint, professors *[]*User) (err error) {
	return m.GetAllProfessors(professors, &GetAllProfessorsOptions{
		DepartmentID: &deptID,
	})
}

// Gets area of study's professors by looking at its courses. This query is an absolute unit so try not to use it.
// Please note that this will only get professors in the factrak system, rather than all professors belonging to this
// area of study.
func (m *ProfessorModel) GetProfessorsByAreaOfStudyViaCourses(areaID uint, professors *[]User) (err error) {
	err = m.DB.Scopes(m.scopeDefault).Where("users.id in (?)",
		m.DB.Model(&FactrakSurvey{}).
			Select("factrak_surveys.professor_id").
			Where("factrak_surveys.course_id in (?)",
				m.DB.Model(&Course{}).
					Select("courses.id").
					Where("courses.area_of_study_id = ?", areaID).QueryExpr(),
			).QueryExpr(),
	).Find(professors).Error
	return
}

// Gets area of study's professors by looking at its department(s). For now, this is the default. However, please notice
// that this will get all department professors, rather than just ones belonging to the area of study.
func (m *ProfessorModel) GetProfessorsByAreaOfStudy(areaID uint, professors *[]*User) (err error) {
	return m.GetAllProfessors(professors, &GetAllProfessorsOptions{
		AreaOfStudyID: &areaID,
	})
}

// Default scope: at williams and is professor
func (m *ProfessorModel) scopeDefault(db *gorm.DB) *gorm.DB {
	return m.scopeAtWilliams(m.scopeIsProfessor(db))
}

func (m *ProfessorModel) scopeIsProfessor(db *gorm.DB) *gorm.DB {
	return db.Where("users.type = ?", UserTypeProfessor)
}

func (m *ProfessorModel) withCourse(courseID uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(
			"users.id in (?)",
			m.DB.Model(&FactrakSurvey{}).Select("professor_id").Where(
				"course_id = ?", courseID,
			).QueryExpr(),
		)
	}
}

func (m *ProfessorModel) withDepartment(deptID uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(&User{DepartmentID: &deptID})
	}
}

func (m *ProfessorModel) withAreaOfStudy(areaID uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(
			"users.department_id in (?)",
			m.DB.Model(&AreaOfStudy{}).Select("department_id").Where(
				"id = ?", areaID,
			).QueryExpr(),
		)
	}
}

func isProfessorMetric(metric string) bool {
	switch metric {
	case
		"course_workload",
		"course_stimulating",
		"would_take_another",
		"approachability",
		"lead_lecture",
		"promote_discussion",
		"outside_helpfulness":
		return true
	}
	return false
}

func (m *ProfessorModel) withRanking(ranking string, ascending bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		var order string

		if ascending {
			order = "ASC"
		} else {
			order = "DESC"
		}

		o := fmt.Sprintf("avg(factrak_surveys.%s) %s", ranking, order)
		h := fmt.Sprintf("count(factrak_surveys.%s) >= 5", ranking)
		q := fmt.Sprintf("factrak_surveys.%s IS NOT NULL", ranking)

		db = db.Joins("left join factrak_surveys on users.id = factrak_surveys.professor_id")
		db = db.Where(q)
		db = db.Where("factrak_surveys.created_at >= ?", time.Now().AddDate(-5, 0, 0))
		db = db.Group("factrak_surveys.professor_id")
		db = db.Having(h)
		db = db.Order(o, true)
		return db
	}
}
