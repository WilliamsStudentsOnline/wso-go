package models

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib"
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

func (m *ProfessorModel) GetAllProfessors(u *[]*User, opts Options) (err error) {
	db := m.DB.Scopes(m.scopeDefault)
	if opts != nil {
		db = opts.Paginate(db)
		db = opts.Preloader(db)
	}
	err = db.Find(u).Error
	return
}

type GetAllProfessorsOptions struct {
	Offset *uint `json:"offset" form:"offset"`
	Limit  *uint `json:"limit" form:"limit"`

	// You can preload: department, office, and surveys
	Preload *[]string `json:"preload" form:"preload"`
}

// Preload specifically allowed parts if requested
func (p *GetAllProfessorsOptions) Preloader(db *gorm.DB) *gorm.DB {
	var scopes []func(*gorm.DB) *gorm.DB

	if p.Preload != nil {
		if lib.StringsContains(*p.Preload, "department") {
			db = db.Preload("Department")
		}
		if lib.StringsContains(*p.Preload, "office") {
			db = db.Preload("Office")
		}
		if lib.StringsContains(*p.Preload, "surveys") {
			db = db.Preload("ProfessorFactrakSurveys")
		}
	}

	return db.Scopes(scopes...)
}

func (p *GetAllProfessorsOptions) Order(db *gorm.DB) *gorm.DB {
	return db.Order("users.id ASC")
}

func (p *GetAllProfessorsOptions) Paginate(db *gorm.DB) *gorm.DB {
	db = p.Order(db)
	if p.Offset != nil {
		db = db.Offset(p.Offset)
	}
	if p.Limit != nil {
		db = db.Limit(p.Limit)
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

func (m *ProfessorModel) GetProfessorsByCourse(courseID uint, professors *[]User) (err error) {
	err = m.DB.Where(
		"users.id in (?)",
		m.DB.Model(&FactrakSurvey{}).Select("professor_id").Where(
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
func (m *ProfessorModel) GetProfessorsByAreaOfStudy(areaID uint, professors *[]User) (err error) {
	err = m.DB.Scopes(m.scopeDefault).Where(
		"department_id in (?)",
		m.DB.Model(&AreaOfStudy{}).Select("department_id").Where(
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
