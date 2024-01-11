package models

type SemesterType string
const (
	SemesterUndefined	SemesterType = ""
	SemesterSpring		SemesterType = "SPRING"
	SemesterFall 		SemesterType = "FALL"
	SemesterWinter 		SemesterType = "WINTER"
)

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
	return errors.New("Invalid semester type") 
} 

type CourseSchedulerSelection struct {
	BaseSchema
	Student   *User   `gorm:"not null" json:"student"`
	Course    *Course `gorm:"not null" json:"course"`
	StudentID *uint   `gorm:"not null" json:"studentID"`
	CourseID  *uint   `gorm:"not null" json:"courseID"`
	Hidden    bool    `gorm:"not null" json:"hidden"`

	Semester   SemesterType `gorm:"not null" json:"semester" enums:",SPRING,FALL,WINTER"` 
	Year       *uint  `gorm:"not null" json:"year"`
}

func (*Enrollment) TableName() string {
	return "course_scheduler_selection"
}