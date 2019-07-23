package models

// AreaOfStudy Schema
type AreaOfStudy struct {
	BaseSchema
	Name string `gorm:"size:4;unique;not null" json:"name"`
	Abbreviation string `gorm:"column:abbrev;unique;not null" json:"abbreviation"`

	// Belongs to department
	DepartmentID *uint `gorm:"not null" json:"departmentID"`
	Department *Department `json:"department,omitempty"`

	// Has many courses
	Courses []*Course `json:"courses,omitempty"`
}

func (*AreaOfStudy) TableName() string {
	return "areas_of_study"
}
