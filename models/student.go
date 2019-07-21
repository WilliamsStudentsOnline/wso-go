package models

import "time"

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
	UserModel
}

func (*StudentModel) SeniorYear() int {
	locTime := time.Now().Local()
	if locTime.Month() >= StudentCutoffMonth {
		return locTime.Year() + 1
	}

	return locTime.Year()
}

type Student struct {
	*User
}

func (s *Student) YearNumber() int {
	if s.ClassYear == nil {
		return 0
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
