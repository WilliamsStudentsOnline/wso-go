package models

import (
	"encoding/json"
	"errors"
)

type SemesterType string

const (
	SemesterUndefined SemesterType = ""
	SemesterSpring    SemesterType = "SPRING"
	SemesterFall      SemesterType = "FALL"
	SemesterWinter    SemesterType = "WINTER"
)

type CourseSchedulerSelection struct {
	BaseSchema

	// Student information
	Student   *User `gorm:"not null" json:"student"`
	StudentID *uint `gorm:"not null" json:"studentID"` // User object UUID as stored in users table

	// Course information
	Course   *Course `gorm:"not null" json:"course"`
	CourseID *uint   `gorm:"not null" json:"courseID"` // Course object UUID as stored in courses table
	Hidden   bool    `gorm:"not null" json:"hidden"`

	Semester SemesterType `gorm:"not null" json:"semester" enums:",SPRING,FALL,WINTER"`
	Year     *uint        `gorm:"not null" json:"year"`
}

func (semesterType *SemesterType) UnmarshalJSON(b []byte) error {
	// Define a secondary type to avoid ending up with a recursive call to json.Unmarshal
	type S SemesterType
	var r = (*S)(semesterType)
	err := json.Unmarshal(b, &r)
	if err != nil {
		panic(err)
	}
	switch *semesterType {
	case SemesterUndefined, SemesterSpring, SemesterFall, SemesterWinter:
		return nil
	}
	return errors.New("invalid semester type")
}

func (*CourseSchedulerSelection) TableName() string {
	return "course_scheduler_selection"
}
