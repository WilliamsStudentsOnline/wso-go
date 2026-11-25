package courses_test

import (
	"net/http"
	"testing"

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

func TestCourseRelated(t *testing.T) {
	assert := testify.New(t)
	env := utils.SetupTest(t, auth.ScopeAllOther)
	SetupRouter(env.Router, env.DB, env.Cfg, zaptest.NewLogger(t).Sugar())

	prof := models.User{Type: models.UserTypeProfessor, Name: "Jane Teach", UnixID: "jt1"}
	require.NoError(t, env.DB.Create(&prof).Error)
	course := func(crseID, subject string, number int, title, attrs, unixID string) catalog_update.RawCourse {
		r := catalog_update.RawCourse{
			AcademicYear: 2027, Offered: "Y", Semester: 1271, CourseID: crseID, Subject: subject,
			CatalogNumber: number, ClassSection: "01", ClassNumber: 2000 + number, SSRComponent: "LEC",
			CourseTitleLong: title, AttributesSearch: attrs,
		}
		if unixID != "" {
			r.FirstName1, r.LastName1, r.UnixID1 = "Jane", "Teach", unixID
		}
		return r
	}
	require.NoError(t, catalog_update.IngestCatalog(env.DB, []catalog_update.RawCourse{
		course("100136", "CSCI", 136, "Data Structures", "DIV_D3,CSCI_CORE", "jt1"),
		course("100134", "CSCI", 134, "Introduction to Computer Science", "CSCI_CORE", "jt1"),
		course("100256", "CSCI", 256, "Algorithm Design and Analysis", "CSCI_CORE", ""),
	}, zap.NewNop().Sugar()))

	code, body := send(t, env, http.MethodGet, "/catalog/100136/related", nil)
	require.Equal(t, http.StatusOK, code, string(body))
	var resp CourseRelatedResponse
	data(t, assert, body, &resp)
	assert.Equal(1271, resp.Strm)

	require.Len(t, resp.Professors, 1)
	assert.Equal("Jane Teach", resp.Professors[0].Instructor.Name)
	require.Len(t, resp.Professors[0].Courses, 1)
	assert.Equal("100134", resp.Professors[0].Courses[0].CrseID)
	assert.Equal(1271, resp.Professors[0].Courses[0].LastTaught.Strm)

	require.Len(t, resp.Attributes, 1)
	assert.Equal("CSCI_CORE", resp.Attributes[0].Code)
	require.Len(t, resp.Attributes[0].Courses, 2)
	assert.Equal("CSCI 134", resp.Attributes[0].Courses[0].Listings[0].Code)
	assert.Equal("CSCI 256", resp.Attributes[0].Courses[1].Listings[0].Code)
	assert.Equal(1271, resp.Attributes[0].Courses[1].Strm)

	code, body = send(t, env, http.MethodGet, "/catalog/100136/related?strm=1273", nil)
	require.Equal(t, http.StatusOK, code, string(body))
	data(t, assert, body, &resp)
	assert.Zero(resp.Strm)
	assert.Empty(resp.Professors)
	assert.Empty(resp.Attributes)

	code, _ = send(t, env, http.MethodGet, "/catalog/999999/related", nil)
	assert.Equal(http.StatusNotFound, code)
}
