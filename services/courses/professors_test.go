package courses_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/jobs/catalog_update"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/courses"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// Jane teaches CSCI 136 this year with two reviews (one recommends); Gone left Williams
func seedProfessors(t *testing.T, scopes ...string) (*utils.TestEnv, models.User, models.User) {
	t.Helper()
	env := utils.SetupTest(t, scopes...)
	utils.AddUserContexts(env.Router, 999)
	SetupRouter(env.Router, env.DB, env.Cfg, zaptest.NewLogger(t).Sugar())

	yes := true
	dept := models.Department{Name: "Computer Science"}
	require.NoError(t, env.DB.Create(&dept).Error)
	phone := "413-555-0101"
	teach := models.User{Type: models.UserTypeProfessor, Name: "Jane Teach", UnixID: "jt1", AtWilliams: &yes, DepartmentID: &dept.ID, CellPhone: &phone}
	gone := models.User{Type: models.UserTypeProfessor, Name: "Gone Prof", UnixID: "gp1"}
	student := models.User{Type: models.UserTypeStudent, Name: "Stu Dent", UnixID: "sd1"}
	for _, u := range []*models.User{&teach, &gone, &student} {
		require.NoError(t, env.DB.Create(u).Error)
	}
	require.NoError(t, env.DB.Model(&gone).Update("at_williams", false).Error)

	year, _ := CurrentTerm(time.Now())
	strm := 1000 + (year%100)*10 + 1 // this academic year's fall
	require.NoError(t, catalog_update.IngestCatalog(env.DB, []catalog_update.RawCourse{{
		AcademicYear: year, Offered: "Y", Semester: strm, CourseID: "100136", Subject: "CSCI",
		CatalogNumber: 136, ClassSection: "01", ClassNumber: 2001, SSRComponent: "LEC",
		CourseTitleLong: "Data Structures", FirstName1: "Jane", LastName1: "Teach", UnixID1: "jt1",
	}}, zap.NewNop().Sugar()))
	var canonical models.CourseCanonical
	require.NoError(t, env.DB.Where("crse_id = ?", "100136").First(&canonical).Error)

	area := models.AreaOfStudy{Name: "Computer Science", Abbreviation: "CSCI", DepartmentID: &dept.ID}
	require.NoError(t, env.DB.Create(&area).Error)
	legacy := models.Course{Number: "136", AreaOfStudyID: &area.ID}
	require.NoError(t, env.DB.Create(&legacy).Error)
	lead, rec, notRec := 5, true, false
	for _, r := range []*bool{&rec, &notRec} {
		require.NoError(t, env.DB.Create(&models.FactrakSurvey{
			UserID: student.ID, ProfessorID: teach.ID, CourseID: legacy.ID, CourseCanonicalID: &canonical.ID,
			WouldRecommendCourse: r, LeadLecture: &lead, Comment: "review",
		}).Error)
	}
	return env, teach, gone
}

func TestSearchProfessors(t *testing.T) {
	env, _, _ := seedProfessors(t, auth.ScopeAllOther)
	assert := env.Assert
	w, err := utils.DoHTTPReq(env.Router, http.MethodGet, "/professors?q=prof", nil)
	assert.NoError(err)
	var profs []ProfessorSummaryResponse
	data(t, assert, w.Body.Bytes(), &profs)
	require.Len(t, profs, 1)
	assert.Equal("Gone Prof", profs[0].Name)
	assert.False(profs[0].AtWilliams) // departed professors stay searchable, flagged

	w, err = utils.DoHTTPReq(env.Router, http.MethodGet, "/professors", nil)
	assert.NoError(err)
	data(t, assert, w.Body.Bytes(), &profs)
	require.Len(t, profs, 2)
	assert.Equal("Jane Teach", profs[0].Name) // current professors first
	assert.True(profs[0].TeachingThisYear)
}

func TestGetProfessor(t *testing.T) {
	env, teach, _ := seedProfessors(t, auth.ScopeAllOther)
	assert := env.Assert
	w, err := utils.DoHTTPReq(env.Router, http.MethodGet, fmt.Sprintf("/professors/%d", teach.ID), nil)
	assert.NoError(err)
	require.Equal(t, http.StatusOK, w.Code)
	var prof ProfessorResponse
	data(t, assert, w.Body.Bytes(), &prof)
	assert.Equal("Computer Science", prof.Department)
	assert.True(prof.AtWilliams)
	assert.True(prof.TeachingThisYear)
	require.Len(t, prof.Courses, 1)
	c := prof.Courses[0]
	assert.Equal("100136", c.CrseID)
	assert.Equal(1, c.TermsTaught)
	assert.Equal(2, c.Reviews)
	require.NotNil(t, c.LastTaught)
	assert.Contains(c.LastTaught.Label, "Fall")
	require.Len(t, c.Terms, 1)
	assert.Equal(c.LastTaught.Strm, c.Terms[0].Strm)

	w, err = utils.DoHTTPReq(env.Router, http.MethodGet, "/professors/999999", nil)
	assert.NoError(err)
	assert.Equal(http.StatusNotFound, w.Code)
}

func TestProfessorRatingsAndReviews(t *testing.T) {
	env, teach, _ := seedProfessors(t, auth.ScopeAllOther, auth.ScopeFactrakFull)
	assert := env.Assert
	var ratings models.FactrakSurveyAvgRatings
	for _, q := range []string{"", "?crseID=100136"} {
		w, err := utils.DoHTTPReq(env.Router, http.MethodGet, fmt.Sprintf("/professors/%d/ratings%s", teach.ID, q), nil)
		assert.NoError(err)
		require.Equal(t, http.StatusOK, w.Code, q)
		data(t, assert, w.Body.Bytes(), &ratings)
		assert.Equal(2, ratings.NumWouldRecommendCourse, q)
		assert.InDelta(0.5, ratings.AvgWouldRecommendCourse, 0.001, q)
		assert.InDelta(5, ratings.AvgLeadLecture, 0.001, q)
	}

	w, err := utils.DoHTTPReq(env.Router, http.MethodGet, fmt.Sprintf("/professors/%d/reviews?crseID=100136", teach.ID), nil)
	assert.NoError(err)
	var reviews []models.FactrakSurvey
	data(t, assert, w.Body.Bytes(), &reviews)
	assert.Len(reviews, 2)
	require.NotNil(t, reviews[0].Professor)
	assert.Nil(reviews[0].Professor.CellPhone) // private user fields aren't exposed
	require.NotNil(t, reviews[0].CourseCanonical)
	assert.Equal("100136", reviews[0].CourseCanonical.CrseID)

	w, err = utils.DoHTTPReq(env.Router, http.MethodGet, fmt.Sprintf("/professors/%d/reviews?crseID=nope", teach.ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusNotFound, w.Code)
}

func TestProfessorRatingsNeedFactrakAccess(t *testing.T) {
	env, teach, _ := seedProfessors(t, auth.ScopeAllOther)
	for _, path := range []string{"ratings", "reviews", "stats"} {
		w, err := utils.DoHTTPReq(env.Router, http.MethodGet, fmt.Sprintf("/professors/%d/%s", teach.ID, path), nil)
		env.Assert.NoError(err)
		env.Assert.Equal(http.StatusForbidden, w.Code, path)
	}
}

func TestCatalogInstructorsCountAsProfessors(t *testing.T) {
	env, _, _ := seedProfessors(t, auth.ScopeAllOther)
	assert := env.Assert
	lecturer := models.User{Type: models.UserTypeStaff, Name: "Lee Lecturer", UnixID: "ll1"}
	clerk := models.User{Type: models.UserTypeStaff, Name: "Lee Clerk", UnixID: "lc1"}
	for _, u := range []*models.User{&lecturer, &clerk} {
		require.NoError(t, env.DB.Create(u).Error)
	}
	var o models.CourseOffering
	require.NoError(t, env.DB.First(&o).Error)
	require.NoError(t, env.DB.Create(&models.CourseOfferingInstructor{CourseOfferingID: o.ID, UserID: &lecturer.ID,
		NameAsListed: lecturer.Name}).Error)

	w, err := utils.DoHTTPReq(env.Router, http.MethodGet, "/professors?q=lee", nil)
	assert.NoError(err)
	var profs []ProfessorSummaryResponse
	data(t, assert, w.Body.Bytes(), &profs)
	require.Len(t, profs, 1)
	assert.Equal("Lee Lecturer", profs[0].Name)

	w, err = utils.DoHTTPReq(env.Router, http.MethodGet, fmt.Sprintf("/professors/%d", lecturer.ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)
	w, err = utils.DoHTTPReq(env.Router, http.MethodGet, fmt.Sprintf("/professors/%d", clerk.ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusNotFound, w.Code)
}

func TestSearchProfessorsByArea(t *testing.T) {
	env, _, _ := seedProfessors(t, auth.ScopeAllOther)
	assert := env.Assert
	ann := models.User{Type: models.UserTypeProfessor, Name: "Ann Algo", UnixID: "aa1"}
	require.NoError(t, env.DB.Create(&ann).Error)
	year, _ := CurrentTerm(time.Now())
	row := func(y int, crseID, subject string, number, classNbr int, first, last, unix string) catalog_update.RawCourse {
		return catalog_update.RawCourse{
			AcademicYear: y, Offered: "Y", Semester: 1000 + (y%100)*10 + 1, CourseID: crseID, Subject: subject,
			CatalogNumber: number, ClassSection: "01", ClassNumber: classNbr, SSRComponent: "LEC",
			CourseTitleLong: "Course " + crseID, FirstName1: first, LastName1: last, UnixID1: unix,
		}
	}
	require.NoError(t, catalog_update.IngestCatalog(env.DB, []catalog_update.RawCourse{
		row(year-2, "100256", "CSCI", 256, 3001, "Ann", "Algo", "aa1"),
		row(year-3, "100134", "CSCI", 134, 3002, "Gone", "Prof", "gp1"),
		row(year-1, "200150", "MATH", 150, 3003, "Gone", "Prof", "gp1"),
	}, zap.NewNop().Sugar()))

	names := func(path string) []string {
		w, err := utils.DoHTTPReq(env.Router, http.MethodGet, path, nil)
		assert.NoError(err)
		var profs []ProfessorSummaryResponse
		data(t, assert, w.Body.Bytes(), &profs)
		out := []string{}
		for _, p := range profs {
			out = append(out, p.Name)
		}
		return out
	}
	assert.Equal([]string{"Ann Algo", "Jane Teach"}, names("/professors?area=csci"))
	assert.Equal([]string{"Jane Teach"}, names("/professors?area=CSCI&q=jane"))
	assert.Equal([]string{"Gone Prof"}, names("/professors?area=MATH"))
}
