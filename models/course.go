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

// When preloading, must adhere to preloading rules defined in FactrakSurveyModel.preloadDefault()
func (m *CourseModel) GetCourseByID(id uint, c *Course) (err error) {
	return m.GetCourseByIDWithProfessor(id, c, nil)
}

// When preloading, must adhere to preloading rules defined in FactrakSurveyModel.preloadDefault()
func (m *CourseModel) GetCourseByIDWithProfessor(id uint, c *Course, profID *uint) (err error) {
	fsM := &FactrakSurveyModel{}

	preloadScopes := []interface{}{
		fsM.preloadDefault,
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
