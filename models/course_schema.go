package models

// Course Schema
type Course struct {
	BaseSchema
	Number string `gorm:"not null" json:"number"`

	// Belongs to area of study
	AreaOfStudyID *uint        `gorm:"not null" json:"areaOfStudyID"`
	AreaOfStudy   *AreaOfStudy `json:"areaOfStudy,omitempty"`

	// Has many factrak surveys
	FactrakSurveys []*FactrakSurvey `json:"factrakSurveys,omitempty"`

	// Has many professors (thru factrak surveys). Ignore this in Gorm. Populate this only whenever needed.
	Professors []*User `gorm:"-" json:"professors,omitempty"`

	FactrakScore *float64 `gorm:"->;-:migration" json:"factrakScore,omitempty"`
	// Many2Many books
	Books []*Book `gorm:"many2many:course_book;" json:"books,omitempty"`
}

func (*Course) TableName() string {
	return "courses"
}
