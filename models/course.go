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

func (m *CourseModel) GetCoursesByProfessor(profID uint, courses *[]Course) (err error) {
	err = m.DB.Where("id in (?)", m.DB.Table("factrak_surveys").Select("course_id").Where("professor_id = ?", profID).QueryExpr()).Find(courses).Error
	return
}