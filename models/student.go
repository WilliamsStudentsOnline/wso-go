package models

import (
	"time"

	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

const StudentCutoffMonth = time.July // The month after which rising X are now X

const (
	StudentYearPrefrosh  = 0
	StudentYearFrosh     = 1
	StudentYearSophomore = 2
	StudentYearJunior    = 3
	StudentYearSenior    = 4
)

// Student Model
type StudentModel struct {
	*UserModel
	// We do this so when testing we can plug in our own clock.
	Clock Clock
}

func NewStudentModel(db *gorm.DB, log *zap.SugaredLogger) *StudentModel {
	return &StudentModel{
		UserModel: NewUserModel(db, log),
		Clock:     localClock{},
	}
}

// Given a user, update its survey deficit. Run this every semester and put however many needed.
// We require 2 surveys this semester in order to access it.
func (m *StudentModel) UpdateFactrakSurveyDeficit(user *User) (err error) {
	// Count written surveys
	fsM := NewFactrakSurveyModel(m.DB, m.log)
	surveyCount, err := fsM.CountSurveysByUser(user.ID)
	if err != nil {
		return
	}

	var deficit int

	if surveyCount >= user.Student().surveyThreshold() {
		deficit = 0
	} else {
		surveysThisSem, err := fsM.CountSurveysThisSemesterByUser(user.ID, m.Clock.Now())
		if err != nil {
			return err
		}

		deficit = 2 - surveysThisSem
	}

	// Minimum 0 deficit
	if deficit < 0 {
		deficit = 0
	}

	err = m.DB.Model(&user).Update("factrak_survey_deficit", deficit).Error
	return
}

func (m *StudentModel) SeniorYear() int {
	// TODO: changing time.Now() below to m.Clock will break Student.YearNumber(), which creates an empty StudentModel{} without specifying Clock (a workaround would be use localClock, but it is best if we can keep it consistent with m.Clock)
	locTime := time.Now().Local()
	if locTime.Month() >= StudentCutoffMonth {
		return locTime.Year() + 1
	}

	return locTime.Year()
}

func (m *StudentModel) GetStudentByID(id uint, u *User) (err error) {
	err = m.DB.Scopes(m.scopeDefault).First(u, id).Error
	return
}

// Default scope: at williams and is student
func (m *StudentModel) scopeDefault(db *gorm.DB) *gorm.DB {
	return m.scopeAtWilliams(m.scopeIsStudent(db))
}

func (m *StudentModel) scopeIsStudent(db *gorm.DB) *gorm.DB {
	return db.Where("users.type = ?", UserTypeStudent)
}

func (m *StudentModel) UpdateAllFactrakSurveyDeficits() (err error) {
	var students []User
	err = m.GetAllUsersByType(&students, UserTypeStudent)
	if err != nil {
		return
	}

	for _, student := range students {
		err = m.UpdateFactrakSurveyDeficit(&student)
		if err != nil {
			return
		}
	}
	return
}

// Increase OnCampusSemesters for students registered this semester
func (m *StudentModel) UpdateOnCampusSemesters() (err error) {
	var students []User

	err = m.GetAtWilliamsUsersByType(&students, UserTypeStudent)
	if err != nil {
		return
	}

	for _, student := range students {
		err = m.DB.Model(&student).Update("on_campus_semesters", *student.OnCampusSemesters+1).Error
		if err != nil {
			return
		}
	}

	return
}

func (m *StudentModel) InitializeOnCampusSemesters(u *User) error {
	return m.DB.Model(u).Update("on_campus_semesters", m.CalculateOnCampusSemesters(u)).Error
}

// CalculateOnCampusSemesters calculates a student's number of semesters on-campus according to
// their class year, created_at time and whether they are off cycle.
func (m *StudentModel) CalculateOnCampusSemesters(u *User) int {
	var OnCampusSemesters int
	s := u.Student()

	yearNumber := s.YearNumber()
	if *s.ClassYear-4 < s.CreatedAt.Year() {
		// possibly transfer student
		yearNumber += *s.ClassYear - 4 - s.CreatedAt.Year()

		// Log if we mark someone as transfer
		m.log.Debugf("[transfer student] %v (%v) '%v: On-Campus Year changed from %v to %v. Student created_at %v which should be %v \n",
			s.Name, s.UnixID, *s.ClassYear, s.YearNumber(), yearNumber, s.CreatedAt.Format("2006-Jan-2"), *s.ClassYear-4)
	} else if *s.ClassYear-4 > s.CreatedAt.Year() {
		// possibly student who took a gap year; do nothing and use current yearNumber
	}

	locTime := m.Clock.Now().Local()
	if locTime.Month() >= StudentCutoffMonth {
		// Fall Semester
		OnCampusSemesters = yearNumber*2 - 1
	} else {
		// Spring Semester
		OnCampusSemesters = yearNumber * 2
	}

	if *s.AtWilliams && s.Junior() {
		// Possibly on a junior year study-away program
		return (yearNumber - 1) * 2
	}

	if *s.OffCycle {
		// For those taking one semester off
		OnCampusSemesters--
	}

	return OnCampusSemesters
}

type Student struct {
	*User
}

func (s *Student) YearNumber() int {
	if s.ClassYear == nil {
		return StudentYearFrosh
	} else {
		return 4 - (*s.ClassYear - (&StudentModel{}).SeniorYear())
	}
}

func (s *Student) Prefrosh() bool {
	return s.YearNumber() == StudentYearPrefrosh
}

func (s *Student) Frosh() bool {
	return s.YearNumber() == StudentYearFrosh
}

func (s *Student) Sophomore() bool {
	return s.YearNumber() == StudentYearSophomore
}

func (s *Student) Junior() bool {
	return s.YearNumber() == StudentYearJunior
}

func (s *Student) Senior() bool {
	return s.YearNumber() == StudentYearSenior
}

func (s *Student) IsUpperClass() bool {
	return s.YearNumber() >= StudentYearSophomore
}

// The Factrak survey requirement count
// To be excluded from the 2 surveys requirement this sem, you must have submitted
// at least 2N reviews, where N is the number of semesters you have stayed on campus.
func (s *Student) surveyThreshold() int {
	if *s.OnCampusSemesters < 1 {
		// Pre-Frosh
		return 0
	} else {
		// Note that OnCampusSemesters signals the current semester, and user only need to write about past semesters
		return (*s.OnCampusSemesters - 1) * 2
	}
}
