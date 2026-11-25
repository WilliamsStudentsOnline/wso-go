package models

// One per PeopleSoft CRSE_ID; cross-listings share it
type CourseCanonical struct {
	BaseSchema
	CrseID      string `gorm:"unique;not null;size:32" json:"crseID"`
	Title       string `gorm:"not null" json:"title"`
	Description string `gorm:"size:65535" json:"description"`

	Listings          []*CourseListing    `json:"listings,omitempty"`
	CoursesInSemester []*CourseInSemester `json:"coursesInSemester,omitempty"`
	Offerings         []*CourseOffering   `json:"offerings,omitempty"`
}

func (*CourseCanonical) TableName() string {
	return "course_canonicals"
}

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
	CourseOfferingStatusOffered    = "offered"
	CourseOfferingStatusNotOffered = "not_offered"
	CourseOfferingStatusCancelled  = "cancelled"
)

const (
	CatalogSourceCatalog = "catalog"
	// Recovered from a "not offered" row for terms before the feed starts; title and instructors only
	CatalogSourceLastOffered = "last_offered"
)

type CourseInSemester struct {
	BaseSchema
	CourseCanonicalID uint             `gorm:"index:idx_course_in_semesters_course;unique_index:idx_course_in_semesters_term;not null" json:"courseCanonicalID"`
	CourseCanonical   *CourseCanonical `json:"courseCanonical,omitempty"`

	Strm   int    `gorm:"unique_index:idx_course_in_semesters_term;index:idx_course_in_semesters_strm;not null" json:"strm"`
	Year   int    `gorm:"not null" json:"year"` // catalog year (spring calendar year): Fall 2026 is 2027
	Term   string `gorm:"not null;size:16" json:"term"`
	Status string `gorm:"not null;size:32" json:"status"` // offered if any section is offered, else cancelled
	Source string `gorm:"not null;size:16;default:'catalog'" json:"source"`

	Title            string `gorm:"not null" json:"title"`
	Description      string `gorm:"size:65535" json:"description"`
	ShortDescription string `gorm:"size:65535" json:"shortDescription"`
	Components       string `gorm:"size:64" json:"components"` // e.g. "lecture/laboratory"
	ClassFormat      string `gorm:"size:65535" json:"classFormat"`
	GradingBasis     string `gorm:"size:16" json:"gradingBasis"` // catalog code: GRD, OPT, WPP, ...
	Units            string `gorm:"size:8" json:"units"`
	Campus           string `gorm:"size:16" json:"campus"`

	// Comma-separated catalog codes, e.g. "DIV_D3,QFR_QFR,AFR_AFRCORE"
	Attributes string `gorm:"size:1024" json:"attributes"`
	Div1       bool   `gorm:"not null;default:false" json:"div1"`
	Div2       bool   `gorm:"not null;default:false" json:"div2"`
	Div3       bool   `gorm:"not null;default:false" json:"div3"`
	DPE        bool   `gorm:"not null;default:false" json:"dpe"`
	QFR        bool   `gorm:"not null;default:false" json:"qfr"`
	WAC        bool   `gorm:"not null;default:false" json:"wac"`

	Prereqs                string `gorm:"size:65535" json:"prereqs"`
	EnrollmentPreferences  string `gorm:"size:65535" json:"enrollmentPreferences"`
	RequirementsEvaluation string `gorm:"size:65535" json:"requirementsEvaluation"`
	DepartmentNotes        string `gorm:"size:65535" json:"departmentNotes"`
	WACNotes               string `gorm:"size:65535" json:"wacNotes"` // writing-intensive notes
	DPENotes               string `gorm:"size:65535" json:"dpeNotes"` // difference, power and equity notes
	QFRNotes               string `gorm:"size:65535" json:"qfrNotes"` // quantitative/formal reasoning notes
	ExtraInfo              string `gorm:"size:65535" json:"extraInfo"`
	MaterialFee            string `gorm:"size:65535" json:"materialFee"`

	Offerings []*CourseOffering `json:"offerings,omitempty"`
}

func (*CourseInSemester) TableName() string {
	return "course_in_semesters"
}

type CourseOffering struct {
	BaseSchema
	CourseCanonicalID  uint              `gorm:"index:idx_course_offerings_course;not null" json:"courseCanonicalID"`
	CourseCanonical    *CourseCanonical  `json:"courseCanonical,omitempty"`
	CourseInSemesterID *uint             `gorm:"index:idx_course_offerings_course_in_semester" json:"courseInSemesterID"`
	CourseInSemester   *CourseInSemester `json:"courseInSemester,omitempty"`
	Source             string            `gorm:"not null;size:16;default:'catalog'" json:"source"`

	Strm      int    `gorm:"unique_index:idx_course_offerings_strm_class;not null" json:"strm"`
	Year      int    `gorm:"not null" json:"year"`
	Term      string `gorm:"not null;size:16" json:"term"`
	Section   string `gorm:"not null;size:16" json:"section"`
	ClassNbr  int    `gorm:"unique_index:idx_course_offerings_strm_class;not null" json:"classNbr"`
	Title     string `gorm:"not null" json:"title"`
	Component string `gorm:"size:32" json:"component"`
	Status    string `gorm:"not null;size:32" json:"status"`

	// Cross-listed classes have one class number per code (MATH 373 is 1263, PHYS 373 is 1264)
	Subject       string `gorm:"size:16" json:"subject"`
	CatalogNumber int    `json:"catalogNumber"`
	CatalogLetter string `gorm:"size:8" json:"catalogLetter"`

	// Cross-listed classes can differ by code: ENGL 215 is Div I, AAS 215 is Div II
	Attributes string `gorm:"size:1024" json:"attributes"`
	Div1       bool   `gorm:"not null;default:false" json:"div1"`
	Div2       bool   `gorm:"not null;default:false" json:"div2"`
	Div3       bool   `gorm:"not null;default:false" json:"div3"`
	DPE        bool   `gorm:"not null;default:false" json:"dpe"`
	QFR        bool   `gorm:"not null;default:false" json:"qfr"`
	WAC        bool   `gorm:"not null;default:false" json:"wac"`

	Consent            string `gorm:"size:8" json:"consent"` // N (none), I (instructor), D (department)
	EnrollmentLimit    *int   `json:"enrollmentLimit"`
	ExpectedEnrollment *int   `json:"expectedEnrollment"`

	Meetings    []*CourseOfferingMeeting    `json:"meetings,omitempty"`
	Instructors []*CourseOfferingInstructor `json:"instructors,omitempty"`
}

func (*CourseOffering) TableName() string {
	return "course_offerings"
}

type CourseOfferingMeeting struct {
	BaseSchema
	CourseOfferingID uint            `gorm:"index:idx_course_offering_meetings_course_offering;not null" json:"offeringID"`
	CourseOffering   *CourseOffering `json:"offering,omitempty"`

	Days     string `gorm:"size:16" json:"days"`
	Start    string `gorm:"size:8" json:"start"`
	End      string `gorm:"size:8" json:"end"`
	Facility string `gorm:"size:255" json:"facility"`
}

func (*CourseOfferingMeeting) TableName() string {
	return "course_offering_meetings"
}

type CourseOfferingInstructor struct {
	BaseSchema
	CourseOfferingID uint            `gorm:"index:idx_course_offering_instructors_course_offering;not null" json:"offeringID"`
	CourseOffering   *CourseOffering `json:"offering,omitempty"`

	UnixID       string `gorm:"index:idx_course_offering_instructors_unix;size:100" json:"unixID"`
	UserID       *uint  `gorm:"index:idx_course_offering_instructors_user" json:"userID"`
	User         *User  `json:"user,omitempty"`
	NameAsListed string `gorm:"not null" json:"nameAsListed"`
}

func (*CourseOfferingInstructor) TableName() string {
	return "course_offering_instructors"
}
