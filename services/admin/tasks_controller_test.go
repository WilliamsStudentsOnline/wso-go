package admin_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/admin"
	testify "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap/zaptest"
)

func setupAdminTasks(t *testing.T) (*testify.Assertions, *utils.TestEnv) {
	t.Helper()
	assert := testify.New(t)
	env := utils.SetupTest(t, auth.ScopeAdminAll)
	utils.AddUserContexts(env.Router, 1)
	SetupRouter(env.Router, env.DB, env.Cfg, zaptest.NewLogger(t).Sugar())
	return assert, env
}

func seedFactrakSurvey(t *testing.T, db *gorm.DB, comment string) models.FactrakSurvey {
	t.Helper()
	assert := testify.New(t)
	prof := models.User{Type: models.UserTypeProfessor, Name: "Prof", UnixID: "prof1", AtWilliams: lib.BoolToPtr(true)}
	student := models.User{Type: models.UserTypeStudent, Name: "Student", UnixID: "stu1"}
	course := models.Course{
		Number: "100",
		AreaOfStudy: &models.AreaOfStudy{
			Name: "Computer Science", Abbreviation: "CSCI",
			Department: &models.Department{Name: "Computer Science"},
		},
	}
	assert.NoError(db.Create(&prof).Error)
	assert.NoError(db.Create(&student).Error)
	survey := models.FactrakSurvey{
		User: &student, Professor: &prof, Course: &course, Comment: comment,
	}
	assert.NoError(db.Create(&survey).Error)
	return survey
}

func TestListAndResolveFactrakTasks(t *testing.T) {
	assert, env := setupAdminTasks(t)
	survey := seedFactrakSurvey(t, env.DB, "bad review")

	uid := uint(99)
	task, err := models.NewFactrakReviewTask(&uid, survey.ID)
	require.NoError(t, err)
	require.NoError(t, env.DB.Create(task).Error)

	w, err := utils.DoHTTPReq(env.Router, http.MethodGet, "/tasks", nil)
	assert.NoError(err)
	resp := utils.GetGoodResp(assert, w)

	var groups []*ModeratorTaskGroupDTO
	assert.NoError(json.Unmarshal(resp.Data, &groups))
	require.Len(t, groups, 1)
	assert.Equal(models.SubjectKeyForFactrakSurvey(survey.ID), groups[0].SubjectKey)
	require.Len(t, groups[0].Tasks, 1)

	var raw []map[string]interface{}
	assert.NoError(json.Unmarshal(resp.Data, &raw))
	ctx := raw[0]["context"].(map[string]interface{})
	surveyJSON := ctx["survey"].(map[string]interface{})
	assert.Equal(float64(0), surveyJSON["userID"])
	assert.Equal("bad review", surveyJSON["comment"])

	body, _ := json.Marshal(ResolveTasksBody{Resolution: models.ModeratorTaskResolutionApproved})
	w, err = utils.DoHTTPReq(env.Router, http.MethodPost,
		fmt.Sprintf("/tasks/by-subject/%s/resolve", task.SubjectKey), bytes.NewReader(body))
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	var gone models.FactrakSurvey
	err = env.DB.First(&gone, survey.ID).Error
	assert.True(gorm.IsRecordNotFoundError(err))

	var resolved models.ModeratorTask
	assert.NoError(env.DB.First(&resolved, task.ID).Error)
	require.NotNil(t, resolved.ResolvedAt)
	assert.Equal(models.ModeratorTaskResolutionApproved, *resolved.Resolution)
}

func TestResolveAlreadyGoneOffersDiscard(t *testing.T) {
	assert, env := setupAdminTasks(t)

	uid := uint(3)
	task, err := models.NewFactrakReviewTask(&uid, 99999)
	require.NoError(t, err)
	require.NoError(t, env.DB.Create(task).Error)

	body, _ := json.Marshal(ResolveTasksBody{Resolution: models.ModeratorTaskResolutionApproved})
	w, err := utils.DoHTTPReq(env.Router, http.MethodPost,
		fmt.Sprintf("/tasks/by-subject/%s/resolve", task.SubjectKey), bytes.NewReader(body))
	assert.NoError(err)
	utils.CheckRespError(assert, w, lib.ErrorModeratorTaskSubjectGone)

	w, err = utils.DoHTTPReq(env.Router, http.MethodDelete,
		fmt.Sprintf("/tasks/by-subject/%s", task.SubjectKey), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	var count int
	assert.NoError(env.DB.Model(&models.ModeratorTask{}).Where("id = ?", task.ID).Count(&count).Error)
	assert.Zero(count)
}

func TestGetSubjectKey(t *testing.T) {
	assert, env := setupAdminTasks(t)
	w, err := utils.DoHTTPReq(env.Router, http.MethodGet, "/tasks/subject-key?taskType=FACTRAK_REVIEW&factrakSurveyID=42", nil)
	assert.NoError(err)
	resp := utils.GetGoodResp(assert, w)
	var out SubjectKeyResponse
	assert.NoError(json.Unmarshal(resp.Data, &out))
	assert.Equal("factrak_survey:42", out.SubjectKey)
}
