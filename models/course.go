package models

import "github.com/jinzhu/gorm"

// Course Model
type CourseModel struct {
	*BaseModel
}

func NewCourseModel(db *gorm.DB) *CourseModel {
	return &CourseModel{
		BaseModel: NewBaseModel(db),
	}
}

func (m *CourseModel) GetAllCourses(c *[]Course) (err error) {
	err = m.DB.Find(c).Error
	return
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

func (m *CourseModel) FindByAbbrevAndNumber(areaAbbreviation string, number string, c *Course) (err error) {
	err = m.DB.
		Preload("AreaOfStudy").
		Preload("AreaOfStudy.Department").
		Joins("JOIN areas_of_study ON areas_of_study.id = courses.area_of_study_id").
		Where("areas_of_study.abbrev = ?", areaAbbreviation).
		Where("courses.number = ?", number).
		First(c).Error
	return
}

// When preloading, must adhere to preloading rules defined in FactrakSurveyModel.preloadDefault()
func (m *CourseModel) GetCourseByIDWithProfessor(id uint, c *Course, profID *uint) (err error) {
	fsM := &FactrakSurveyModel{}

	// Make sure we only return with surveys from profs at Williams
	preloadScopes := []interface{}{
		fsM.preloadDefault,
		fsM.scopeProfAtWilliams,
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
		m.DB.Table("factrak_surveys").Select("course_id").Where(
			"professor_id = ?", profID,
		).QueryExpr(),
	).Find(courses).Error
	return
}

func (m *CourseModel) GetCoursesByDepartment(deptID uint, courses *[]Course) (err error) {
	err = m.DB.Where(
		"area_of_study_id in (?)",
		m.DB.Table("areas_of_study").Select("id").Where(
			"department_id = ?", deptID,
		).QueryExpr(),
	).Find(courses).Error
	return
}

func (m *CourseModel) GetCoursesByAreaOfStudy(areaID uint, courses *[]Course) (err error) {
	err = m.DB.Where("area_of_study_id = ?", areaID).Find(courses).Error
	return
}
