package models

// AreaOfStudy Schema
type AreaOfStudy struct {
	BaseSchema
	Name         string `gorm:"unique;not null" json:"name"`
	Abbreviation string `gorm:"size:4;column:abbrev;unique;not null" json:"abbreviation"`

	// Non-empty CatalogURL means the subject is in the current catalog
	CatalogKind      string `gorm:"size:32" json:"catalogKind"` // major, concentration, "major,concentration", or empty
	CatalogURL       string `json:"catalogURL"`
	CurriculumPDFURL string `json:"curriculumPDFURL"`
	WebsiteURL       string `json:"websiteURL"`
	Description      string `gorm:"type:text" json:"description"`

	// Belongs to department
	DepartmentID *uint       `gorm:"not null" json:"departmentID"`
	Department   *Department `json:"department,omitempty"`

	// Has many courses
	Courses []*Course `json:"courses,omitempty"`

	// Many2Many professors  (computed periodically from courses)
	Professors []*User `gorm:"many2many:user_areaOfStudy;" json:"professors"`
}

func (*AreaOfStudy) TableName() string {
	return "areas_of_study"
}
