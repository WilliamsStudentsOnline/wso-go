package courses_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/jobs/catalog_update"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/courses"
	testify "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// CSCI 136 (cross-listed as STAT 136) in Fall 2026 and Fall 2024, and a recovered Fall 2019 AFR 105
func seed(t *testing.T) (*testify.Assertions, *utils.TestEnv) {
	t.Helper()
	assert := testify.New(t)
	env := utils.SetupTest(t, auth.ScopeAllOther)
	SetupRouter(env.Router, env.DB, env.Cfg, zaptest.NewLogger(t).Sugar())
	ingestSample(t, env)
	return assert, env
}

func ingestSample(t *testing.T, env *utils.TestEnv) {
	t.Helper()
	csci := func(year, strm, classNbr int, section, component, subject string) catalog_update.RawCourse {
		return catalog_update.RawCourse{
			AcademicYear: year, Offered: "Y", Semester: strm, CourseID: "100136",
			Subject: subject, CatalogNumber: 136, ClassSection: section, ClassNumber: classNbr,
			SSRComponent: component, CourseTitleLong: "Data Structures", DescriptionSearch: "Lists and trees.",
			ShortDescription: "Lists.", AttributesSearch: "DIV_D3,QFR_QFR", GradingBasis: "OPX",
			EnrollmentLimit: "30", FirstName1: "Bill", LastName1: "Jannen",
			StandardMeeting1: "MWF", StartTime1: "10:00", EndTime1: "10:50", Facility1: "TCL 123",
			PreReqs: "CSCI 134",
		}
	}
	require.NoError(t, catalog_update.IngestCatalog(env.DB, []catalog_update.RawCourse{
		csci(2025, 1251, 1001, "01", "LEC", "CSCI"),
	}, zap.NewNop().Sugar()))
	require.NoError(t, catalog_update.IngestCatalog(env.DB, []catalog_update.RawCourse{
		csci(2027, 1271, 2001, "01", "LEC", "CSCI"),
		csci(2027, 1271, 2002, "L1", "LAB", "STAT"),
		{
			AcademicYear: 2027, Offered: "N", Semester: 1201, CourseID: "200105", Subject: "AFR",
			CatalogNumber: 105, ClassNumber: 1089, CourseTitleLong: "Arts of Africa",
			FirstName1: "Michelle", LastName1: "Apotsos",
		},
	}, zap.NewNop().Sugar()))
}

func data(t *testing.T, assert *testify.Assertions, w []byte, out interface{}) {
	t.Helper()
	resp := utils.GetHTTPDataResp(assert, w)
	require.Nil(t, resp.Error)
	require.NoError(t, json.Unmarshal(resp.Data, out))
}

func TestListTerms(t *testing.T) {
	assert, env := seed(t)
	w, err := utils.DoHTTPReq(env.Router, http.MethodGet, "/terms", nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	var terms []TermResponse
	data(t, assert, w.Body.Bytes(), &terms)
	// Newest first; the recovered Fall 2019 term isn't browsable
	require.Len(t, terms, 2)
	assert.Equal(1271, terms[0].Strm)
	assert.Equal("Fall 2026", terms[0].Label)
	assert.Equal(1, terms[0].CourseCount)
	assert.Equal("Fall 2024", terms[1].Label)
}

func TestGetTermCatalog(t *testing.T) {
	assert, env := seed(t)
	w, err := utils.DoHTTPReq(env.Router, http.MethodGet, "/terms/1271", nil)
	assert.NoError(err)
	require.Equal(t, http.StatusOK, w.Code)

	var catalog TermCatalogResponse
	data(t, assert, w.Body.Bytes(), &catalog)
	assert.Equal("Fall 2026", catalog.Term.Label)
	require.Len(t, catalog.Courses, 1)
	course := catalog.Courses[0]
	assert.Equal("100136", course.CrseID)
	assert.Equal("Lists.", course.ShortDescription)
	assert.Equal("", course.Description)
	codes := []string{}
	for _, l := range course.Listings {
		codes = append(codes, l.Code)
	}
	assert.Equal([]string{"CSCI 136", "STAT 136"}, codes)
	assert.True(course.Div3)
	assert.True(course.QFR)
	assert.False(course.PassFail) // OPX: pass/fail unavailable, fifth course available
	assert.True(course.FifthCourse)
	assert.Equal([]string{"DIV_D3", "QFR_QFR"}, course.Attributes)
	require.Len(t, course.Sections, 2)
	assert.Equal("01", course.Sections[0].Section)
	// The lab is registered as STAT
	assert.Equal("CSCI 136", course.Sections[0].Code)
	assert.Equal("STAT 136", course.Sections[1].Code)
	assert.Equal("MWF", course.Sections[0].Meetings[0].Days)
	assert.Equal("Bill Jannen", course.Sections[0].Instructors[0].Name)
	require.NotNil(t, course.Sections[0].EnrollmentLimit)
	assert.Equal(30, *course.Sections[0].EnrollmentLimit)

	etag := w.Header().Get("ETag")
	assert.NotEmpty(etag)
	req := httptest.NewRequest(http.MethodGet, "/terms/1271", nil)
	req.Header.Set("If-None-Match", etag)
	w2 := httptest.NewRecorder()
	env.Router.ServeHTTP(w2, req)
	assert.Equal(http.StatusNotModified, w2.Code)

	w, err = utils.DoHTTPReq(env.Router, http.MethodGet, "/terms/1271?include=description", nil)
	assert.NoError(err)
	data(t, assert, w.Body.Bytes(), &catalog)
	assert.Equal("Lists and trees.", catalog.Courses[0].Description)
}

func TestGetTermCatalogErrors(t *testing.T) {
	assert, env := seed(t)
	w, err := utils.DoHTTPReq(env.Router, http.MethodGet, "/terms/nope", nil)
	assert.NoError(err)
	assert.Equal(http.StatusBadRequest, w.Code)

	w, err = utils.DoHTTPReq(env.Router, http.MethodGet, "/terms/1999", nil)
	assert.NoError(err)
	assert.Equal(http.StatusNotFound, w.Code)
}

func TestGetCourseInSemester(t *testing.T) {
	assert, env := seed(t)
	w, err := utils.DoHTTPReq(env.Router, http.MethodGet, "/terms/1271/100136", nil)
	assert.NoError(err)
	require.Equal(t, http.StatusOK, w.Code)

	var cis CourseInSemesterResponse
	data(t, assert, w.Body.Bytes(), &cis)
	assert.Equal("Data Structures", cis.Title)
	assert.Equal("Lists and trees.", cis.Description)
	assert.Equal("CSCI 134", cis.Prereqs)
	assert.Equal("Fall 2026", cis.Term.Label)

	w, err = utils.DoHTTPReq(env.Router, http.MethodGet, "/terms/1201/200105", nil)
	assert.NoError(err)
	require.Equal(t, http.StatusOK, w.Code)
	data(t, assert, w.Body.Bytes(), &cis)
	assert.Equal("last_offered", cis.Source)
	assert.Equal("Fall 2019", cis.Term.Label)

	w, err = utils.DoHTTPReq(env.Router, http.MethodGet, "/terms/1271/999999", nil)
	assert.NoError(err)
	assert.Equal(http.StatusNotFound, w.Code)
}

func TestGetCourse(t *testing.T) {
	assert, env := seed(t)
	w, err := utils.DoHTTPReq(env.Router, http.MethodGet, "/catalog/100136", nil)
	assert.NoError(err)
	require.Equal(t, http.StatusOK, w.Code)

	var course CourseResponse
	data(t, assert, w.Body.Bytes(), &course)
	assert.Equal("Data Structures", course.Title)
	require.Len(t, course.History, 2)
	assert.Equal("Fall 2026", course.History[0].Label)
	assert.Equal("Fall 2024", course.History[1].Label)
	require.Len(t, course.Professors, 1)
	assert.Equal("Bill Jannen", course.Professors[0].Name)
}

func TestSearchCourses(t *testing.T) {
	assert, env := seed(t)
	for _, q := range []string{"/search?q=data", "/search?q=csci%20136", "/search?q=stat13", "/search?subject=STAT"} {
		w, err := utils.DoHTTPReq(env.Router, http.MethodGet, q, nil)
		assert.NoError(err)
		require.Equal(t, http.StatusOK, w.Code, q)
		var results []CourseSearchResult
		data(t, assert, w.Body.Bytes(), &results)
		if assert.Len(results, 1, q) {
			assert.Equal("100136", results[0].CrseID)
			assert.Len(results[0].Terms, 2)
			assert.Equal("Fall 2026", results[0].LastOffered.Label)
		}
	}

	w, err := utils.DoHTTPReq(env.Router, http.MethodGet, "/search?q=africa", nil)
	assert.NoError(err)
	var results []CourseSearchResult
	data(t, assert, w.Body.Bytes(), &results)
	if assert.Len(results, 1) {
		assert.Equal("Fall 2019", results[0].Terms[0].Label)
	}

	for _, q := range []string{"/search?subject=AFR&subject=STAT", "/search?subject[]=AFR&subject[]=STAT"} {
		w, err = utils.DoHTTPReq(env.Router, http.MethodGet, q, nil)
		assert.NoError(err)
		require.Equal(t, http.StatusOK, w.Code, q)
		results = nil
		data(t, assert, w.Body.Bytes(), &results)
		assert.Len(results, 2, q)
	}
}

func TestListAreasAndAttributes(t *testing.T) {
	assert, env := seed(t)
	dept := models.Department{Name: "Computer Science"}
	require.NoError(t, env.DB.Create(&dept).Error)
	csci := models.AreaOfStudy{Name: "Computer Science", Abbreviation: "CSCI", DepartmentID: &dept.ID}
	wgst := models.AreaOfStudy{Name: "Women's and Gender Studies", Abbreviation: "WGST", DepartmentID: &dept.ID}
	require.NoError(t, env.DB.Create(&csci).Error)
	require.NoError(t, env.DB.Create(&wgst).Error)
	units := []catalog_update.CatalogUnit{{
		Code: "csci", Name: "Computer Science", Kind: "major", CatalogURL: "https://catalog.williams.edu/csci/",
		WebsiteURL: "http://csci.williams.edu/", Attributes: []catalog_update.CatalogAttribute{{Code: "CSCI_CSCISYS", Name: "CSCI Systems"}},
	}}
	_, err := catalog_update.UpdateAreasAndAttributes(env.DB, units, zap.NewNop().Sugar())
	require.NoError(t, err)

	w, err := utils.DoHTTPReq(env.Router, http.MethodGet, "/areas", nil)
	assert.NoError(err)
	require.Equal(t, http.StatusOK, w.Code)
	var areas []AreaResponse
	data(t, assert, w.Body.Bytes(), &areas)
	byCode := map[string]AreaResponse{}
	for _, a := range areas {
		byCode[a.Abbreviation] = a
	}
	assert.Equal("major", byCode["CSCI"].CatalogKind)
	assert.Equal("http://csci.williams.edu/", byCode["CSCI"].WebsiteURL)
	require.Len(t, byCode["CSCI"].Attributes, 1)
	assert.Equal("CSCI Systems", byCode["CSCI"].Attributes[0].Name)
	assert.Equal([]string{"WGST", "WGSS"}, byCode["WGST"].CatalogSubjects)
	assert.Equal("", byCode["WGST"].CatalogURL) // not in the catalog under this code yet

	w, err = utils.DoHTTPReq(env.Router, http.MethodGet, "/attributes", nil)
	assert.NoError(err)
	var attrs []AttributeResponse
	data(t, assert, w.Body.Bytes(), &attrs)
	names := map[string]AttributeResponse{}
	for _, a := range attrs {
		names[a.Code] = a
	}
	assert.Equal("CSCI Systems", names["CSCI_CSCISYS"].Name)
	assert.Equal("CSCI", names["CSCI_CSCISYS"].Area)
	assert.Equal("Quantitative/Formal Reasoning", names["QFR_QFR"].Name)
	assert.Equal("distribution", names["QFR_QFR"].Kind)
}

func TestGetTermFactrak(t *testing.T) {
	assert := testify.New(t)
	env := utils.SetupTest(t, auth.ScopeAllOther, auth.ScopeFactrakFull)
	SetupRouter(env.Router, env.DB, env.Cfg, zaptest.NewLogger(t).Sugar())

	here := true
	current := models.User{Type: models.UserTypeProfessor, Name: "Bill Jannen", UnixID: "bj1", AtWilliams: &here}
	former := models.User{Type: models.UserTypeProfessor, Name: "Old Prof", UnixID: "op1", AtWilliams: &here}
	departed := models.User{Type: models.UserTypeProfessor, Name: "Gone Prof", UnixID: "gp1"}
	student := models.User{Type: models.UserTypeStudent, Name: "Stu Dent", UnixID: "sd1"}
	for _, u := range []*models.User{&current, &former, &departed, &student} {
		require.NoError(t, env.DB.Create(u).Error)
	}
	require.NoError(t, env.DB.Model(&departed).Update("at_williams", false).Error)
	require.NoError(t, catalog_update.IngestCatalog(env.DB, []catalog_update.RawCourse{{
		AcademicYear: 2027, Offered: "Y", Semester: 1271, CourseID: "100136", Subject: "CSCI",
		CatalogNumber: 136, ClassSection: "01", ClassNumber: 2001, SSRComponent: "LEC",
		CourseTitleLong: "Data Structures", FirstName1: "Bill", LastName1: "Jannen", UnixID1: "bj1",
	}}, zap.NewNop().Sugar()))
	var canonical models.CourseCanonical
	require.NoError(t, env.DB.Where("crse_id = ?", "100136").First(&canonical).Error)

	dept := models.Department{Name: "CS"}
	require.NoError(t, env.DB.Create(&dept).Error)
	area := models.AreaOfStudy{Name: "Computer Science", Abbreviation: "CSCI", DepartmentID: &dept.ID}
	require.NoError(t, env.DB.Create(&area).Error)
	legacy := models.Course{Number: "136", AreaOfStudyID: &area.ID}
	require.NoError(t, env.DB.Create(&legacy).Error)
	yes, no := true, false
	for _, s := range []struct {
		prof      uint
		recommend *bool
	}{{current.ID, &yes}, {current.ID, &no}, {former.ID, &yes}, {former.ID, nil}, {departed.ID, &yes}, {current.ID, &yes}} {
		require.NoError(t, env.DB.Create(&models.FactrakSurvey{
			UserID: student.ID, ProfessorID: s.prof, CourseID: legacy.ID,
			CourseCanonicalID: &canonical.ID, WouldRecommendCourse: s.recommend,
		}).Error)
	}
	// The departed professor's review and a 6-year-old review don't count
	var last models.FactrakSurvey
	require.NoError(t, env.DB.Order("id desc").First(&last).Error)
	require.NoError(t, env.DB.Model(&last).UpdateColumn("created_at", time.Now().AddDate(-6, 0, 0)).Error)

	w, err := utils.DoHTTPReq(env.Router, http.MethodGet, "/terms/1271/factrak", nil)
	assert.NoError(err)
	require.Equal(t, http.StatusOK, w.Code)
	var scores []CourseFactrakResponse
	data(t, assert, w.Body.Bytes(), &scores)
	require.Len(t, scores, 1)
	assert.Equal(CourseFactrakResponse{
		CrseID: "100136", Reviews: 3, Recommends: 2, CurrentReviews: 2, CurrentRecommends: 1,
	}, scores[0])

	require.NoError(t, env.DB.Model(&models.FactrakSurvey{}).Where("professor_id = ?", former.ID).
		Update("course_link_verification_source", models.CourseLinkVerificationSourceMismatch).Error)
	w, err = utils.DoHTTPReq(env.Router, http.MethodGet, "/terms/1271/factrak", nil)
	assert.NoError(err)
	data(t, assert, w.Body.Bytes(), &scores)
	assert.Equal(2, scores[0].Reviews)
	w, err = utils.DoHTTPReq(env.Router, http.MethodGet, "/terms/1271/factrak?includeMismatch=true", nil)
	assert.NoError(err)
	data(t, assert, w.Body.Bytes(), &scores)
	assert.Equal(3, scores[0].Reviews)

	// The fixed /factrak route doesn't shadow course lookups
	w, err = utils.DoHTTPReq(env.Router, http.MethodGet, "/terms/1271/100136", nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)
}

func TestGetTermFactrakNeedsFactrakAccess(t *testing.T) {
	assert, env := seed(t) // logged in, but no Factrak scope
	w, err := utils.DoHTTPReq(env.Router, http.MethodGet, "/terms/1271/factrak", nil)
	assert.NoError(err)
	assert.Equal(http.StatusForbidden, w.Code)
}

func TestSearchCoursesFilters(t *testing.T) {
	assert, env := seed(t)
	search := func(query string) []string {
		t.Helper()
		w, err := utils.DoHTTPReq(env.Router, http.MethodGet, "/search?"+query, nil)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, w.Code, query)
		var results []CourseSearchResult
		data(t, assert, w.Body.Bytes(), &results)
		ids := []string{}
		for _, r := range results {
			ids = append(ids, r.CrseID)
		}
		return ids
	}

	assert.Equal([]string{"100136"}, search("q=jannen"))
	assert.Equal([]string{"100136"}, search("q=trees"))
	assert.Equal([]string{"100136"}, search("q=stat%20136%20lists"))
	assert.Equal([]string{"200105"}, search("q=apotsos"))

	assert.ElementsMatch([]string{"100136", "200105"}, search("level=1"))
	assert.Empty(search("level=2"))
	assert.Equal([]string{"100136"}, search("division=div3"))
	assert.Equal([]string{"100136"}, search("division=div1&division=div3"))
	assert.Empty(search("division=div3&distribution=dpe"))
	assert.Equal([]string{"100136"}, search("distribution=qfr"))
	assert.Equal([]string{"100136"}, search("division%5B%5D=div1&division%5B%5D=div3"))
	assert.Empty(search("level%5B%5D=2"))

	// The recovered 2019 section is never recent; Fall 2026 is while it's within 3 years
	assert.NotContains(search("offered=3-years"), "200105")
	if year, _ := CurrentTerm(time.Now()); year-3 <= 2027 {
		assert.Contains(search("offered=last-year"), "100136")
	}

	w, err := utils.DoHTTPReq(env.Router, http.MethodGet, "/search?division=nope", nil)
	assert.NoError(err)
	assert.Equal(http.StatusBadRequest, w.Code)
	w, err = utils.DoHTTPReq(env.Router, http.MethodGet, "/search?offered=yesterday", nil)
	assert.NoError(err)
	assert.Equal(http.StatusBadRequest, w.Code)
}

func TestListCourseReviews(t *testing.T) {
	assert := testify.New(t)
	env := utils.SetupTest(t, auth.ScopeAllOther, auth.ScopeFactrakFull)
	utils.AddUserContexts(env.Router, 999)
	SetupRouter(env.Router, env.DB, env.Cfg, zaptest.NewLogger(t).Sugar())

	yes := true
	phone := "413-555-0100"
	here := models.User{Type: models.UserTypeProfessor, Name: "Here Prof", UnixID: "hp1", AtWilliams: &yes,
		CellPhone: &phone, WilliamsEmail: "hp1@williams.edu"}
	left := models.User{Type: models.UserTypeProfessor, Name: "Left Prof", UnixID: "lp1"}
	author := models.User{Type: models.UserTypeStudent, Name: "Stu Dent", UnixID: "sd1"}
	for _, u := range []*models.User{&here, &left, &author} {
		require.NoError(t, env.DB.Create(u).Error)
	}
	require.NoError(t, env.DB.Model(&left).Update("at_williams", false).Error)

	canonical := models.CourseCanonical{CrseID: "100136", Title: "Data Structures"}
	require.NoError(t, env.DB.Create(&canonical).Error)
	dept := models.Department{Name: "CS"}
	require.NoError(t, env.DB.Create(&dept).Error)
	csci := models.AreaOfStudy{Name: "Computer Science", Abbreviation: "CSCI", DepartmentID: &dept.ID}
	stat := models.AreaOfStudy{Name: "Statistics", Abbreviation: "STAT", DepartmentID: &dept.ID}
	require.NoError(t, env.DB.Create(&csci).Error)
	require.NoError(t, env.DB.Create(&stat).Error)
	asCSCI := models.Course{Number: "136", AreaOfStudyID: &csci.ID}
	asSTAT := models.Course{Number: "136", AreaOfStudyID: &stat.ID} // cross-listing filed separately
	require.NoError(t, env.DB.Create(&asCSCI).Error)
	require.NoError(t, env.DB.Create(&asSTAT).Error)

	survey := func(prof uint, course uint, comment string, age time.Duration) models.FactrakSurvey {
		s := models.FactrakSurvey{UserID: author.ID, ProfessorID: prof, CourseID: course,
			CourseCanonicalID: &canonical.ID, Comment: comment}
		require.NoError(t, env.DB.Create(&s).Error)
		require.NoError(t, env.DB.Model(&s).UpdateColumn("created_at", time.Now().Add(-age)).Error)
		return s
	}
	survey(here.ID, asCSCI.ID, "newer, filed as CSCI", time.Hour)
	survey(here.ID, asSTAT.ID, "older, filed as STAT", 48*time.Hour)
	survey(left.ID, asCSCI.ID, "professor left Williams", time.Hour)
	survey(here.ID, asCSCI.ID, "six years old", 6*365*24*time.Hour)

	get := func(query string) []models.FactrakSurvey {
		t.Helper()
		w, err := utils.DoHTTPReq(env.Router, http.MethodGet, "/catalog/100136/reviews"+query, nil)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, w.Code)
		var out []models.FactrakSurvey
		data(t, assert, w.Body.Bytes(), &out)
		return out
	}
	got := get("")
	require.Len(t, got, 2) // both cross-listings; not the departed professor or the 6-year-old review
	assert.Equal("newer, filed as CSCI", got[0].Comment)
	assert.Equal("older, filed as STAT", got[1].Comment)
	assert.Equal(uint(0), got[0].UserID) // author hidden
	require.NotNil(t, got[0].Professor)
	assert.Equal("Here Prof", got[0].Professor.Name)
	assert.Nil(got[0].Professor.CellPhone) // private user fields aren't exposed
	require.NotNil(t, got[0].CourseCanonical)
	assert.Equal("100136", got[0].CourseCanonical.CrseID)
	assert.Equal("", got[0].Professor.WilliamsEmail)

	assert.Len(get(fmt.Sprintf("?professorID=%d", left.ID)), 0)

	var ratings models.FactrakSurveyAvgRatings
	w, err := utils.DoHTTPReq(env.Router, http.MethodGet, "/catalog/100136/ratings", nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, w.Code)
	data(t, assert, w.Body.Bytes(), &ratings)
	assert.Equal(0, ratings.NumWouldRecommendCourse)
	w, err = utils.DoHTTPReq(env.Router, http.MethodGet, fmt.Sprintf("/catalog/100136/ratings?professorID=%d", here.ID), nil)
	require.NoError(t, err)
	assert.Equal(http.StatusOK, w.Code)
	assert.Len(get("?limit=1"), 1)

	w, err = utils.DoHTTPReq(env.Router, http.MethodGet, "/catalog/999999/reviews", nil)
	assert.NoError(err)
	assert.Equal(http.StatusNotFound, w.Code)
}

func TestListCourseReviewsNeedsFactrakAccess(t *testing.T) {
	assert, env := seed(t)
	w, err := utils.DoHTTPReq(env.Router, http.MethodGet, "/catalog/100136/reviews", nil)
	assert.NoError(err)
	assert.Equal(http.StatusForbidden, w.Code)
}
