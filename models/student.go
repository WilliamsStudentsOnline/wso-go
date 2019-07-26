package models

import (
	"time"

	"github.com/jinzhu/gorm"
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
}

func NewStudentModel(db *gorm.DB) *StudentModel {
	return &StudentModel{
		UserModel: NewUserModel(db),
	}
}

func (*StudentModel) SeniorYear() int {
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

func (s *Student) Senior() bool {
	return s.YearNumber() == StudentYearSenior
}

// The Factrak survey requirement count
// To be excluded from the 2 surveys requirement this sem, you must have submitted
// at least N - 2 reviews, where N is the number of classes you've taken.
// N is not linear with class year because people might be abroad all jr year.
// it allows 2 non-reviews per semester to account for people taking fewer than 4 courses
// per semester -- we don't want to force them to review more classes than they've had
func (s *Student) surveyTheshold() int {
	// Check semester
	if time.Now().Local().Month() >= StudentCutoffMonth {
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
