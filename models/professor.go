package models

import (
	"github.com/jinzhu/gorm"
)

// Professor Model
type ProfessorModel struct {
	*UserModel
}

func NewProfessorModel(db *gorm.DB) *ProfessorModel {
	return &ProfessorModel{
		UserModel: NewUserModel(db),
	}
}

func (m *ProfessorModel) GetAllProfessors(u *[]User) (err error) {
	err = m.DB.Scopes(m.scopeDefault).Find(u).Error
	return
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

// When preloading, must adhere to preloading rules defined in FactrakSurveyModel.preloadDefault()
func (m *ProfessorModel) GetProfessorByIDWithCourse(id uint, u *User, courseID *uint) (err error) {
	fsM := &FactrakSurveyModel{}

	preloadScopes := []interface{}{
		fsM.preloadDefault,
	}
	if courseID != nil {
		preloadScopes = append(preloadScopes, fsM.withCourseID(*courseID))
	}

	err = m.DB.Scopes(m.scopeDefault).Preload(
		"ProfessorFactrakSurveys", preloadScopes...).Where(NewUserWithID(id)).First(u).Error
	return
}

func (m *ProfessorModel) GetProfessorsByCourse(courseID uint, professors *[]User) (err error) {
	err = m.DB.Where(
		"users.id in (?)",
		m.DB.Table("factrak_surveys").Select("professor_id").Where(
			"course_id = ?", courseID,
		).QueryExpr(),
	).Find(professors).Error
	return
}

func (m *ProfessorModel) GetProfessorsByDepartment(deptID uint, professors *[]User) (err error) {
	err = m.DB.Scopes(m.scopeDefault).Where(&User{DepartmentID: &deptID}).Find(&professors).Error
	return
}

// Gets area of study's professors by looking at its courses. This query is an absolute unit so try not to use it.
// Please note that this will only get professors in the factrak system, rather than all professors belonging to this
// area of study.
func (m *ProfessorModel) GetProfessorsByAreaOfStudyViaCourses(areaID uint, professors *[]User) (err error) {
	err = m.DB.Scopes(m.scopeDefault).Where("users.id in (?)",
		m.DB.Table("factrak_surveys").
			Select("factrak_surveys.professor_id").
			Where("factrak_surveys.course_id in (?)",
				m.DB.Table("courses").
					Select("courses.id").
					Where("courses.area_of_study_id = ?", areaID).QueryExpr(),
			).QueryExpr(),
	).Find(professors).Error
	return
}

// Gets area of study's professors by looking at its department(s). For now, this is the default. However, please notice
// that this will get all department professors, rather than just ones belonging to the area of study.
func (m *ProfessorModel) GetProfessorsByAreaOfStudy(areaID uint, professors *[]User) (err error) {
	err = m.DB.Scopes(m.scopeDefault).Where(
		"department_id in (?)",
		m.DB.Table("areas_of_study").Select("department_id").Where(
			"id = ?", areaID,
		).QueryExpr(),
	).Find(professors).Error
	return
}

// Default scope: at williams and is professor
func (m *ProfessorModel) scopeDefault(db *gorm.DB) *gorm.DB {
	return m.scopeAtWilliams(m.scopeIsProfessor(db))
}

func (m *ProfessorModel) scopeIsProfessor(db *gorm.DB) *gorm.DB {
	return db.Where("users.type = ?", UserTypeProfessor)
}
