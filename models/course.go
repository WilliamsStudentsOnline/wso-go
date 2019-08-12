package models

import (
	"strings"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/jinzhu/gorm"
)

// Course Model
type CourseModel struct {
	*BaseModel
}

func NewCourseModel(db *gorm.DB) *CourseModel {
	return &CourseModel{
		BaseModel: NewBaseModel(db),
	}
}

func (m *CourseModel) GetAllCourses(c *[]*Course, opts Options) (err error) {
	db := m.DB
	if opts != nil {
		db = opts.Paginate(db)
		db = opts.Preloader(db)
	}
	err = db.Find(c).Error
	return
}

type GetAllCoursesOptions struct {
	Offset *uint `json:"offset" form:"offset"`
	Limit  *uint `json:"limit" form:"limit"`

	// You can preload: areaOfStudy, professors, and surveys
	Preload *[]string `json:"preload" form:"preload"`
}

// Preload specifically allowed parts if requested
func (p *GetAllCoursesOptions) Preloader(db *gorm.DB) *gorm.DB {
	if p.Preload != nil {
		if lib.StringsContains(*p.Preload, "areaOfStudy") {
			db = db.Preload("AreaOfStudy")
		}
		if lib.StringsContains(*p.Preload, "professors") {
			db = db.Preload("Professors")
		}
		if lib.StringsContains(*p.Preload, "surveys") {
			db = db.Preload("FactrakSurveys")
		}
	}

	return db
}

func (p *GetAllCoursesOptions) Order(db *gorm.DB) *gorm.DB {
	return db.Order("courses.id ASC")
}

func (p *GetAllCoursesOptions) Paginate(db *gorm.DB) *gorm.DB {
	db = p.Order(db)
	if p.Offset != nil {
		db = db.Offset(*p.Offset)
	}
	if p.Limit != nil {
		db = db.Limit(*p.Limit)
	}

	return db
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

func (m *CourseModel) GetCoursesByProfessor(profID uint, courses *[]Course) (err error) {
	err = m.DB.Where(
		"id in (?)",
		m.DB.Model(&FactrakSurvey{}).Select("course_id").Where(
			"professor_id = ?", profID,
		).QueryExpr(),
	).Find(courses).Error
	return
}

func (m *CourseModel) GetCoursesByDepartment(deptID uint, courses *[]Course) (err error) {
	err = m.DB.Where(
		"area_of_study_id in (?)",
		m.DB.Model(&AreaOfStudy{}).Select("id").Where(
			"department_id = ?", deptID,
		).QueryExpr(),
	).Find(courses).Error
	return
}

func (m *CourseModel) GetCoursesByAreaOfStudy(areaID uint, courses *[]Course) (err error) {
	err = m.DB.Where("area_of_study_id = ?", areaID).Find(courses).Error
	return
}

func (m *CourseModel) GetCoursesByAreaOfStudyAndProfessors(areaID uint, courses *[]*Course) (err error) {
	err = m.DB.Where("area_of_study_id = ?", areaID).Preload("FactrakSurveys.Professor").Find(courses).Error

	// Go through all preloaded survey professors and make a unique list of profs
	for _, course := range *courses {
		// Unique set of professors of all surveys in this course
		profsSet := make(map[uint]*User)
		for _, survey := range course.FactrakSurveys {
			profsSet[survey.Professor.ID] = survey.Professor
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

		// Now delete the attached factrak surveys because that is bloated
		course.FactrakSurveys = nil
	}
	return
}
