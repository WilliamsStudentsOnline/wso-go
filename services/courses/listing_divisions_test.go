package courses_test

import (
	"net/http"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/jobs/catalog_update"
	. "github.com/WilliamsStudentsOnline/wso-go/services/courses"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// A cross-listed class counts by the code you register under: ENGL 215 is Div I, AAS 215 Div II
func TestDivisionsByListing(t *testing.T) {
	assert, env, _ := seedMine(t)
	row := func(subject string, classNbr int, attrs string) catalog_update.RawCourse {
		return catalog_update.RawCourse{
			AcademicYear: 2027, Offered: "Y", Semester: 1271, CourseID: "022643", Subject: subject,
			CatalogNumber: 215, ClassSection: "01", ClassNumber: classNbr, SSRComponent: "SEM",
			CourseTitleLong: "Introduction to Asian American Literature", AttributesSearch: attrs,
			StandardMeeting1: "TR", StartTime1: "11:20", EndTime1: "12:35",
		}
	}
	require.NoError(t, catalog_update.IngestCatalog(env.DB, []catalog_update.RawCourse{
		row("ENGL", 1796, "DIV_D1,DPE_DPE"),
		row("AAS", 1797, "AAS_AASCORE,DIV_D2,DPE_DPE"),
	}, zap.NewNop().Sugar()))

	code, body := send(t, env, http.MethodGet, "/terms/1271/022643", nil)
	require.Equal(t, http.StatusOK, code, string(body))
	var course struct {
		TermCourseResponse
	}
	data(t, assert, body, &course)
	assert.True(course.Div1)
	assert.True(course.Div2)
	require.Len(t, course.ByCode, 2)
	byCode := map[string]CodeDistributions{}
	for _, c := range course.ByCode {
		byCode[c.Code] = c
	}
	assert.True(byCode["ENGL 215"].Div1)
	assert.False(byCode["ENGL 215"].Div2)
	assert.True(byCode["AAS 215"].Div2)

	// Registered as AAS 215: Div II only
	code, body = send(t, env, http.MethodPost, "/mine", AddMyCourseRequest{CrseID: "022643", Strm: 1271, ClassNbrs: []int{1797}})
	require.Equal(t, http.StatusCreated, code, string(body))
	var mc MyCourseResponse
	data(t, assert, body, &mc)
	assert.False(mc.Div1)
	assert.True(mc.Div2)
	code, body = send(t, env, http.MethodGet, "/mine/distributions", nil)
	require.Equal(t, http.StatusOK, code)
	var d DistributionsResponse
	data(t, assert, body, &d)
	assert.Equal(0, d.Earned.Div1)
	assert.Equal(1, d.Earned.Div2)

	code, body = send(t, env, http.MethodGet, "/terms/1271/100136", nil)
	require.Equal(t, http.StatusOK, code)
	var csci TermCourseResponse
	data(t, assert, body, &csci)
	assert.Empty(csci.ByCode)
}
