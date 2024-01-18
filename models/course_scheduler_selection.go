package models

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type CourseSchedulerSelectionModel struct {
	*BaseModel
}

func NewCourseSchedulerSelectionModel(db *gorm.DB, log *zap.SugaredLogger) *CourseSchedulerSelectionModel {
	return &CourseSchedulerSelectionModel{
		BaseModel: NewBaseModel(db, log),
	}
}

func (m *CourseSchedulerSelectionModel) GetAllCourseSchedulerSelections(e *[]*CourseSchedulerSelection, opts *GetAllCourseSchedulerSelectionsOptions) (err error) {
	db := m.DB
	if opts != nil {
		db = opts.Run(db)
	}

	// Query db
	err = db.Find(e).Error
	if err != nil {
		return
	}

	return
}

type GetAllCourseSchedulerSelectionsOptions struct {
	// You can preload users ("user") and courses ("course")
	Preload []string `json:"preload" form:"preload[]"`

	// Filters
	UserID        *uint        `json:"userID" form:"userID"`
	CourseID      *uint        `json:"courseID" form:"courseID"`
	DepartmentID  *uint        `json:"departmentID" form:"departmentID"`
	ProfessorID   *uint        `json:"professorID" form:"professorID"`
	UserClassYear *uint        `json:"userClassYear" form:"userClassYear"`
	Hidden        bool         `json:"hidden" form:"hidden"`
	Semester      SemesterType `json:"semester" form:"semester"`
	Year          *uint        `json:"year" form:"year"`
}

// Preload users or courses if requested
func (o *GetAllCourseSchedulerSelectionsOptions) Preloader(db *gorm.DB) *gorm.DB {
	if o.Preload == nil {
		return db
	}

	if lib.StringsContains(o.Preload, "user") {
		db = db.Preload("User")
	}

	if lib.StringsContains(o.Preload, "course") {
		db = db.Preload("Course")
	}

	return db
}

func (o *GetAllCourseSchedulerSelectionsOptions) Run(db *gorm.DB) *gorm.DB {
	db = o.Preloader(db)

	m := NewCourseSchedulerSelectionModel(db.New(), nil)

	if o.UserID != nil {
		db = m.withUser(*o.UserID)(db)
	}
	if o.CourseID != nil {
		db = m.withCourse(*o.CourseID)(db)
	}
	if o.DepartmentID != nil {
		db = m.withDepartment(*o.DepartmentID)(db)
	}
	if o.ProfessorID != nil {
		db = m.withProfessor(*o.ProfessorID)(db)
	}
	if o.UserClassYear != nil {
		db = m.withUserClassYear(*o.UserClassYear)(db)
	}
	if o.Semester != "" {
		db = m.withSemester(o.Semester)(db)
	}
	if o.Year != nil {
		db = m.withYear(*o.Year)(db)
	}

	return db
}

func (m *CourseSchedulerSelectionModel) withUser(userID uint) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(
			"user_id = ?", userID,
		)
	}
}

func (m *CourseSchedulerSelectionModel) withCourse(courseID uint) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(
			"course_id = ?", courseID,
		)
	}
}

func (m *CourseSchedulerSelectionModel) withDepartment(departmentID uint) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(
			"course_id in (?)",
			m.DB.Model(&Course{}).Select("id").Where(
				"area_of_study_id in (?)",
				m.DB.Model(&AreaOfStudy{}).Select("id").Where(
					"department_id = ?", departmentID,
				).QueryExpr(),
			).QueryExpr(),
		)
	}
}

func (m *CourseSchedulerSelectionModel) withProfessor(professorID uint) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(
			"course_id in (?)",
			m.DB.Model(&Course{}).Select("id").Where(
				"courses.id in (?)",
				m.DB.Model(&FactrakSurvey{}).Select("course_id").Where(
					"professor_id = ?", professorID,
				).QueryExpr(),
			).QueryExpr(),
		)
	}
}

func (m *CourseSchedulerSelectionModel) withUserClassYear(userClassYear uint) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(
			"user_id in (?)",
			m.DB.Model(&User{}).Select("id").Where(
				"type = ?", "user",
				"class_year = ?", userClassYear,
			).QueryExpr(),
		)
	}
}

func (m *CourseSchedulerSelectionModel) withSemester(semester SemesterType) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(
			"semester = ?", semester,
		)
	}
}

func (m *CourseSchedulerSelectionModel) withYear(year uint) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(
			"year = ?", year,
		)
	}
}

func (m *CourseSchedulerSelectionModel) CreateSelection(selection CourseSchedulerSelection) (err error) {
	return m.DB.Create(&selection).Error
}

func (m *CourseSchedulerSelectionModel) DeleteAllSelectionsByUserID(userID uint) (err error) {
	return m.DB.Where(
		"user_id = ?", userID,
	).Delete(&CourseSchedulerSelection{}).Error
}

func (m *CourseSchedulerSelectionModel) DeleteAllSelectionsByCourseID(courseID uint) (err error) {
	return m.DB.Where(
		"course_id = ?", courseID,
	).Delete(&CourseSchedulerSelection{}).Error
}

func (m *CourseSchedulerSelectionModel) DeleteAllSelectionsBySemesterAndYear(semester SemesterType, year uint) (err error) {
	return m.DB.Where(
		"semester = ?", semester,
		"year = ?", year,
	).Delete(&CourseSchedulerSelection{}).Error
}

func (m *CourseSchedulerSelectionModel) DeleteAllSelectionsByUserIDAndCourseID(userID uint, courseID uint) (err error) {
	return m.DB.Where(
		"user_id = ?", userID,
		"course_id = ?", courseID,
	).Delete(&CourseSchedulerSelection{}).Error
}

func (m *CourseSchedulerSelectionModel) DeleteAllSelectionsByUserIDAndSemesterAndYear(userID uint, semester string, year uint) (err error) {
	return m.DB.Where(
		"user_id = ?", userID,
		"semester = ?", semester,
		"year = ?", year,
	).Delete(&CourseSchedulerSelection{}).Error
}

func (m *CourseSchedulerSelectionModel) GetSelectionsByUserID(userID uint, courseSchedulerSelections *[]*CourseSchedulerSelection) (err error) {
	return m.GetAllCourseSchedulerSelections(courseSchedulerSelections, &GetAllCourseSchedulerSelectionsOptions{
		UserID: &userID,
	})
}

func (m *CourseSchedulerSelectionModel) GetSelectionsByCourseID(courseID uint, courseSchedulerSelections *[]*CourseSchedulerSelection) (err error) {
	return m.GetAllCourseSchedulerSelections(courseSchedulerSelections, &GetAllCourseSchedulerSelectionsOptions{
		CourseID: &courseID,
	})
}

func (m *CourseSchedulerSelectionModel) GetSelectionsByUserIDAndSemesterAndYear(userID uint, semester SemesterType, year uint, courseSchedulerSelections *[]*CourseSchedulerSelection) (err error) {
	return m.GetAllCourseSchedulerSelections(courseSchedulerSelections, &GetAllCourseSchedulerSelectionsOptions{
		UserID:   &userID,
		Semester: semester,
		Year:     &year,
	})
}

func (m *CourseSchedulerSelectionModel) SetSelectionHiddenByUserIDAndCourseID(userID uint, courseID uint, hidden bool) (err error) {
	return m.DB.Model(&CourseSchedulerSelection{}).Where(
		"user_id = ?", userID,
		"course_id = ?", courseID,
	).Update("Hidden", hidden).Error
}

func (m *CourseSchedulerSelectionModel) SetSelectionHiddenByUserIDAndSemesterAndYear(userID uint, semester SemesterType, year uint, hidden bool) (err error) {
	return m.DB.Model(&CourseSchedulerSelection{}).Where(
		"user_id = ?", userID,
		"semester = ?", semester,
		"year = ?", year,
	).Update("Hidden", hidden).Error
}
