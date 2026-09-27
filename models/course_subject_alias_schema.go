package models

// CourseSubjectAlias maps a Factrak / legacy area-of-study abbrev to a catalog SUBJECT code.
type CourseSubjectAlias struct {
	BaseSchema
	FactrakAbbrev  string `gorm:"unique;not null;size:16" json:"factrakAbbrev"`
	CatalogSubject string `gorm:"not null;size:16" json:"catalogSubject"`
}

func (*CourseSubjectAlias) TableName() string {
	return "course_subject_aliases"
}
