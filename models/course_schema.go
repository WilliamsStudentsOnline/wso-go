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

	//for course retrival for rankings
	FactrakScore *float64 `gorm:"->;-:migration" json:"factrakScore,omitempty"`
}

func (*Course) TableName() string {
	return "courses"
}
