package courses

import (
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/WilliamsStudentsOnline/wso-go/services/factrak"
	"github.com/gin-gonic/gin"
)

type ProfessorSummaryResponse struct {
	ID               uint   `json:"id"`
	Name             string `json:"name"`
	UnixID           string `json:"unixID"`
	Title            string `json:"title,omitempty"`
	Department       string `json:"department,omitempty"`
	AtWilliams       bool   `json:"atWilliams"`
	TeachingThisYear bool   `json:"teachingThisYear"` // teaches an offered section this academic year
}

type ProfessorCourseResponse struct {
	CrseID      string            `json:"crseID"`
	Title       string            `json:"title"`
	Listings    []ListingResponse `json:"listings"`
	LastTaught  *CourseTermRef    `json:"lastTaught,omitempty"` // empty when only reviewed
	TermsTaught int               `json:"termsTaught"`
	Terms       []CourseTermRef   `json:"terms"`   // every term taught, newest first
	Reviews     int               `json:"reviews"` // reviews Factrak shows, without mismatched catalog links
}

type ProfessorResponse struct {
	ProfessorSummaryResponse
	Office  string                    `json:"office,omitempty"` // room, e.g. "TBL 301"; only for visible professors at Williams
	Courses []ProfessorCourseResponse `json:"courses"`          // most recently taught first
}

func (t *Controller) professorSummaries(profs []*models.User) ([]ProfessorSummaryResponse, error) {
	ids := make([]uint, 0, len(profs))
	for _, p := range profs {
		ids = append(ids, p.ID)
	}
	year, _ := CurrentTerm(t.now())
	teaching, err := t.catalog.ProfessorsTeachingInYear(ids, year)
	if err != nil {
		return nil, err
	}
	out := make([]ProfessorSummaryResponse, 0, len(profs))
	for _, p := range profs {
		r := ProfessorSummaryResponse{
			ID: p.ID, Name: p.Name, UnixID: p.UnixID,
			// The directory flag misses visiting faculty and staff who teach
			AtWilliams:       (p.AtWilliams != nil && *p.AtWilliams) || teaching[p.ID],
			TeachingThisYear: teaching[p.ID],
		}
		if p.Title != nil {
			r.Title = *p.Title
		}
		if p.Department != nil {
			r.Department = p.Department.Name
		}
		out = append(out, r)
	}
	return out, nil
}

type SearchProfessorsParams struct {
	Q      string `form:"q"`
	Area   string `form:"area"`
	Limit  int    `form:"limit"`
	Offset int    `form:"offset"`
}

// SearchProfessors godoc
// @Summary Search professors
// @Description Professors by name (every word must match), including ones who have left Williams. Current professors first. With area, only professors who taught an offered section listed under the area's catalog subjects in the last 3 academic years, ordered by name.
// @ID searchProfessors
// @Tags courses
// @Produce json
// @Param q query string false "Name"
// @Param area query string false "Area of study abbreviation, e.g. AFR"
// @Param limit query int false "Default 25, max 100"
// @Param offset query int false "Offset pagination"
// @Success 200 {array} courses.ProfessorSummaryResponse
// @Failure 400 {object} services.BaseErrorResponse "1100: could not parse malformed request data"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /courses/professors [get]
func (t *Controller) SearchProfessors(c *gin.Context) {
	var params SearchProfessorsParams
	if err := c.ShouldBindQuery(&params); err != nil {
		t.RespondBadBind(c, err)
		return
	}
	var profs []*models.User
	var total int
	var err error
	if params.Area != "" {
		year, _ := CurrentTerm(t.now())
		profs, total, err = t.catalog.AreaProfessors(areaCatalogSubjects(params.Area), year-2, year, params.Q, params.Limit, params.Offset)
	} else {
		profs, total, err = t.catalog.SearchProfessors(params.Q, params.Limit, params.Offset)
	}
	if err != nil {
		t.RespondError(c, err)
		return
	}
	out, err := t.professorSummaries(profs)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	t.SetPaginationTotal(c, total)
	t.RespondOK(c, out)
}

// GetProfessor godoc
// @Summary Get a professor
// @Description A professor's page: name, title, department, office (visible professors at Williams only), whether they're at Williams and teaching this year, and every catalog course they taught or have Factrak reviews for.
// @ID getCourseProfessor
// @Tags courses
// @Produce json
// @Param professorID path int true "Professor (WSO user) ID"
// @Success 200 {object} courses.ProfessorResponse
// @Failure 400 {object} services.BaseErrorResponse "1100: could not parse malformed request data"
// @Failure 404 {object} services.BaseErrorResponse "404: record not found"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /courses/professors/{professorID} [get]
func (t *Controller) GetProfessor(c *gin.Context) {
	profID, err := services.GetUIntParam(c, "professorID")
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}
	prof, err := t.catalog.GetProfessor(profID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	summaries, err := t.professorSummaries([]*models.User{prof})
	if err != nil {
		t.RespondError(c, err)
		return
	}
	courses, err := t.catalog.ProfessorCourses(profID, models.FactrakReviewWindow(t.now()))
	if err != nil {
		t.RespondError(c, err)
		return
	}
	resp := ProfessorResponse{ProfessorSummaryResponse: summaries[0], Courses: make([]ProfessorCourseResponse, 0, len(courses))}
	// Same rule as Facebook: only show where to find people who are here and visible
	if prof.Office != nil && prof.Visible != nil && *prof.Visible && summaries[0].AtWilliams {
		resp.Office = prof.Office.Number
	}
	for _, pc := range courses {
		resp.Courses = append(resp.Courses, professorCourseResponse(pc))
	}
	t.RespondOK(c, resp)
}

func professorCourseResponse(pc *models.ProfessorCourse) ProfessorCourseResponse {
	r := ProfessorCourseResponse{
		CrseID: pc.CrseID, Title: pc.Title, Listings: listingsForYear(pc.Listings, 0),
		TermsTaught: pc.TermsTaught, Reviews: pc.Reviews,
	}
	r.Terms = make([]CourseTermRef, 0, len(pc.TaughtStrms))
	for _, strm := range pc.TaughtStrms {
		year, _ := catalogYearOfStrm(strm)
		r.Terms = append(r.Terms, CourseTermRef{Strm: strm, Label: TermLabel(year, termNameOfStrm(strm)), Status: models.CourseOfferingStatusOffered})
	}
	if len(r.Terms) > 0 {
		last := r.Terms[0]
		r.LastTaught = &last
	}
	return r
}

type ProfessorFactrakParams struct {
	CrseID string `form:"crseID"`
	Limit  uint   `form:"limit"`
	Offset uint   `form:"offset"`
}

func (t *Controller) professorAndCourse(c *gin.Context) (uint, *uint, ProfessorFactrakParams, bool) {
	var params ProfessorFactrakParams
	profID, err := services.GetUIntParam(c, "professorID")
	if err == nil {
		err = c.ShouldBindQuery(&params)
	}
	if err != nil {
		t.RespondBadBind(c, err)
		return 0, nil, params, false
	}
	if _, err := t.catalog.GetProfessor(profID); err != nil {
		t.RespondError(c, err)
		return 0, nil, params, false
	}
	if params.CrseID == "" {
		return profID, nil, params, true
	}
	courseID, err := t.catalog.CanonicalID(params.CrseID)
	if err != nil {
		t.RespondError(c, err)
		return 0, nil, params, false
	}
	return profID, &courseID, params, true
}

// GetProfessorRatings godoc
// @Summary Get a professor's Factrak ratings
// @Description Averaged Factrak ratings for a professor, like Factrak's professor ratings (reviews from the last 5 years). With crseID, only reviews of that catalog course (all its cross-listings). Needs full Factrak access.
// @ID getCourseProfessorRatings
// @Tags courses
// @Produce json
// @Param professorID path int true "Professor (WSO user) ID"
// @Param crseID query string false "Only reviews of this catalog course"
// @Success 200 {object} models.FactrakSurveyAvgRatings
// @Failure 400 {object} services.BaseErrorResponse "1100: could not parse malformed request data"
// @Failure 404 {object} services.BaseErrorResponse "404: record not found"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /courses/professors/{professorID}/ratings [get]
func (t *Controller) GetProfessorRatings(c *gin.Context) {
	profID, courseID, _, ok := t.professorAndCourse(c)
	if !ok {
		return
	}
	var ratings models.FactrakSurveyAvgRatings
	if err := t.surveys.GetSurveyRatingsForProfessor(profID, courseID, &ratings); err != nil {
		t.RespondError(c, err)
		return
	}
	t.RespondOK(c, ratings)
}

// ListProfessorReviews godoc
// @Summary List a professor's Factrak reviews
// @Description A professor's Factrak reviews, newest first, optionally for one catalog course (all its cross-listings). Same rules as Factrak's survey list: full Factrak access, last 5 years, professors still at Williams, authors hidden.
// @ID listCourseProfessorReviews
// @Tags courses
// @Produce json
// @Param professorID path int true "Professor (WSO user) ID"
// @Param crseID query string false "Only reviews of this catalog course"
// @Param limit query int false "Default 20, max 100"
// @Param offset query int false "Offset pagination"
// @Success 200 {array} models.FactrakSurvey
// @Failure 400 {object} services.BaseErrorResponse "1100: could not parse malformed request data"
// @Failure 404 {object} services.BaseErrorResponse "404: record not found"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /courses/professors/{professorID}/reviews [get]
func (t *Controller) ListProfessorReviews(c *gin.Context) {
	profID, courseID, params, ok := t.professorAndCourse(c)
	if !ok {
		return
	}
	limit := params.Limit
	if limit == 0 || limit > 100 {
		limit = 20
	}
	opts := models.GetAllFactrakSurveysOptions{
		ProfessorID:             &profID,
		CourseCanonicalID:       courseID,
		Limit:                   &limit,
		Offset:                  &params.Offset,
		Preload:                 []string{"professor", "course"},
		PopulateAgreements:      true,
		PopulateClientAgreement: true,
		ClientAgreementUserID:   services.GetUserID(c),
		ProfAtWilliams:          true,
	}
	var surveys []*models.FactrakSurvey
	if err := t.surveys.GetAllSurveys(&surveys, &opts); err != nil {
		t.RespondError(c, err)
		return
	}
	factrak.RemoveUserIDFromSurveys(c, surveys)
	trimReviewProfessors(surveys)
	if err := t.attachCatalogCourses(surveys); err != nil {
		t.RespondError(c, err)
		return
	}
	t.RespondOK(c, surveys)
}
