package models

import (
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type EnrollmentModel struct {
	*BaseModel
}

func NewEnrollmentModel(db *gorm.DB, log *zap.SugaredLogger) *EnrollmentModel {
	return &EnrollmentModel{
		BaseModel: NewBaseModel(db, log),
	}
}

func (m *EnrollmentModel) GetAllEnrollments(e *[]*Enrollment, opts *GetAllEnrollmentsOptions) (err error) {
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

type GetAllEnrollmentsOptions struct {
	// Filters
	StudentID    *uint  `json:"studentID" form:"studentID"`
	CourseID     *uint  `json:"courseID" form:"courseID"`
	DepartmentID *uint  `json:"departmentID" form:"departmentID"`
	ProfessorID  *uint  `json:"professorID" form:"professorID"`
	StudentYear  *uint  `json:"studentYear" form:"studentYear"`
	SemesterID   *uint  `json:"semesterID" form:"semesterID"`
	SemesterType string `json:"semesterType" form:"semesterType"`
	Year         *uint  `json:"year" form:"year"`
}

func (o *GetAllEnrollmentsOptions) Run(db *gorm.DB) *gorm.DB {
	m := NewEnrollmentModel(db.New(), nil)

	if o.StudentID != nil {
		db = m.withStudent(*o.StudentID)(db)
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
	if o.StudentYear != nil {
		db = m.withStudentYear(*o.StudentYear)(db)
	}
	if o.SemesterID != nil {
		db = m.withSemesterID(*o.SemesterID)(db)
	}
	if o.SemesterType != "" {
		db = m.withSemesterType(o.SemesterType)(db)
	}
	if o.Year != nil {
		db = m.withYear(*o.Year)(db)
	}

	return db
}

func (m *EnrollmentModel) withStudent(studentID uint) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(
			"student_id = ?", studentID,
		)
	}
}

func (m *EnrollmentModel) withCourse(courseID uint) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(
			"course_id = ?", courseID,
		)
	}
}

// TODO CHECK DOUBLE SUBQUERY
func (m *EnrollmentModel) withDepartment(departmentID uint) func(*gorm.DB) *gorm.DB {
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

func (m *EnrollmentModel) withProfessor(professorID uint) func(*gorm.DB) *gorm.DB {
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

func (m *EnrollmentModel) withStudentYear(studentYear uint) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(
			"student_id in (?)",
			m.DB.Model(&User{}).Select("id").Where(
				"type = ?", "student",
				"class_year = ?", studentYear,
			).QueryExpr(),
		)
	}
}

func (m *EnrollmentModel) withSemesterID(semesterID uint) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(
			"semester_id = ?", semesterID,
		)
	}
}

func (m *EnrollmentModel) withSemesterType(semesterType string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(
			"semester_type = ?", semesterType,
		)
	}
}

func (m *EnrollmentModel) withYear(year uint) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(
			"year = ?", year,
		)
	}
}

func (m *EnrollmentModel) AddCourse(student *User, course *Course, semesterID uint, semesterType string, year uint) (err error) {
	return m.DB.FirstOrCreate(course, Enrollment{
		Student:      student,
		Course:       course,
		StudentID:    &student.ID,
		CourseID:     &course.ID,
		SemesterID:   &semesterID,
		SemesterType: semesterType,
		Year:         &year,
	}).Error
}

func (m *EnrollmentModel) DeleteEntry(course *Enrollment) (err error) {
	return m.DB.Where(
		"student_id = ?", course.StudentID,
		"course_id = ?", course.CourseID,
		"semester_id = ?", course.SemesterID,
		"semester_type = ?", course.SemesterType,
		"year = ?", course.Year,
	).Delete(&Enrollment{}).Error
}

func (m *EnrollmentModel) DeleteEntriesByUserAndCourse(student *User, course *Course) (err error) {
	return m.DB.Where(
		"student_id = ?", &student.ID,
		"course_id = ?", &course.ID,
	).Delete(&Enrollment{}).Error
}

func (m *EnrollmentModel) DeleteEntriesByUser(student *User) (err error) {
	return m.DB.Where(
		"student_id = ?", &student.ID,
	).Delete(&Enrollment{}).Error
}

func (m *EnrollmentModel) DeleteEntriesBySemesterID(semesterID uint) (err error) {
	return m.DB.Where(
		"semester_id = ?", &semesterID,
	).Delete(&Enrollment{}).Error
}

func (m *EnrollmentModel) DeleteEntriesBySemesterAndYear(semesterType string, year uint) (err error) {
	return m.DB.Where(
		"semester_type = ?", semesterType,
		"year = ?", &year,
	).Delete(&Enrollment{}).Error
}

func (m *EnrollmentModel) GetCoursesByStudentID(studentID uint, courses *[]*Enrollment) (err error) {
	return m.GetAllEnrollments(courses, &GetAllEnrollmentsOptions{
		StudentID: &studentID,
	})
}

func (m *EnrollmentModel) GetCoursesByCourseID(courseID uint, courses *[]*Enrollment) (err error) {
	return m.GetAllEnrollments(courses, &GetAllEnrollmentsOptions{
		CourseID: &courseID,
	})
}
