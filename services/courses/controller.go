package courses

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/redis_util"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/WilliamsStudentsOnline/wso-go/services/factrak"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

var client *redis_util.RedisClient

func ConfigureClient(addr string) error {
	if addr == "" {
		addr = redis_util.DefaultRedisAddr
	}
	var err error
	client, err = redis_util.SetupClient(addr, "", redis_util.CoursesDatabaseID)
	return err
}

const termCatalogCacheTTL = 24 * time.Hour

type Controller struct {
	services.BaseController
	catalog *models.CourseCatalogModel
	surveys *models.FactrakSurveyModel
	now     func() time.Time
}

func NewController(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) *Controller {
	return &Controller{
		BaseController: services.BaseController{Log: log},
		catalog:        models.NewCourseCatalogModel(db, log),
		surveys:        models.NewFactrakSurveyModel(db, log),
		now:            time.Now,
	}
}

func parseStrm(c *gin.Context) (int, bool) {
	strm, err := strconv.Atoi(c.Param("strm"))
	if err != nil || strm <= 0 {
		return 0, false
	}
	return strm, true
}

// ListTerms godoc
// @Summary List catalog terms
// @Description Terms with catalog data, newest first. `current` marks the term in session (over the summer, the coming fall).
// @ID listCourseTerms
// @Tags courses
// @Produce json
// @Success 200 {array} courses.TermResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /courses/terms [get]
func (t *Controller) ListTerms(c *gin.Context) {
	terms, err := t.catalog.ListTerms()
	if err != nil {
		t.RespondError(c, err)
		return
	}
	now := t.now()
	out := make([]TermResponse, 0, len(terms))
	for _, term := range terms {
		r := termResponse(term.Year, term.Term, term.Strm, now)
		r.CourseCount = term.CourseCount
		out = append(out, r)
	}
	t.RespondOK(c, out)
}

// GetTermCatalog godoc
// @Summary Get a term's courses
// @Description Every catalog course in a term with listings, attributes and sections, for browsing and client-side filtering. Cached per term; supports If-None-Match.
// @ID getCourseTermCatalog
// @Tags courses
// @Produce json
// @Param strm path int true "Term code, e.g. 1271 (Fall 2026)"
// @Param include query string false "Set to description to include full course descriptions" Enums(description)
// @Success 200 {object} courses.TermCatalogResponse
// @Success 304 "Not modified"
// @Failure 400 {object} services.BaseErrorResponse "1100: could not parse malformed request data"
// @Failure 404 {object} services.BaseErrorResponse "404: record not found"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /courses/terms/{strm} [get]
func (t *Controller) GetTermCatalog(c *gin.Context) {
	strm, ok := parseStrm(c)
	if !ok {
		t.RespondAPIError(c, lib.NewErrorMalformedRequestData(errBadStrm))
		return
	}
	withDescription := c.Query("include") == "description"

	version, err := t.catalog.TermVersion(strm)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	key := redis_util.GetCourseTermCatalogKey(strm, withDescription, version.Unix())

	payload, err := t.cachedPayload(c, key)
	if err != nil {
		payload, err = t.buildTermCatalog(strm, withDescription)
		if err != nil {
			t.RespondError(c, err)
			return
		}
		if setErr := client.SetWithExpiration(c, key, payload, termCatalogCacheTTL); setErr != nil && setErr != redis_util.ErrRedisClientNotConfigured {
			t.Log.Warnf("cache term catalog %d: %v", strm, setErr)
		}
	}

	sum := sha256.Sum256(payload)
	etag := `"` + hex.EncodeToString(sum[:8]) + `"`
	c.Header("ETag", etag)
	c.Header("Cache-Control", "private, no-cache")
	if c.GetHeader("If-None-Match") == etag {
		c.Status(http.StatusNotModified)
		return
	}
	t.RespondOK(c, json.RawMessage(payload))
}

func (t *Controller) cachedPayload(c *gin.Context, key string) ([]byte, error) {
	v, err := client.Get(c, key)
	if err != nil {
		return nil, err
	}
	s, ok := v.(string)
	if !ok {
		return nil, errBadCache
	}
	return []byte(s), nil
}

func (t *Controller) buildTermCatalog(strm int, withDescription bool) ([]byte, error) {
	courses, err := t.catalog.TermCourses(strm)
	if err != nil {
		return nil, err
	}
	if len(courses) == 0 {
		return nil, lib.ErrorRecordNotFound
	}
	resp := TermCatalogResponse{
		Term:    termResponse(courses[0].Year, courses[0].Term, strm, t.now()),
		Courses: make([]TermCourseResponse, 0, len(courses)),
	}
	resp.Term.CourseCount = len(courses)
	for _, cis := range courses {
		resp.Courses = append(resp.Courses, termCourseResponse(cis, withDescription))
	}
	return json.Marshal(resp)
}

// GetCourseInSemester godoc
// @Summary Get a course in a term
// @Description One course in one term with every catalog detail and its sections. Also works for recovered terms before Fall 2022.
// @ID getCourseInSemester
// @Tags courses
// @Produce json
// @Param strm path int true "Term code, e.g. 1271 (Fall 2026)"
// @Param crseID path string true "Catalog course ID"
// @Success 200 {object} courses.CourseInSemesterResponse
// @Failure 400 {object} services.BaseErrorResponse "1100: could not parse malformed request data"
// @Failure 404 {object} services.BaseErrorResponse "404: record not found"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /courses/terms/{strm}/{crseID} [get]
func (t *Controller) GetCourseInSemester(c *gin.Context) {
	strm, ok := parseStrm(c)
	if !ok {
		t.RespondAPIError(c, lib.NewErrorMalformedRequestData(errBadStrm))
		return
	}
	cis, err := t.catalog.GetCourseInSemester(strm, c.Param("crseID"))
	if err != nil {
		t.RespondError(c, err)
		return
	}
	t.RespondOK(c, courseInSemesterResponse(cis, t.now()))
}

type SearchCoursesParams struct {
	Q            string   `form:"q"`
	Subject      []string `form:"subject"`
	Level        []int    `form:"level" binding:"dive,min=0,max=4"`
	Division     []string `form:"division" binding:"dive,oneof=div1 div2 div3"`
	Distribution []string `form:"distribution" binding:"dive,oneof=dpe qfr wac"`
	// The API client (axios) sends arrays as level[]=1&level[]=2
	SubjectBrackets      []string `form:"subject[]" swaggerignore:"true"`
	LevelBrackets        []int    `form:"level[]" binding:"dive,min=0,max=4" swaggerignore:"true"`
	DivisionBrackets     []string `form:"division[]" binding:"dive,oneof=div1 div2 div3" swaggerignore:"true"`
	DistributionBrackets []string `form:"distribution[]" binding:"dive,oneof=dpe qfr wac" swaggerignore:"true"`
	Offered              string   `form:"offered" binding:"omitempty,oneof=last-year 3-years"`
	Limit                int      `form:"limit"`
	Offset               int      `form:"offset"`
}

func offeredSince(offered string, now time.Time) int {
	year, _ := CurrentTerm(now)
	switch offered {
	case "last-year":
		return year - 1
	case "3-years":
		return year - 3
	default:
		return 0
	}
}

// SearchCourses godoc
// @Summary Search the all-time catalog
// @Description Search canonical courses across every term we have. Text matches like the per-semester search: every word must appear in the title, a code, an instructor's name or the description ("csci136" also works). Filters match the per-semester ones: any option within a group, every group must match; divisions and distributions count if the course had them in any term. Each result lists the terms it was offered.
// @ID searchCourses
// @Tags courses
// @Produce json
// @Param q query string false "Words to match in title, code, instructor or description"
// @Param subject query []string false "Only courses with a listing in one of these subjects, e.g. CSCI" collectionFormat(multi)
// @Param level query []int false "Listing level: 0 (winter study), 1 (100-level) to 4 (400-level)" collectionFormat(multi)
// @Param division query []string false "Divisions" Enums(div1,div2,div3) collectionFormat(multi)
// @Param distribution query []string false "Distribution requirements" Enums(dpe,qfr,wac) collectionFormat(multi)
// @Param offered query string false "Recency: offered this or last academic year, or within the past 3" Enums(last-year,3-years)
// @Param limit query int false "Default 25, max 100"
// @Param offset query int false "Offset pagination"
// @Success 200 {array} courses.CourseSearchResult
// @Failure 400 {object} services.BaseErrorResponse "1100: could not parse malformed request data; 1101: request data validation failed"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /courses/search [get]
func (t *Controller) SearchCourses(c *gin.Context) {
	var params SearchCoursesParams
	if err := c.ShouldBindQuery(&params); err != nil {
		t.RespondBadBind(c, err)
		return
	}
	results, total, err := t.catalog.SearchCanonical(models.CatalogSearchOptions{
		Query: params.Q, Subjects: append(params.Subject, params.SubjectBrackets...), Levels: append(params.Level, params.LevelBrackets...),
		Divisions:     append(params.Division, params.DivisionBrackets...),
		Distributions: append(params.Distribution, params.DistributionBrackets...),
		OfferedSince:  offeredSince(params.Offered, t.now()),
		Limit:         params.Limit, Offset: params.Offset,
	})
	if err != nil {
		t.RespondError(c, err)
		return
	}
	out := make([]CourseSearchResult, 0, len(results))
	for _, r := range results {
		out = append(out, searchResult(r))
	}
	t.SetPaginationTotal(c, total)
	t.RespondOK(c, out)
}

// GetCourse godoc
// @Summary Get a course
// @Description A canonical course across all terms: listings with the years each was used, offering history (newest first) and every professor who taught it.
// @ID getCatalogCourse
// @Tags courses
// @Produce json
// @Param crseID path string true "Catalog course ID"
// @Success 200 {object} courses.CourseResponse
// @Failure 404 {object} services.BaseErrorResponse "404: record not found"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /courses/catalog/{crseID} [get]
func (t *Controller) GetCourse(c *gin.Context) {
	course, err := t.catalog.GetCanonical(c.Param("crseID"))
	if err != nil {
		t.RespondError(c, err)
		return
	}
	t.RespondOK(c, courseResponse(course))
}

// ListAttributes godoc
// @Summary List course attributes
// @Description Readable names for catalog attribute codes (distributions and department/program groups like AFR_AFRCORE).
// @ID listCourseAttributes
// @Tags courses
// @Produce json
// @Success 200 {array} courses.AttributeResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /courses/attributes [get]
func (t *Controller) ListAttributes(c *gin.Context) {
	attrs, err := t.catalog.ListCourseAttributes()
	if err != nil {
		t.RespondError(c, err)
		return
	}
	out := make([]AttributeResponse, 0, len(attrs))
	for _, a := range attrs {
		r := AttributeResponse{Code: a.Code, Name: a.Name, Kind: a.Kind}
		if a.AreaOfStudy != nil {
			r.Area = a.AreaOfStudy.Abbreviation
		}
		out = append(out, r)
	}
	t.RespondOK(c, out)
}

// ListAreas godoc
// @Summary List areas of study
// @Description Every area of study (subject) with its catalog details and attribute groups. catalogURL is set when the area is in the current catalog.
// @ID listCourseAreas
// @Tags courses
// @Produce json
// @Success 200 {array} courses.AreaResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /courses/areas [get]
func (t *Controller) ListAreas(c *gin.Context) {
	areas, attrs, err := t.catalog.ListAreasOfStudy()
	if err != nil {
		t.RespondError(c, err)
		return
	}
	byArea := map[uint][]AttributeResponse{}
	for _, a := range attrs {
		byArea[*a.AreaOfStudyID] = append(byArea[*a.AreaOfStudyID], AttributeResponse{Code: a.Code, Name: a.Name, Kind: a.Kind})
	}
	out := make([]AreaResponse, 0, len(areas))
	for _, a := range areas {
		r := AreaResponse{
			ID: a.ID, Abbreviation: a.Abbreviation, Name: a.Name, CatalogSubjects: areaCatalogSubjects(a.Abbreviation),
			CatalogKind: a.CatalogKind, CatalogURL: a.CatalogURL, CurriculumPDFURL: a.CurriculumPDFURL,
			WebsiteURL: a.WebsiteURL, Description: a.Description, Attributes: byArea[a.ID],
		}
		if r.Attributes == nil {
			r.Attributes = []AttributeResponse{}
		}
		out = append(out, r)
	}
	t.RespondOK(c, out)
}

// GetTermFactrak godoc
// @Summary Get Factrak scores for a term
// @Description "Would recommend" counts for each course in a term that has linked Factrak reviews, overall and for the term's own instructors. Uses the reviews Factrak shows: from the last 5 years, for professors still at Williams. Reviews whose catalog link is a mismatch (the professor never taught this course in the catalog) are left out unless includeMismatch=true. Needs full Factrak access.
// @ID getCourseTermFactrak
// @Tags courses
// @Produce json
// @Param strm path int true "Term code, e.g. 1271 (Fall 2026)"
// @Param includeMismatch query bool false "Also count reviews with a mismatched catalog link"
// @Success 200 {array} courses.CourseFactrakResponse
// @Failure 400 {object} services.BaseErrorResponse "1100: could not parse malformed request data"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /courses/terms/{strm}/factrak [get]
func (t *Controller) GetTermFactrak(c *gin.Context) {
	strm, ok := parseStrm(c)
	if !ok {
		t.RespondAPIError(c, lib.NewErrorMalformedRequestData(errBadStrm))
		return
	}
	includeMismatch := c.Query("includeMismatch") == "true"
	summaries, err := t.catalog.TermFactrakSummaries(strm, includeMismatch)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	out := make([]CourseFactrakResponse, 0, len(summaries))
	for _, s := range summaries {
		out = append(out, CourseFactrakResponse{
			CrseID: s.CrseID, Reviews: s.Reviews, Recommends: s.Recommends,
			CurrentReviews: s.CurrentReviews, CurrentRecommends: s.CurrentRecommends,
		})
	}
	t.RespondOK(c, out)
}

type ListCourseReviewsParams struct {
	ProfessorID *uint `form:"professorID"`
	Limit       uint  `form:"limit"`
	Offset      uint  `form:"offset"`
}

// ListCourseReviews godoc
// @Summary List a course's Factrak reviews
// @Description Factrak reviews linked to a catalog course, across all its cross-listings, newest first. Same rules as Factrak's own survey list: full Factrak access, reviews from the last 5 years, professors still at Williams, authors hidden.
// @ID listCourseReviews
// @Tags courses
// @Produce json
// @Param crseID path string true "Catalog course ID"
// @Param professorID query int false "Only reviews of this professor"
// @Param limit query int false "Default 20, max 100"
// @Param offset query int false "Offset pagination"
// @Success 200 {array} models.FactrakSurvey
// @Failure 400 {object} services.BaseErrorResponse "1100: could not parse malformed request data"
// @Failure 404 {object} services.BaseErrorResponse "404: record not found"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /courses/catalog/{crseID}/reviews [get]
func (t *Controller) ListCourseReviews(c *gin.Context) {
	var params ListCourseReviewsParams
	if err := c.ShouldBindQuery(&params); err != nil {
		t.RespondBadBind(c, err)
		return
	}
	courseID, err := t.catalog.CanonicalID(c.Param("crseID"))
	if err != nil {
		t.RespondError(c, err)
		return
	}
	limit := params.Limit
	if limit == 0 || limit > 100 {
		limit = 20
	}
	userID := services.GetUserID(c)
	opts := models.GetAllFactrakSurveysOptions{
		CourseCanonicalID:       &courseID,
		ProfessorID:             params.ProfessorID,
		Limit:                   &limit,
		Offset:                  &params.Offset,
		Preload:                 []string{"professor", "course"},
		PopulateAgreements:      true,
		PopulateClientAgreement: true,
		ClientAgreementUserID:   userID,
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

func (t *Controller) attachCatalogCourses(surveys []*models.FactrakSurvey) error {
	ids := []uint{}
	for _, s := range surveys {
		if s.CourseCanonicalID != nil {
			ids = append(ids, *s.CourseCanonicalID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	courses, err := t.catalog.CanonicalCoursesByID(ids)
	if err != nil {
		return err
	}
	for _, s := range surveys {
		if s.CourseCanonicalID == nil {
			continue
		}
		if c, ok := courses[*s.CourseCanonicalID]; ok {
			s.CourseCanonical = &models.CourseCanonical{BaseSchema: models.BaseSchema{ID: c.ID}, CrseID: c.CrseID, Title: c.Title}
		}
	}
	return nil
}

// The preloaded user has private fields (phone numbers, email, ...) that reviews must not expose
func trimReviewProfessors(surveys []*models.FactrakSurvey) {
	for _, s := range surveys {
		if p := s.Professor; p != nil {
			s.Professor = &models.User{
				BaseSchema: models.BaseSchema{ID: p.ID},
				Type:       p.Type,
				Name:       p.Name,
				UnixID:     p.UnixID,
				Title:      p.Title,
				AtWilliams: p.AtWilliams,
			}
		}
	}
}

type GetCourseRatingsParams struct {
	ProfessorID *uint `form:"professorID"`
}

// GetCourseRatings godoc
// @Summary Get a course's Factrak ratings
// @Description Averaged Factrak ratings for a catalog course across all its cross-listings, like Factrak's course ratings (reviews from the last 5 years), optionally for one professor. Needs full Factrak access.
// @ID getCatalogCourseRatings
// @Tags courses
// @Produce json
// @Param crseID path string true "Catalog course ID"
// @Param professorID query int false "Only reviews of this professor"
// @Success 200 {object} models.FactrakSurveyAvgRatings
// @Failure 400 {object} services.BaseErrorResponse "1100: could not parse malformed request data"
// @Failure 404 {object} services.BaseErrorResponse "404: record not found"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /courses/catalog/{crseID}/ratings [get]
func (t *Controller) GetCourseRatings(c *gin.Context) {
	var params GetCourseRatingsParams
	if err := c.ShouldBindQuery(&params); err != nil {
		t.RespondBadBind(c, err)
		return
	}
	courseID, err := t.catalog.CanonicalID(c.Param("crseID"))
	if err != nil {
		t.RespondError(c, err)
		return
	}
	var ratings models.FactrakSurveyAvgRatings
	if err := t.surveys.GetSurveyRatingsForCatalogCourse(courseID, params.ProfessorID, &ratings); err != nil {
		t.RespondError(c, err)
		return
	}
	t.RespondOK(c, ratings)
}
