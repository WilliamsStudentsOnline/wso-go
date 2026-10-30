package admin_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/admin"
	"github.com/jinzhu/gorm"
	testify "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	var listed []*ModeratorTaskResponse
	assert.NoError(json.Unmarshal(resp.Data, &listed))
	require.Len(t, listed, 1)
	assert.Equal(models.TaskReference(models.ModeratorTaskTypeFactrakReview, survey.ID), listed[0].TaskReference)

	var raw []map[string]interface{}
	assert.NoError(json.Unmarshal(resp.Data, &raw))
	ctx := raw[0]["context"].(map[string]interface{})
	surveyJSON := ctx["survey"].(map[string]interface{})
	assert.Equal(float64(0), surveyJSON["userID"])
	assert.Equal("bad review", surveyJSON["comment"])

	w, err = utils.DoHTTPReq(env.Router, http.MethodGet, fmt.Sprintf("/tasks/%d", task.ID), nil)
	assert.NoError(err)
	resp = utils.GetGoodResp(assert, w)
	var one ModeratorTaskResponse
	assert.NoError(json.Unmarshal(resp.Data, &one))
	assert.Equal(task.ID, one.ID)
	assert.NotNil(one.Context)

	body, _ := json.Marshal(ResolveTasksBody{Resolution: models.ModeratorTaskResolutionApproved})
	w, err = utils.DoHTTPReq(env.Router, http.MethodPost,
		fmt.Sprintf("/tasks/%d/resolve", task.ID), bytes.NewReader(body))
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

func TestResolveAlreadyGoneThenDelete(t *testing.T) {
	assert, env := setupAdminTasks(t)

	uid := uint(3)
	task, err := models.NewFactrakReviewTask(&uid, 99999)
	require.NoError(t, err)
	require.NoError(t, env.DB.Create(task).Error)

	body, _ := json.Marshal(ResolveTasksBody{Resolution: models.ModeratorTaskResolutionApproved})
	w, err := utils.DoHTTPReq(env.Router, http.MethodPost,
		fmt.Sprintf("/tasks/%d/resolve", task.ID), bytes.NewReader(body))
	assert.NoError(err)
	utils.CheckRespError(assert, w, lib.ErrorModeratorTaskSubjectGone)

	w, err = utils.DoHTTPReq(env.Router, http.MethodDelete, fmt.Sprintf("/tasks/%d", task.ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	var count int
	assert.NoError(env.DB.Model(&models.ModeratorTask{}).Where("id = ?", task.ID).Count(&count).Error)
	assert.Zero(count)
}

func TestResolveAppliesToSharedReference(t *testing.T) {
	assert, env := setupAdminTasks(t)
	survey := seedFactrakSurvey(t, env.DB, "flagged twice")

	uid1, uid2 := uint(1), uint(2)
	t1, err := models.NewFactrakReviewTask(&uid1, survey.ID)
	require.NoError(t, err)
	t2, err := models.NewFactrakReviewTask(&uid2, survey.ID)
	require.NoError(t, err)
	require.NoError(t, env.DB.Create(t1).Error)
	require.NoError(t, env.DB.Create(t2).Error)

	body, _ := json.Marshal(ResolveTasksBody{Resolution: models.ModeratorTaskResolutionRejected})
	w, err := utils.DoHTTPReq(env.Router, http.MethodPost,
		fmt.Sprintf("/tasks/%d/resolve", t1.ID), bytes.NewReader(body))
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	var resolved []*models.ModeratorTask
	assert.NoError(env.DB.Where("task_reference = ?", t1.TaskReference).Find(&resolved).Error)
	require.Len(t, resolved, 2)
	for _, task := range resolved {
		require.NotNil(t, task.ResolvedAt)
		assert.Equal(models.ModeratorTaskResolutionRejected, *task.Resolution)
	}
}

func TestResolveBoardThreadReviewDeletesThread(t *testing.T) {
	assert, env := setupAdminTasks(t)

	author := models.User{Type: models.UserTypeStudent, UnixID: "author", Name: "Author"}
	assert.NoError(env.DB.Create(&author).Error)
	thread := models.BoardThread{
		Type: models.BoardThreadTypeDiscussion, Title: "flagged", UserID: author.ID,
		ExUserName: author.Name, RepliesEnabled: lib.TruePtr(), LastActive: time.Now(),
	}
	assert.NoError(env.DB.Create(&thread).Error)
	op := models.BoardPost{
		ThreadID: thread.ID, UserID: author.ID, ExUserName: author.Name,
		Content: "body", IsOP: lib.TruePtr(),
	}
	assert.NoError(env.DB.Create(&op).Error)

	uid := uint(7)
	task, err := models.NewBoardThreadReviewTask(&uid, thread.ID)
	require.NoError(t, err)
	require.NoError(t, env.DB.Create(task).Error)

	body, _ := json.Marshal(ResolveTasksBody{Resolution: models.ModeratorTaskResolutionApproved})
	w, err := utils.DoHTTPReq(env.Router, http.MethodPost,
		fmt.Sprintf("/tasks/%d/resolve", task.ID), bytes.NewReader(body))
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	var gone models.BoardThread
	err = env.DB.First(&gone, thread.ID).Error
	assert.True(gorm.IsRecordNotFoundError(err))

	var resolved models.ModeratorTask
	assert.NoError(env.DB.First(&resolved, task.ID).Error)
	require.NotNil(t, resolved.ResolvedAt)
	assert.Equal(models.ModeratorTaskResolutionApproved, *resolved.Resolution)
}
