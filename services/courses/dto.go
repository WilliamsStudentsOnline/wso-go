package courses

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/models"
)

type TermResponse struct {
	Strm        int    `json:"strm"`
	Year        int    `json:"year"` // catalog year (spring calendar year): Fall 2026 is 2027
	Term        string `json:"term"` // Fall, Winter or Spring
	Label       string `json:"label"`
	Current     bool   `json:"current"`
	CourseCount int    `json:"courseCount,omitempty"`
}

type ListingResponse struct {
	Code    string `json:"code"`
	Subject string `json:"subject"`
	Number  int    `json:"number"`
	Letter  string `json:"letter,omitempty"`
}

type MeetingResponse struct {
	Days     string `json:"days"`
	Start    string `json:"start,omitempty"`
	End      string `json:"end,omitempty"`
	Facility string `json:"facility,omitempty"`
}

type InstructorResponse struct {
	Name   string `json:"name"`
	UserID *uint  `json:"userID,omitempty"` // WSO user, when matched
}

type SectionResponse struct {
	ClassNbr           int                  `json:"classNbr"`
	Code               string               `json:"code,omitempty"` // the code this class number is under; cross-listed classes have one section per code
	Section            string               `json:"section"`
	Component          string               `json:"component"` // LEC, SEM, LAB, TUT, ...
	Status             string               `json:"status"`    // offered or cancelled
	Consent            string               `json:"consent,omitempty"`
	EnrollmentLimit    *int                 `json:"enrollmentLimit,omitempty"`
	ExpectedEnrollment *int                 `json:"expectedEnrollment,omitempty"`
	Meetings           []MeetingResponse    `json:"meetings"`
	Instructors        []InstructorResponse `json:"instructors"`
}

type TermCourseResponse struct {
	CrseID           string            `json:"crseID"`
	Strm             int               `json:"strm"`
	Title            string            `json:"title"`
	ShortDescription string            `json:"shortDescription,omitempty"`
	Description      string            `json:"description,omitempty"` // only with include=description
	Listings         []ListingResponse `json:"listings"`
	Status           string            `json:"status"`
	Source           string            `json:"source"` // catalog, or last_offered for recovered history
	Components       string            `json:"components,omitempty"`
	ClassFormat      string            `json:"classFormat,omitempty"`
	GradingBasis     string            `json:"gradingBasis,omitempty"`
	GradingBasisDesc string            `json:"gradingBasisDesc,omitempty"`
	PassFail         bool              `json:"passFail"`
	FifthCourse      bool              `json:"fifthCourse"`
	Units            string            `json:"units,omitempty"`
	Campus           string            `json:"campus,omitempty"`
	Div1             bool              `json:"div1"`
	Div2             bool              `json:"div2"`
	Div3             bool              `json:"div3"`
	DPE              bool              `json:"dpe"`
	QFR              bool              `json:"qfr"`
	WAC              bool              `json:"wac"`
	Attributes       []string          `json:"attributes"`
	Sections         []SectionResponse `json:"sections"`
	// Only when the codes count differently (ENGL 215 is Div I, AAS 215 is Div II); the flags above are the union
	ByCode []CodeDistributions `json:"byCode,omitempty"`
}

type CodeDistributions struct {
	Code string `json:"code"`
	Div1 bool   `json:"div1"`
	Div2 bool   `json:"div2"`
	Div3 bool   `json:"div3"`
	DPE  bool   `json:"dpe"`
	QFR  bool   `json:"qfr"`
	WAC  bool   `json:"wac"`
}

func codeDistributions(offerings []*models.CourseOffering) []CodeDistributions {
	var out []CodeDistributions
	seen := map[string]bool{}
	for _, o := range offerings {
		code := offeringCode(o)
		if code == "" || seen[code] || o.Status != models.CourseOfferingStatusOffered {
			continue
		}
		seen[code] = true
		out = append(out, CodeDistributions{Code: code, Div1: o.Div1, Div2: o.Div2, Div3: o.Div3, DPE: o.DPE, QFR: o.QFR, WAC: o.WAC})
	}
	for i := 1; i < len(out); i++ {
		a, b := out[0], out[i]
		a.Code, b.Code = "", ""
		if a != b {
			return out
		}
	}
	return nil
}

type TermCatalogResponse struct {
	Term    TermResponse         `json:"term"`
	Courses []TermCourseResponse `json:"courses"`
}

type CourseInSemesterResponse struct {
	TermCourseResponse
	Term                   TermResponse `json:"term"`
	Description            string       `json:"description,omitempty"`
	Prereqs                string       `json:"prereqs,omitempty"`
	EnrollmentPreferences  string       `json:"enrollmentPreferences,omitempty"`
	RequirementsEvaluation string       `json:"requirementsEvaluation,omitempty"`
	DepartmentNotes        string       `json:"departmentNotes,omitempty"`
	WACNotes               string       `json:"wacNotes,omitempty"`
	DPENotes               string       `json:"dpeNotes,omitempty"`
	QFRNotes               string       `json:"qfrNotes,omitempty"`
	ExtraInfo              string       `json:"extraInfo,omitempty"`
	MaterialFee            string       `json:"materialFee,omitempty"`
}

type CourseTermRef struct {
	Strm   int    `json:"strm"`
	Label  string `json:"label"`
	Status string `json:"status"`
	Source string `json:"source"`
}

type CourseHistoryEntry struct {
	CourseTermRef
	Title       string               `json:"title"`
	Instructors []InstructorResponse `json:"instructors"`
}

type CanonicalListingResponse struct {
	ListingResponse
	FirstValidYear int `json:"firstValidYear"`
	LastValidYear  int `json:"lastValidYear"`
}

type CourseResponse struct {
	CrseID      string                     `json:"crseID"`
	Title       string                     `json:"title"`
	Description string                     `json:"description,omitempty"`
	Listings    []CanonicalListingResponse `json:"listings"`
	History     []CourseHistoryEntry       `json:"history"` // newest first
	Professors  []InstructorResponse       `json:"professors"`
}

type CourseSearchResult struct {
	CrseID      string            `json:"crseID"`
	Title       string            `json:"title"`
	Listings    []ListingResponse `json:"listings"`
	Terms       []CourseTermRef   `json:"terms"` // newest first
	LastOffered *CourseTermRef    `json:"lastOffered,omitempty"`
}

// Catalog year 2027 Fall is "Fall 2026"
func TermLabel(year int, term string) string {
	if term == "Fall" {
		return fmt.Sprintf("Fall %d", year-1)
	}
	return fmt.Sprintf("%s %d", term, year)
}

// Over the summer, the coming fall
func CurrentTerm(now time.Time) (year int, term string) {
	switch m := now.Month(); {
	case m >= time.September:
		return now.Year() + 1, "Fall"
	case m == time.January:
		return now.Year(), "Winter"
	case m <= time.May:
		return now.Year(), "Spring"
	default:
		return now.Year() + 1, "Fall"
	}
}

func termResponse(year int, term string, strm int, now time.Time) TermResponse {
	cy, ct := CurrentTerm(now)
	return TermResponse{
		Strm:    strm,
		Year:    year,
		Term:    term,
		Label:   TermLabel(year, term),
		Current: cy == year && ct == term,
	}
}

func areaCatalogSubjects(abbrev string) []string {
	subjects := []string{strings.ToUpper(abbrev)}
	if alias, ok := models.RenamedSubjects[subjects[0]]; ok {
		subjects = append(subjects, alias)
	}
	return subjects
}

func listingCode(subject string, number int, letter string) string {
	return fmt.Sprintf("%s %d%s", subject, number, letter)
}

func listingsForYear(listings []*models.CourseListing, year int) []ListingResponse {
	var valid, all []ListingResponse
	for _, l := range listings {
		r := ListingResponse{Code: listingCode(l.Subject, l.Number, l.Letter), Subject: l.Subject, Number: l.Number, Letter: l.Letter}
		all = append(all, r)
		if year == 0 || (year >= l.FirstValidYear && year <= l.LastValidYear) {
			valid = append(valid, r)
		}
	}
	if len(valid) == 0 {
		valid = all
	}
	sort.Slice(valid, func(i, j int) bool { return valid[i].Code < valid[j].Code })
	if valid == nil {
		valid = []ListingResponse{}
	}
	return valid
}

func instructorResponses(instructors []*models.CourseOfferingInstructor) []InstructorResponse {
	out := make([]InstructorResponse, 0, len(instructors))
	for _, in := range instructors {
		out = append(out, InstructorResponse{Name: in.NameAsListed, UserID: in.UserID})
	}
	return out
}

func offeringCode(o *models.CourseOffering) string {
	if o.Subject == "" {
		return ""
	}
	return listingCode(o.Subject, o.CatalogNumber, o.CatalogLetter)
}

func sectionResponses(offerings []*models.CourseOffering) []SectionResponse {
	out := make([]SectionResponse, 0, len(offerings))
	for _, o := range offerings {
		s := SectionResponse{
			ClassNbr:           o.ClassNbr,
			Code:               offeringCode(o),
			Section:            o.Section,
			Component:          o.Component,
			Status:             o.Status,
			Consent:            o.Consent,
			EnrollmentLimit:    o.EnrollmentLimit,
			ExpectedEnrollment: o.ExpectedEnrollment,
			Meetings:           make([]MeetingResponse, 0, len(o.Meetings)),
			Instructors:        instructorResponses(o.Instructors),
		}
		for _, m := range o.Meetings {
			s.Meetings = append(s.Meetings, MeetingResponse{Days: m.Days, Start: m.Start, End: m.End, Facility: m.Facility})
		}
		out = append(out, s)
	}
	return out
}

func splitAttributes(attrs string) []string {
	out := []string{}
	for _, a := range strings.Split(attrs, ",") {
		if a = strings.TrimSpace(a); a != "" {
			out = append(out, a)
		}
	}
	return out
}

func termCourseResponse(cis *models.CourseInSemester, withDescription bool) TermCourseResponse {
	gradingDesc, passFail, fifth := models.GradingBasisOptions(cis.GradingBasis)
	r := TermCourseResponse{
		Strm:             cis.Strm,
		Title:            cis.Title,
		ShortDescription: cis.ShortDescription,
		Status:           cis.Status,
		Source:           cis.Source,
		Components:       cis.Components,
		ClassFormat:      cis.ClassFormat,
		GradingBasis:     cis.GradingBasis,
		GradingBasisDesc: gradingDesc,
		PassFail:         passFail,
		FifthCourse:      fifth,
		Units:            cis.Units,
		Campus:           cis.Campus,
		Div1:             cis.Div1,
		Div2:             cis.Div2,
		Div3:             cis.Div3,
		DPE:              cis.DPE,
		QFR:              cis.QFR,
		WAC:              cis.WAC,
		Attributes:       splitAttributes(cis.Attributes),
		Sections:         sectionResponses(cis.Offerings),
		ByCode:           codeDistributions(cis.Offerings),
		Listings:         []ListingResponse{},
	}
	if withDescription {
		r.Description = cis.Description
	}
	if cis.CourseCanonical != nil {
		r.CrseID = cis.CourseCanonical.CrseID
		r.Listings = listingsForYear(cis.CourseCanonical.Listings, cis.Year)
	}
	return r
}

func courseInSemesterResponse(cis *models.CourseInSemester, now time.Time) CourseInSemesterResponse {
	return CourseInSemesterResponse{
		TermCourseResponse:     termCourseResponse(cis, false),
		Term:                   termResponse(cis.Year, cis.Term, cis.Strm, now),
		Description:            cis.Description,
		Prereqs:                cis.Prereqs,
		EnrollmentPreferences:  cis.EnrollmentPreferences,
		RequirementsEvaluation: cis.RequirementsEvaluation,
		DepartmentNotes:        cis.DepartmentNotes,
		WACNotes:               cis.WACNotes,
		DPENotes:               cis.DPENotes,
		QFRNotes:               cis.QFRNotes,
		ExtraInfo:              cis.ExtraInfo,
		MaterialFee:            cis.MaterialFee,
	}
}

func termRef(cis *models.CourseInSemester) CourseTermRef {
	return CourseTermRef{Strm: cis.Strm, Label: TermLabel(cis.Year, cis.Term), Status: cis.Status, Source: cis.Source}
}

func courseResponse(c *models.CourseCanonical) CourseResponse {
	r := CourseResponse{
		CrseID:      c.CrseID,
		Title:       c.Title,
		Description: c.Description,
		Listings:    make([]CanonicalListingResponse, 0, len(c.Listings)),
		History:     make([]CourseHistoryEntry, 0, len(c.CoursesInSemester)),
		Professors:  []InstructorResponse{},
	}
	for _, l := range c.Listings {
		r.Listings = append(r.Listings, CanonicalListingResponse{
			ListingResponse: ListingResponse{Code: listingCode(l.Subject, l.Number, l.Letter), Subject: l.Subject, Number: l.Number, Letter: l.Letter},
			FirstValidYear:  l.FirstValidYear,
			LastValidYear:   l.LastValidYear,
		})
	}
	sort.Slice(r.Listings, func(i, j int) bool { return r.Listings[i].Code < r.Listings[j].Code })

	seen := map[string]bool{}
	for _, cis := range c.CoursesInSemester {
		entry := CourseHistoryEntry{CourseTermRef: termRef(cis), Title: cis.Title, Instructors: []InstructorResponse{}}
		inTerm := map[string]bool{}
		for _, o := range cis.Offerings {
			for _, in := range instructorResponses(o.Instructors) {
				key := in.Name
				if in.UserID != nil {
					key = fmt.Sprint(*in.UserID)
				}
				if !inTerm[key] {
					inTerm[key] = true
					entry.Instructors = append(entry.Instructors, in)
				}
				if !seen[key] {
					seen[key] = true
					r.Professors = append(r.Professors, in)
				}
			}
		}
		r.History = append(r.History, entry)
	}
	return r
}

func searchResult(c *models.CourseCanonical) CourseSearchResult {
	r := CourseSearchResult{
		CrseID:   c.CrseID,
		Title:    c.Title,
		Listings: listingsForYear(c.Listings, 0),
		Terms:    make([]CourseTermRef, 0, len(c.CoursesInSemester)),
	}
	for _, cis := range c.CoursesInSemester {
		ref := termRef(cis)
		r.Terms = append(r.Terms, ref)
		if r.LastOffered == nil && cis.Status == models.CourseOfferingStatusOffered {
			last := ref
			r.LastOffered = &last
		}
	}
	return r
}

type AttributeResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Kind string `json:"kind"`           // distribution or program
	Area string `json:"area,omitempty"` // owning area's abbreviation
}

type AreaResponse struct {
	ID               uint                `json:"id"`
	Abbreviation     string              `json:"abbreviation"`
	Name             string              `json:"name"`
	CatalogSubjects  []string            `json:"catalogSubjects"` // listing subjects for this area, e.g. WGST covers WGSS
	CatalogKind      string              `json:"catalogKind,omitempty"`
	CatalogURL       string              `json:"catalogURL,omitempty"` // set when the area is in the current catalog
	CurriculumPDFURL string              `json:"curriculumPDFURL,omitempty"`
	WebsiteURL       string              `json:"websiteURL,omitempty"`
	Description      string              `json:"description,omitempty"`
	Attributes       []AttributeResponse `json:"attributes"`
}

type CourseFactrakResponse struct {
	CrseID            string `json:"crseID"`
	Reviews           int    `json:"reviews"`           // reviews that answered "would recommend"
	Recommends        int    `json:"recommends"`        // of those, how many said yes
	CurrentReviews    int    `json:"currentReviews"`    // reviews of this term's instructors
	CurrentRecommends int    `json:"currentRecommends"` // Course Scheduler's score is currentRecommends/currentReviews
}

// 1YYT, e.g. 1271 is catalog year 2027
func catalogYearOfStrm(strm int) (int, bool) {
	if strm < 1000 || strm > 1999 {
		return 0, false
	}
	return 2000 + (strm/10)%100, true
}

func termNameOfStrm(strm int) string {
	switch strm % 10 {
	case 1:
		return "Fall"
	case 2:
		return "Winter"
	case 3:
		return "Spring"
	default:
		return ""
	}
}
