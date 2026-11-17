package models

const (
	CourseAttributeKindDistribution = "distribution" // Div I/II/III, DPE, QFR, WAC
	CourseAttributeKindProgram      = "program"      // department or program groups, e.g. AFR_AFRCORE
)

type CourseAttribute struct {
	BaseSchema
	Code string `gorm:"size:64;unique;not null" json:"code"`
	Name string `gorm:"not null" json:"name"`
	Kind string `gorm:"size:32;not null" json:"kind"`

	// The area whose catalog page lists this attribute, when known
	AreaOfStudyID *uint        `gorm:"index:idx_course_attributes_area" json:"areaOfStudyID"`
	AreaOfStudy   *AreaOfStudy `json:"areaOfStudy,omitempty"`
}

func (*CourseAttribute) TableName() string {
	return "course_attributes"
}
