package models

// CourseCanonical is the PeopleSoft CRSE_ID identity for a course (cross-listings share one row).
type CourseCanonical struct {
	BaseSchema
	CrseID      string `gorm:"unique;not null;size:32" json:"crseID"`
	Title       string `gorm:"not null" json:"title"`
	Description string `gorm:"size:65535" json:"description"`

	Listings  []*CourseListing `json:"listings,omitempty"`
	Offerings []*Offering      `json:"offerings,omitempty"`
}

func (*CourseCanonical) TableName() string {
	return "courses_canonical"
}

// CourseListing is a subject+number(+letter) code that maps to a canonical course for a span of years.
type CourseListing struct {
	BaseSchema
	CourseCanonicalID uint             `gorm:"index:idx_course_listings_course;unique_index:idx_course_listings_identity;not null" json:"courseCanonicalID"`
	CourseCanonical   *CourseCanonical `json:"courseCanonical,omitempty"`

	Subject string `gorm:"unique_index:idx_course_listings_identity;not null;size:16" json:"subject"`
	Number  int    `gorm:"unique_index:idx_course_listings_identity;not null" json:"number"`
	Letter  string `gorm:"unique_index:idx_course_listings_identity;not null;size:8" json:"letter"`

	FirstValidYear int `gorm:"not null" json:"firstValidYear"`
	LastValidYear  int `gorm:"not null" json:"lastValidYear"`
}

func (*CourseListing) TableName() string {
	return "course_listings"
}

const (
	OfferingStatusOffered    = "offered"
	OfferingStatusNotOffered = "not_offered"
	OfferingStatusCancelled  = "cancelled"
)

// Offering is a scheduled section for a canonical course in a given STRM.
type Offering struct {
	BaseSchema
	CourseCanonicalID uint             `gorm:"index:idx_offerings_course;not null" json:"courseCanonicalID"`
	CourseCanonical   *CourseCanonical `json:"courseCanonical,omitempty"`

	Strm     int    `gorm:"unique_index:idx_offerings_strm_class;not null" json:"strm"`
	Year     int    `gorm:"not null" json:"year"`
	Term     string `gorm:"not null;size:16" json:"term"`
	Section  string `gorm:"not null;size:16" json:"section"`
	ClassNbr int    `gorm:"unique_index:idx_offerings_strm_class;not null" json:"classNbr"`
	Title    string `gorm:"not null" json:"title"`
	Component string `gorm:"size:32" json:"component"`
	Status   string `gorm:"not null;size:32" json:"status"`

	Meetings    []*OfferingMeeting    `json:"meetings,omitempty"`
	Instructors []*OfferingInstructor `json:"instructors,omitempty"`
}

func (*Offering) TableName() string {
	return "offerings"
}

// OfferingMeeting is a weekly meeting pattern for an offering.
type OfferingMeeting struct {
	BaseSchema
	OfferingID uint      `gorm:"index:idx_offering_meetings_offering;not null" json:"offeringID"`
	Offering   *Offering `json:"offering,omitempty"`

	Days     string `gorm:"size:16" json:"days"`
	Start    string `gorm:"size:8" json:"start"`
	End      string `gorm:"size:8" json:"end"`
	Facility string `gorm:"size:255" json:"facility"`
}

func (*OfferingMeeting) TableName() string {
	return "offering_meetings"
}

// OfferingInstructor links an offering to a catalog instructor (by unix ID and/or display name).
type OfferingInstructor struct {
	BaseSchema
	OfferingID uint      `gorm:"index:idx_offering_instructors_offering;not null" json:"offeringID"`
	Offering   *Offering `json:"offering,omitempty"`

	UnixID       string `gorm:"index:idx_offering_instructors_unix;size:100" json:"unixID"`
	UserID       *uint  `gorm:"index:idx_offering_instructors_user" json:"userID"`
	User         *User  `json:"user,omitempty"`
	NameAsListed string `gorm:"not null" json:"nameAsListed"`
}

func (*OfferingInstructor) TableName() string {
	return "offering_instructors"
}
