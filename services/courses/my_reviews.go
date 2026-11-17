package courses

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

const maxLinkSuggestions = 12

type ReviewSectionRef struct {
	CrseID    string            `json:"crseID"`
	Title     string            `json:"title"`
	Strm      int               `json:"strm"`
	TermLabel string            `json:"termLabel"`
	ClassNbr  int               `json:"classNbr"`
	Code      string            `json:"code,omitempty"` // e.g. "MATH 373" (cross-listed classes have one class number per code)
	Section   string            `json:"section"`
	Component string            `json:"component"`
	Meetings  []MeetingResponse `json:"meetings"` // when and where it met, to help pick the right one
}

type MyReviewResponse struct {
	ID          uint   `json:"id"`
	ProfessorID uint   `json:"professorID"`
	Professor   string `json:"professor"`
	Course      string `json:"course"`           // the Factrak course it was filed under, e.g. "CSCI 136"
	CrseID      string `json:"crseID,omitempty"` // the catalog course it's about (linked or our best guess)
	Comment     string `json:"comment"`
	Semester    string `json:"semester,omitempty"`
	// The survey's yes/no answers, when given
	WouldRecommendCourse         *bool     `json:"wouldRecommendCourse,omitempty"`
	WouldTakeAnother             *bool     `json:"wouldTakeAnother,omitempty"`
	CreatedTime                  time.Time `json:"createdTime"`
	CourseLinkVerificationSource string    `json:"courseLinkVerificationSource,omitempty"`
	// The section it's linked to, by exact match or because the author said so
	Linked *ReviewSectionRef `json:"linked,omitempty"`
	// Sections of our best-guess course this professor taught, when not linked
	Suggestions []ReviewSectionRef `json:"suggestions"`
}

type LinkMyReviewRequest struct {
	Strm     int `json:"strm" binding:"required"`
	ClassNbr int `json:"classNbr" binding:"required"`
}

func sectionRef(o *models.CourseOffering, crseID, title string) ReviewSectionRef {
	r := ReviewSectionRef{CrseID: crseID, Title: title, Strm: o.Strm, ClassNbr: o.ClassNbr, Code: offeringCode(o), Section: o.Section, Component: o.Component}
	if year, ok := catalogYearOfStrm(o.Strm); ok {
		r.TermLabel = TermLabel(year, termNameOfStrm(o.Strm))
	}
	if r.Title == "" {
		r.Title = o.Title
	}
	r.Meetings = make([]MeetingResponse, 0, len(o.Meetings))
	for _, m := range o.Meetings {
		r.Meetings = append(r.Meetings, MeetingResponse{Days: m.Days, Start: m.Start, End: m.End, Facility: m.Facility})
	}
	return r
}

// Cross-listed sections repeat under each code; prefer the one in filedUnder
func oneCopyPerSection(sections []*models.CourseOffering, filedUnder map[string]bool) []*models.CourseOffering {
	type key struct {
		strm               int
		section, component string
	}
	pick := map[key]int{}
	var out []*models.CourseOffering
	for _, o := range sections {
		k := key{o.Strm, o.Section, o.Component}
		i, seen := pick[k]
		switch {
		case !seen:
			pick[k] = len(out)
			out = append(out, o)
		case filedUnder[o.Subject] && !filedUnder[out[i].Subject]:
			out[i] = o
		}
	}
	return out
}

func (t *Controller) myReviewResponse(s *models.FactrakSurvey) (MyReviewResponse, error) {
	r := MyReviewResponse{
		ID: s.ID, ProfessorID: s.ProfessorID, Comment: s.Comment, CreatedTime: s.CreatedAt,
		CourseLinkVerificationSource: s.CourseLinkVerificationSource, Suggestions: []ReviewSectionRef{},
		WouldRecommendCourse: s.WouldRecommendCourse, WouldTakeAnother: s.WouldTakeAnother,
	}
	if s.Professor != nil {
		r.Professor = s.Professor.Name
	}
	r.Semester = surveySemesterLabel(s.SemesterSeason, s.SemesterYear)
	if c := s.Course; c != nil {
		r.Course = c.Number
		if c.AreaOfStudy != nil {
			r.Course = c.AreaOfStudy.Abbreviation + " " + c.Number
		}
	}
	crseID, title := "", ""
	if s.CourseCanonical != nil {
		crseID, title = s.CourseCanonical.CrseID, s.CourseCanonical.Title
		r.CrseID = crseID
	}
	trusted := s.CourseLinkVerificationSource == models.CourseLinkVerificationSourceVerified || s.CourseLinkVerificationSource == models.CourseLinkVerificationSourceAuthorConfirmed
	if s.CourseOffering != nil && trusted {
		ref := sectionRef(s.CourseOffering, crseID, title)
		r.Linked = &ref
		return r, nil
	}
	if s.CourseCanonicalID != nil {
		sections, err := t.catalog.SectionsTaughtBy(*s.CourseCanonicalID, s.ProfessorID, t.now(), 5*maxLinkSuggestions)
		if err != nil {
			return r, err
		}
		sections = oneCopyPerSection(sections, t.filedUnderSubjects(s))
		if len(sections) > maxLinkSuggestions {
			sections = sections[:maxLinkSuggestions]
		}
		for _, o := range sections {
			r.Suggestions = append(r.Suggestions, sectionRef(o, crseID, title))
		}
	}
	return r, nil
}

// ListMyReviews godoc
// @Summary List my Factrak reviews
// @Description The signed-in user's own reviews, newest first. Each says which section it's linked to or, if none, suggests sections of the course we think it's about that its professor taught.
// @ID listMyReviews
// @Tags courses
// @Produce json
// @Success 200 {array} courses.MyReviewResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /courses/mine/reviews [get]
func (t *Controller) ListMyReviews(c *gin.Context) {
	surveys, err := t.catalog.ListUserSurveys(services.GetUserID(c))
	if err != nil {
		t.RespondError(c, err)
		return
	}
	out := make([]MyReviewResponse, 0, len(surveys))
	for _, s := range surveys {
		r, err := t.myReviewResponse(s)
		if err != nil {
			t.RespondError(c, err)
			return
		}
		out = append(out, r)
	}
	t.RespondOK(c, out)
}

// LinkMyReview godoc
// @Summary Link my review to a section
// @Description Records the section (term and class number) one of your reviews was about. The section must be of the review's course, taught by its professor, in a term that's over. The review's course and term follow the section, and the section goes on your My courses.
// @ID linkMyReview
// @Tags courses
// @Accept json
// @Produce json
// @Param surveyID path int true "Review (survey) ID"
// @Param body body courses.LinkMyReviewRequest true "Section"
// @Success 200 {object} courses.MyReviewResponse
// @Failure 400 {object} services.BaseErrorResponse "1100: could not parse malformed request data; 1101: request data validation failed; 1541: no offered catalog section with that term and class number; 1542: professor did not teach that section; 2830: courses can be reviewed once their term has started; 2831: a review can only be linked to a section of its own course"
// @Failure 404 {object} services.BaseErrorResponse "404: record not found"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /courses/mine/reviews/{surveyID}/link [post]
func (t *Controller) LinkMyReview(c *gin.Context) {
	surveyID, err := services.GetUIntParam(c, "surveyID")
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}
	var req LinkMyReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		t.RespondBadBind(c, err)
		return
	}
	userID := services.GetUserID(c)
	var survey *models.FactrakSurvey
	surveys, err := t.catalog.ListUserSurveys(userID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	for _, s := range surveys {
		if s.ID == surveyID {
			survey = s
		}
	}
	if survey == nil {
		t.RespondAPIError(c, lib.ErrorRecordNotFound)
		return
	}
	section, err := t.catalog.GetReviewSection(req.Strm, req.ClassNbr, t.now())
	switch {
	case errors.Is(err, models.ErrReviewSectionNotFound):
		t.RespondAPIError(c, lib.ErrorSurveySectionNotFound)
		return
	case errors.Is(err, models.ErrReviewSectionNotStarted):
		t.RespondAPIError(c, lib.ErrorReviewTermNotStarted)
		return
	case errors.Is(err, models.ErrReviewSectionNoLegacyCourse):
		// Linking doesn't need a Factrak course; the review keeps the one it has
	case err != nil:
		t.RespondError(c, err)
		return
	}
	if survey.CourseCanonicalID != nil && *survey.CourseCanonicalID != section.CourseOffering.CourseCanonicalID {
		t.RespondAPIError(c, lib.ErrorReviewOtherCourse)
		return
	}
	if !section.TaughtBy(survey.ProfessorID) {
		t.RespondAPIError(c, lib.ErrorSurveySectionProfessor)
		return
	}
	if err := t.catalog.LinkSurveyToSection(survey, section); err != nil {
		t.RespondError(c, err)
		return
	}
	surveys, err = t.catalog.ListUserSurveys(userID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	for _, s := range surveys {
		if s.ID == surveyID {
			r, err := t.myReviewResponse(s)
			if err != nil {
				t.RespondError(c, err)
				return
			}
			t.RespondOK(c, r)
			return
		}
	}
	t.RespondAPIError(c, lib.ErrorRecordNotFound)
}

// Includes the catalog's code when it differs (WGST is WGSS)
func (t *Controller) filedUnderSubjects(s *models.FactrakSurvey) map[string]bool {
	out := map[string]bool{}
	if s.Course == nil || s.Course.AreaOfStudy == nil {
		return out
	}
	abbrev := strings.ToUpper(s.Course.AreaOfStudy.Abbreviation)
	out[abbrev] = true
	if subject, ok := models.RenamedSubjects[abbrev]; ok {
		out[subject] = true
	}
	return out
}

func surveySemesterLabel(season *string, year *int) string {
	if season == nil || year == nil || *year == 0 {
		return ""
	}
	switch *season {
	case models.FactrakSurveySemesterSeasonFall:
		return fmt.Sprintf("Fall %d", *year)
	case models.FactrakSurveySemesterSeasonWinterStudy:
		return fmt.Sprintf("Winter Study %d", *year)
	case models.FactrakSurveySemesterSeasonSpring:
		return fmt.Sprintf("Spring %d", *year)
	}
	return ""
}
