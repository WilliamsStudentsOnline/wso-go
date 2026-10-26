package user_test

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
	admin "github.com/WilliamsStudentsOnline/wso-go/services/admin"
	. "github.com/WilliamsStudentsOnline/wso-go/services/user"
	testify "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestCreateTagRequest_EnforcesOpenLimit(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	cfg := utils.SetupConfig()

	user := models.User{Type: models.UserTypeStudent, Name: "Stu", UnixID: "stu"}
	require.NoError(t, db.Create(&user).Error)

	router := utils.SetupRouter(auth.ScopeUsers, auth.ScopeWriteSelf)
	utils.AddUserContexts(router, user.ID)
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	tags := []models.Tag{{Name: "Alpha"}, {Name: "Beta"}, {Name: "Gamma"}, {Name: "Delta"}}
	for i := range tags {
		require.NoError(t, db.Create(&tags[i]).Error)
	}

	for i := 0; i < models.MaxOpenUserTagTasksPerRequester; i++ {
		body, _ := json.Marshal(CreateTagRequestParams{
			TaskType: models.ModeratorTaskTypeNewUserTag,
			TagID:    tags[i].ID,
		})
		w, err := utils.DoHTTPReq(router, http.MethodPost, fmt.Sprintf("/%d/tag-requests", user.ID), bytes.NewReader(body))
		require.NoError(t, err)
		assert.Equal(http.StatusOK, w.Code)
	}

	body, _ := json.Marshal(CreateTagRequestParams{
		TaskType: models.ModeratorTaskTypeNewUserTag,
		TagID:    tags[3].ID,
	})
	w, err := utils.DoHTTPReq(router, http.MethodPost, fmt.Sprintf("/%d/tag-requests", user.ID), bytes.NewReader(body))
	require.NoError(t, err)
	utils.CheckRespError(assert, w, lib.ErrorModeratorTaskTagRequestLimit)
}

func TestCreateTagRequest_ApproveAddsTag(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	cfg := utils.SetupConfig()

	user := models.User{Type: models.UserTypeStudent, Name: "Stu", UnixID: "stu"}
	require.NoError(t, db.Create(&user).Error)

	router := utils.SetupRouter(auth.ScopeUsers, auth.ScopeWriteSelf)
	utils.AddUserContexts(router, user.ID)
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	tag := models.Tag{Name: "WSO"}
	require.NoError(t, db.Create(&tag).Error)

	body, _ := json.Marshal(CreateTagRequestParams{
		TaskType: models.ModeratorTaskTypeNewUserTag,
		TagID:    tag.ID,
	})
	w, err := utils.DoHTTPReq(router, http.MethodPost, fmt.Sprintf("/%d/tag-requests", user.ID), bytes.NewReader(body))
	require.NoError(t, err)
	assert.Equal(http.StatusOK, w.Code)

	adminRouter := utils.SetupRouter(auth.ScopeAdminAll)
	utils.AddUserContexts(adminRouter, 99)
	admin.SetupRouter(adminRouter, db, cfg, zaptest.NewLogger(t).Sugar())

	var task models.ModeratorTask
	require.NoError(t, db.Where("task_type = ? AND task_reference = ?",
		models.ModeratorTaskTypeNewUserTag,
		models.TaskReference(models.ModeratorTaskTypeNewUserTag, user.ID, tag.ID),
	).First(&task).Error)

	resolveBody, _ := json.Marshal(admin.ResolveTasksBody{Resolution: models.ModeratorTaskResolutionApproved})
	w, err = utils.DoHTTPReq(adminRouter, http.MethodPost,
		fmt.Sprintf("/tasks/%d/resolve", task.ID), bytes.NewReader(resolveBody))
	require.NoError(t, err)
	assert.Equal(http.StatusOK, w.Code)

	var u models.User
	require.NoError(t, db.Preload("Tags").First(&u, user.ID).Error)
	require.Len(t, u.Tags, 1)
	assert.Equal(tag.Name, u.Tags[0].Name)
}
