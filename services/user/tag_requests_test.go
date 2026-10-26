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

	for i := 0; i < models.MaxOpenUserTagTasksPerRequester; i++ {
		body, _ := json.Marshal(CreateTagRequestParams{
			TagName: fmt.Sprintf("New Club %d", i),
		})
		w, err := utils.DoHTTPReq(router, http.MethodPost, fmt.Sprintf("/%d/tag-requests", user.ID), bytes.NewReader(body))
		require.NoError(t, err)
		assert.Equal(http.StatusOK, w.Code)
	}

	body, _ := json.Marshal(CreateTagRequestParams{TagName: "Another Club"})
	w, err := utils.DoHTTPReq(router, http.MethodPost, fmt.Sprintf("/%d/tag-requests", user.ID), bytes.NewReader(body))
	require.NoError(t, err)
	utils.CheckRespError(assert, w, lib.ErrorModeratorTaskTagRequestLimit)
}

func TestCreateTagRequest_RejectsExistingTag(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	cfg := utils.SetupConfig()

	user := models.User{Type: models.UserTypeStudent, Name: "Stu", UnixID: "stu"}
	require.NoError(t, db.Create(&user).Error)
	require.NoError(t, db.Create(&models.Tag{Name: "WSO"}).Error)

	router := utils.SetupRouter(auth.ScopeUsers, auth.ScopeWriteSelf)
	utils.AddUserContexts(router, user.ID)
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	body, _ := json.Marshal(CreateTagRequestParams{TagName: "WSO"})
	w, err := utils.DoHTTPReq(router, http.MethodPost, fmt.Sprintf("/%d/tag-requests", user.ID), bytes.NewReader(body))
	require.NoError(t, err)
	utils.CheckRespError(assert, w, lib.ErrorUserTagAlreadyExists)
}

func TestCreateTagRequest_ApproveCreatesAndAddsTag(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	cfg := utils.SetupConfig()

	user := models.User{Type: models.UserTypeStudent, Name: "Stu", UnixID: "stu"}
	require.NoError(t, db.Create(&user).Error)

	router := utils.SetupRouter(auth.ScopeUsers, auth.ScopeWriteSelf)
	utils.AddUserContexts(router, user.ID)
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	const tagName = "Brand New Org"
	body, _ := json.Marshal(CreateTagRequestParams{TagName: tagName})
	w, err := utils.DoHTTPReq(router, http.MethodPost, fmt.Sprintf("/%d/tag-requests", user.ID), bytes.NewReader(body))
	require.NoError(t, err)
	assert.Equal(http.StatusOK, w.Code)

	adminRouter := utils.SetupRouter(auth.ScopeAdminAll)
	utils.AddUserContexts(adminRouter, 99)
	admin.SetupRouter(adminRouter, db, cfg, zaptest.NewLogger(t).Sugar())

	var task models.ModeratorTask
	require.NoError(t, db.Where("task_type = ? AND task_reference = ?",
		models.ModeratorTaskTypeNewUserTag,
		models.UserTagNameTaskReference(user.ID, tagName),
	).First(&task).Error)

	resolveBody, _ := json.Marshal(admin.ResolveTasksBody{Resolution: models.ModeratorTaskResolutionApproved})
	w, err = utils.DoHTTPReq(adminRouter, http.MethodPost,
		fmt.Sprintf("/tasks/%d/resolve", task.ID), bytes.NewReader(resolveBody))
	require.NoError(t, err)
	assert.Equal(http.StatusOK, w.Code)

	var tag models.Tag
	require.NoError(t, db.Where("name = ?", tagName).First(&tag).Error)

	var u models.User
	require.NoError(t, db.Preload("Tags").First(&u, user.ID).Error)
	require.Len(t, u.Tags, 1)
	assert.Equal(tagName, u.Tags[0].Name)
}

func TestCreateTagRequest_DuplicateOpenName(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	cfg := utils.SetupConfig()

	user := models.User{Type: models.UserTypeStudent, Name: "Stu", UnixID: "stu"}
	require.NoError(t, db.Create(&user).Error)

	router := utils.SetupRouter(auth.ScopeUsers, auth.ScopeWriteSelf)
	utils.AddUserContexts(router, user.ID)
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	body, _ := json.Marshal(CreateTagRequestParams{TagName: "Pending Club"})
	w, err := utils.DoHTTPReq(router, http.MethodPost, fmt.Sprintf("/%d/tag-requests", user.ID), bytes.NewReader(body))
	require.NoError(t, err)
	assert.Equal(http.StatusOK, w.Code)

	w, err = utils.DoHTTPReq(router, http.MethodPost, fmt.Sprintf("/%d/tag-requests", user.ID), bytes.NewReader(body))
	require.NoError(t, err)
	utils.CheckRespError(assert, w, lib.ErrorModeratorTaskTagRequestOpen)
}
