package models

// Course Schema
type Course struct {
	BaseSchema
	Number string `gorm:"not null" json:"number"`

	// Belongs to area of study
	AreaOfStudyID *uint `gorm:"not null" json:"areaOfStudyID"`
	AreaOfStudy *AreaOfStudy `json:"areaOfStudy,omitempty"`

	// Has many factrak surveys
	FactrakSurveys []*FactrakSurvey `json:"factrakSurveys"`
}

func (*Course) TableName() string {
	return "courses"
}
