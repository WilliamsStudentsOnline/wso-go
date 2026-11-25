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
	"go.uber.org/zap/zaptest"
)

type usersFixture struct {
	admin, student, moderator models.User
}

func setupAdminUsers(t *testing.T) (*testify.Assertions, *utils.TestEnv, usersFixture) {
	t.Helper()
	assert := testify.New(t)
	env := utils.SetupTest(t, auth.ScopeAdminAll)

	f := usersFixture{
		admin:     models.User{Type: models.UserTypeStudent, Name: "Ada Admin", UnixID: "aa1", Admin: lib.TruePtr()},
		student:   models.User{Type: models.UserTypeStudent, Name: "Sam Student", UnixID: "ss1"},
		moderator: models.User{Type: models.UserTypeStudent, Name: "Mo Moderator", UnixID: "mm1", FactrakAdmin: lib.TruePtr()},
	}
	assert.NoError(env.DB.Create(&f.admin).Error)
	assert.NoError(env.DB.Create(&f.student).Error)
	assert.NoError(env.DB.Create(&f.moderator).Error)

	utils.AddUserContexts(env.Router, f.admin.ID)
	SetupRouter(env.Router, env.DB, env.Cfg, zaptest.NewLogger(t).Sugar())
	return assert, env, f
}

func listUsers(assert *testify.Assertions, env *utils.TestEnv, query string) []UserAccess {
	w, err := utils.DoHTTPReq(env.Router, http.MethodGet, "/users"+query, nil)
	assert.NoError(err)
	var resp struct {
		Users []UserAccess `json:"users"`
	}
	assert.NoError(json.Unmarshal(utils.GetGoodResp(assert, w).Data, &resp))
	return resp.Users
}

func putAccess(assert *testify.Assertions, env *utils.TestEnv, userID uint, params UpdateUserAccessParams) (int, []byte) {
	body, _ := json.Marshal(params)
	w, err := utils.DoHTTPReq(env.Router, http.MethodPut, fmt.Sprintf("/users/%d/access", userID), bytes.NewReader(body))
	assert.NoError(err)
	return w.Code, w.Body.Bytes()
}

func TestAdminUsers_ListAndSearch(t *testing.T) {
	assert, env, f := setupAdminUsers(t)

	users := listUsers(assert, env, "")
	assert.Len(users, 2)
	assert.Equal(f.admin.ID, users[0].UserID)
	assert.Equal(f.moderator.ID, users[1].UserID)

	users = listUsers(assert, env, "?q=ss1")
	assert.Len(users, 1)
	assert.Equal("Sam Student", users[0].Name)

	users = listUsers(assert, env, "?q=student")
	assert.Len(users, 1)
}

func TestAdminUsers_BanAndUnban(t *testing.T) {
	assert, env, f := setupAdminUsers(t)

	code, _ := putAccess(assert, env, f.student.ID, UpdateUserAccessParams{
		Bans: UserBans{BulletinWrite: true, Factrak: true, Reason: "  spam  "},
	})
	assert.Equal(http.StatusOK, code)

	w, err := utils.DoHTTPReq(env.Router, http.MethodGet, fmt.Sprintf("/users/%d/access", f.student.ID), nil)
	assert.NoError(err)
	var got UserAccess
	assert.NoError(json.Unmarshal(utils.GetGoodResp(assert, w).Data, &got))
	assert.True(got.Bans.BulletinWrite)
	assert.True(got.Bans.Factrak)
	assert.False(got.Bans.BulletinRead)
	assert.Equal("spam", got.Bans.Reason)

	users := listUsers(assert, env, "")
	assert.Len(users, 3)

	code, _ = putAccess(assert, env, f.student.ID, UpdateUserAccessParams{
		Bans: UserBans{BulletinWrite: true, Reason: "spam"},
	})
	assert.Equal(http.StatusOK, code)
	var count int
	assert.NoError(env.DB.Model(&models.BannedUser{}).Where("user_id = ?", f.student.ID).Count(&count).Error)
	assert.Equal(1, count)

	code, _ = putAccess(assert, env, f.student.ID, UpdateUserAccessParams{})
	assert.Equal(http.StatusOK, code)
	assert.NoError(env.DB.Model(&models.BannedUser{}).Where("user_id = ?", f.student.ID).Count(&count).Error)
	assert.Equal(0, count)
	assert.Len(listUsers(assert, env, ""), 2)
}

func TestAdminUsers_ToggleFactrakModerator(t *testing.T) {
	assert, env, f := setupAdminUsers(t)

	code, _ := putAccess(assert, env, f.student.ID, UpdateUserAccessParams{FactrakAdmin: true})
	assert.Equal(http.StatusOK, code)
	code, _ = putAccess(assert, env, f.moderator.ID, UpdateUserAccessParams{FactrakAdmin: false})
	assert.Equal(http.StatusOK, code)

	var student, moderator models.User
	assert.NoError(env.DB.First(&student, f.student.ID).Error)
	assert.NoError(env.DB.First(&moderator, f.moderator.ID).Error)
	assert.True(*student.FactrakAdmin)
	assert.False(*moderator.FactrakAdmin)
}

func TestAdminUsers_AdminTargetsLocked(t *testing.T) {
	assert, env, f := setupAdminUsers(t)

	body, _ := json.Marshal(UpdateUserAccessParams{Bans: UserBans{BulletinWrite: true}})
	w, err := utils.DoHTTPReq(env.Router, http.MethodPut, fmt.Sprintf("/users/%d/access", f.admin.ID), bytes.NewReader(body))
	assert.NoError(err)
	utils.CheckRespError(assert, w, lib.ErrorAdminUserAccessLocked)

	var count int
	assert.NoError(env.DB.Model(&models.BannedUser{}).Count(&count).Error)
	assert.Zero(count)
}
