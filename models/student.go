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

	if surveyCount >= user.Student().surveyThreshold(m.Clock.Now()) {
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
// at least N - 2 reviews, where N is the number of classes you've taken.
// N is not linear with class year because people might be abroad all jr year.
// it allows 2 non-reviews per semester to account for people taking fewer than 4 courses
// per semester -- we don't want to force them to review more classes than they've had
func (s *Student) surveyThreshold(now time.Time) int {
	// Check semester
	if now.Month() >= StudentCutoffMonth {
		// Fall Semester
		switch s.YearNumber() {
		case StudentYearPrefrosh:
			return 0
		case StudentYearFrosh:
			return 0
		case StudentYearSophomore:
			return 6
		case StudentYearJunior:
			return 14
		case StudentYearSenior:
			return 14
		default:
			return 0
		}
	} else {
		// Spring Semester
		switch s.YearNumber() {
		case StudentYearPrefrosh:
			return 0
		case StudentYearFrosh:
			return 2
		case StudentYearSophomore:
			return 10
		case StudentYearJunior:
			return 14
		case StudentYearSenior:
			return 18
		default:
			return 0
		}
	}
}
