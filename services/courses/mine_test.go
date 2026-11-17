package courses_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/courses"
	testify "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// CSCI 136 in Fall 2026 has lecture 2001 and lab 2002
func seedMine(t *testing.T) (*testify.Assertions, *utils.TestEnv, models.User) {
	t.Helper()
	assert := testify.New(t)
	env := utils.SetupTest(t, auth.ScopeAllOther, auth.ScopeWriteSelf)
	student := models.User{Type: models.UserTypeStudent, Name: "Stu Dent", UnixID: "sd1"}
	require.NoError(t, env.DB.Create(&student).Error)
	utils.AddUserContexts(env.Router, student.ID)
	SetupRouter(env.Router, env.DB, env.Cfg, zaptest.NewLogger(t).Sugar())
	ingestSample(t, env)
	return assert, env, student
}

func TestMyCoursesAddSectionsOneAtATime(t *testing.T) {
	assert, env, _ := seedMine(t)

	code, body := send(t, env, http.MethodPost, "/mine", AddMyCourseRequest{CrseID: "100136", Strm: 1271, ClassNbrs: []int{2001}})
	require.Equal(t, http.StatusCreated, code, string(body))
	var mc MyCourseResponse
	data(t, assert, body, &mc)
	assert.Equal("Data Structures", mc.Title)
	require.NotNil(t, mc.Term)
	assert.Equal("Fall 2026", mc.Term.Label)
	require.Len(t, mc.Sections, 1)
	assert.Equal("MWF", mc.Sections[0].Meetings[0].Days)
	assert.Equal([]string{"LAB", "LEC"}, mc.Components)
	assert.Equal([]string{"LAB"}, mc.MissingComponents)
	assert.True(mc.Div3)
	assert.True(mc.QFR)

	code, body = send(t, env, http.MethodPost, "/mine", AddMyCourseRequest{CrseID: "100136", Strm: 1271, ClassNbrs: []int{2002}})
	require.Equal(t, http.StatusCreated, code, string(body))
	var again MyCourseResponse
	data(t, assert, body, &again)
	assert.Equal(mc.ID, again.ID)
	assert.Len(again.Sections, 2)
	assert.Empty(again.MissingComponents)

	code, body = send(t, env, http.MethodGet, "/mine?strm=1271", nil)
	require.Equal(t, http.StatusOK, code)
	var list []MyCourseResponse
	data(t, assert, body, &list)
	require.Len(t, list, 1)

	code, body = send(t, env, http.MethodDelete, fmt.Sprintf("/mine/%d/sections/2002", mc.ID), nil)
	require.Equal(t, http.StatusOK, code, string(body))
	data(t, assert, body, &mc)
	assert.Equal([]string{"LAB"}, mc.MissingComponents)
}

func TestMyCoursesRejectsSectionsFromOtherTerms(t *testing.T) {
	assert, env, _ := seedMine(t)
	// 1001 is the Fall 2024 lecture, not Fall 2026
	code, body := send(t, env, http.MethodPost, "/mine", AddMyCourseRequest{CrseID: "100136", Strm: 1271, ClassNbrs: []int{1001}})
	assert.Equal(http.StatusBadRequest, code)
	assert.Contains(string(body), fmt.Sprint(lib.ErrorMyCourseSection.Code))

	code, _ = send(t, env, http.MethodPost, "/mine", AddMyCourseRequest{CrseID: "100136", ClassNbrs: []int{2001}})
	assert.Equal(http.StatusBadRequest, code) // sections need a term
}

func TestMyCoursesSomedayHideAndList(t *testing.T) {
	assert, env, _ := seedMine(t)
	hidden := true
	code, body := send(t, env, http.MethodPost, "/mine", AddMyCourseRequest{CrseID: "100136", Hidden: &hidden})
	require.Equal(t, http.StatusCreated, code, string(body))
	var someday MyCourseResponse
	data(t, assert, body, &someday)
	assert.Nil(someday.Term)
	assert.True(someday.Hidden)
	assert.True(someday.Div3) // a someday course shows its newest term's chips

	code, body = send(t, env, http.MethodPost, "/mine", AddMyCourseRequest{CrseID: "100136"})
	require.Equal(t, http.StatusCreated, code)
	var again MyCourseResponse
	data(t, assert, body, &again)
	assert.Equal(someday.ID, again.ID)

	code, _ = send(t, env, http.MethodPost, "/mine", AddMyCourseRequest{CrseID: "100136", Strm: 1251})
	require.Equal(t, http.StatusCreated, code)

	show := false
	code, body = send(t, env, http.MethodPatch, fmt.Sprintf("/mine/%d", someday.ID), UpdateMyCourseRequest{Hidden: &show})
	require.Equal(t, http.StatusOK, code, string(body))
	data(t, assert, body, &someday)
	assert.False(someday.Hidden)

	code, body = send(t, env, http.MethodGet, "/mine", nil)
	require.Equal(t, http.StatusOK, code)
	var list []MyCourseResponse
	data(t, assert, body, &list)
	require.Len(t, list, 2)
	require.NotNil(t, list[0].Term)
	assert.Equal("Fall 2024", list[0].Term.Label)
	assert.Nil(list[1].Term)
}

func TestMyCoursesDeleteBlockedByReview(t *testing.T) {
	assert, env, student := seedMine(t)
	code, body := send(t, env, http.MethodPost, "/mine", AddMyCourseRequest{CrseID: "100136", Strm: 1251, ClassNbrs: []int{1001}})
	require.Equal(t, http.StatusCreated, code)
	var mc MyCourseResponse
	data(t, assert, body, &mc)
	assert.False(mc.HasReview)

	var canonical models.CourseCanonical
	require.NoError(t, env.DB.Where("crse_id = ?", "100136").First(&canonical).Error)
	prof := models.User{Type: models.UserTypeProfessor, Name: "Bill Jannen", UnixID: "bj1"}
	require.NoError(t, env.DB.Create(&prof).Error)
	dept := models.Department{Name: "CS"}
	require.NoError(t, env.DB.Create(&dept).Error)
	area := models.AreaOfStudy{Name: "Computer Science", Abbreviation: "CSCI", DepartmentID: &dept.ID}
	require.NoError(t, env.DB.Create(&area).Error)
	legacy := models.Course{Number: "136", AreaOfStudyID: &area.ID}
	require.NoError(t, env.DB.Create(&legacy).Error)
	review := models.FactrakSurvey{UserID: student.ID, ProfessorID: prof.ID, CourseID: legacy.ID, CourseCanonicalID: &canonical.ID}
	require.NoError(t, env.DB.Create(&review).Error)

	code, body = send(t, env, http.MethodDelete, fmt.Sprintf("/mine/%d", mc.ID), nil)
	assert.Equal(http.StatusConflict, code)
	assert.Contains(string(body), fmt.Sprint(lib.ErrorMyCourseHasReview.Code))

	code, body = send(t, env, http.MethodPost, "/mine", AddMyCourseRequest{CrseID: "100136", Strm: 1271, ClassNbrs: []int{2001}})
	require.Equal(t, http.StatusCreated, code)
	var current MyCourseResponse
	data(t, assert, body, &current)
	assert.True(current.HasReview)
	code, _ = send(t, env, http.MethodDelete, fmt.Sprintf("/mine/%d", current.ID), nil)
	assert.Equal(http.StatusConflict, code)

	require.NoError(t, env.DB.Delete(&review).Error)
	for _, id := range []uint{mc.ID, current.ID} {
		code, _ = send(t, env, http.MethodDelete, fmt.Sprintf("/mine/%d", id), nil)
		assert.Equal(http.StatusNoContent, code)
	}
	var left int
	require.NoError(t, env.DB.Model(&models.UserCourseOffering{}).Count(&left).Error)
	assert.Zero(left)
}

func TestMyCoursesAreOwnerOnly(t *testing.T) {
	assert, env, _ := seedMine(t)
	other := models.User{Type: models.UserTypeStudent, Name: "Other", UnixID: "ot1"}
	require.NoError(t, env.DB.Create(&other).Error)
	var canonical models.CourseCanonical
	require.NoError(t, env.DB.Where("crse_id = ?", "100136").First(&canonical).Error)
	theirs := models.UserCourse{UserID: other.ID, CourseCanonicalID: canonical.ID}
	require.NoError(t, env.DB.Create(&theirs).Error)

	code, body := send(t, env, http.MethodGet, "/mine", nil)
	require.Equal(t, http.StatusOK, code)
	var list []MyCourseResponse
	data(t, assert, body, &list)
	assert.Empty(list)

	hide := true
	code, _ = send(t, env, http.MethodPatch, fmt.Sprintf("/mine/%d", theirs.ID), UpdateMyCourseRequest{Hidden: &hide})
	assert.Equal(http.StatusNotFound, code)
	code, _ = send(t, env, http.MethodDelete, fmt.Sprintf("/mine/%d", theirs.ID), nil)
	assert.Equal(http.StatusNotFound, code)
}
