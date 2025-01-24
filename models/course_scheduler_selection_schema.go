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

	// User information
	User   *User `gorm:"not null" json:"user"`
	UserID *uint `gorm:"index:index_user_id;not null" json:"userID"` // User object UUID as stored in users table, used for indexing

	// Course information
	Course   *Course `gorm:"not null" json:"course"`
	CourseID *uint   `gorm:"not null" json:"courseID"` // Course object UUID as stored in courses table
	Hidden   bool    `gorm:"not null" json:"hidden"`

	// Professor information
	ProfessorID *uint `gorm:"index:index_professor_id;not null" json:"professorID"`
	Professor   *User `gorm:"foreignKey:ProfessorID" json:"professor,omitempty"`

	PeoplesoftID *uint `gorm:"not null" json:"peoplesoftID"`

	Department *string `gorm:"not null" json:"department"` // String representing the shortened department name eg AAS

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

func ParseSemesterString(semesterString string) SemesterType {
	switch semesterString {
	case "FALL":
		return SemesterFall
	case "WINTER":
		return SemesterWinter
	case "SPRING":
		return SemesterSpring
	default:
		return SemesterUndefined
	}
}

func (*CourseSchedulerSelection) TableName() string {
	return "course_scheduler_selection"
}
