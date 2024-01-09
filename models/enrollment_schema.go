package models

type Enrollment struct {
	BaseSchema
	Student   *User   `gorm:"not null" json:"student"`
	Course    *Course `gorm:"not null" json:"course"`
	StudentID *uint   `gorm:"not null" json:"studentID"`
	CourseID  *uint   `gorm:"not null" json:"courseID"`

	SemesterID   *uint  `gorm:"not null" json:"semesterID"`
	SemesterType string `gorm:"not null" json:"semesterType"`
	Year         *uint  `gorm:"not null" json:"year"`
}

func (*Enrollment) TableName() string {
	return "enrollment"
}
