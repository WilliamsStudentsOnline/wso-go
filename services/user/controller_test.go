package user

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
	"github.com/gin-gonic/gin"

	testify "github.com/stretchr/testify/assert"
)

func TestController_ListUsers(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeUsers, auth.ScopeWriteSelf)
	SetupRouter(router, db)

	// Insert test users into db
	u1 := models.User{
		Name:      "Test 1",
		UnixID:    "u1",
		ClassYear: lib.IntToPtr(3),
	}
	u2 := models.User{
		Name:      "Test 2",
		UnixID:    "u2",
		ClassYear: lib.IntToPtr(3),
		Visible:   lib.BoolToPtr(false),
	}
	u3 := models.User{
		Name:       "Test 3",
		UnixID:     "u3",
		ClassYear:  lib.IntToPtr(3),
		AtWilliams: lib.BoolToPtr(false),
	}
	u4 := models.User{
		Name:      "Test 4",
		UnixID:    "u4",
		ClassYear: lib.IntToPtr(3),
	}
	assert.NoError(db.Create(&u1).Create(&u2).Create(&u3).Create(&u4).Error)

	// Update until user pointers fixed
	assert.NoError(db.Model(&u3).Update("at_williams", false).Error)

	// Get test user (expect success)
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/", nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	resp := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(resp.Error)
	respUsers := []models.User{}
	err = json.Unmarshal(resp.Data, &respUsers)
	assert.NoError(err)

	// Check if correct user
	assert.Len(respUsers, 2)
	assert.Equal(u1.UnixID, respUsers[0].UnixID)
	assert.Equal(u4.UnixID, respUsers[1].UnixID)
}

func TestController_GetUser(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	// Insert test users into db
	u1 := models.User{
		Name:      "Test 1",
		UnixID:    "u1",
		ClassYear: lib.IntToPtr(3),
	}
	u2 := models.User{
		Name:      "Test 2",
		UnixID:    "u2",
		ClassYear: lib.IntToPtr(3),
		Visible:   lib.BoolToPtr(false),
	}
	u3 := models.User{
		Name:       "Test 3",
		UnixID:     "u3",
		ClassYear:  lib.IntToPtr(3),
		AtWilliams: lib.BoolToPtr(false),
	}
	assert.NoError(db.Create(&u1).Create(&u2).Create(&u3).Error)

	// Update until user pointers fixed
	assert.NoError(db.Model(&u3).Update("at_williams", false).Error)

	// Initialize Routing
	router := gin.Default()
	utils.AddUserContexts(router, u1.ID)
	SetupRouter(router, db)

	// Get test user (expect success)
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/%d", u1.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	resp := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(resp.Error)

	respUser := models.User{}
	err = json.Unmarshal(resp.Data, &respUser)
	assert.NoError(err)

	// Check if correct user
	assert.Equal(u1.ID, respUser.ID)
	assert.Equal(u1.UnixID, respUser.UnixID)
	assert.Equal(u1.Title, respUser.Title)

	// Get self user (expect success)
	w, err = utils.DoHTTPReq(router, http.MethodGet, "/me", nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	resp = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(resp.Error)
	respUser = models.User{}
	err = json.Unmarshal(resp.Data, &respUser)
	assert.NoError(err)

	// Check if correct user
	assert.Equal(u1.ID, respUser.ID)
	assert.Equal(u1.UnixID, respUser.UnixID)
	assert.Equal(u1.Title, respUser.Title)

	// Get user not visible (expect failure)
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/%d", u2.ID), nil)
	assert.NoError(err)
	// Error status
	assert.Equal(http.StatusBadRequest, w.Code)

	// Get user not at williams (expect failure)
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/%d", u3.ID), nil)
	assert.NoError(err)
	// Error status
	assert.Equal(http.StatusBadRequest, w.Code)

	// Get bad userID (expect failure)
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/%d", 42), nil)
	assert.NoError(err)
	// Error status
	assert.Equal(http.StatusNotFound, w.Code)
}

func TestController_UpdateUser(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	// Insert test users into db
	u1 := models.User{
		Name:        "Test 1",
		UnixID:      "u1",
		ClassYear:   lib.IntToPtr(3),
		Visible:     lib.BoolToPtr(true),
		AtWilliams:  lib.BoolToPtr(true),
		DormVisible: lib.BoolToPtr(true),
		OffCycle:    lib.BoolToPtr(false),
	}
	u2 := models.User{
		Name:      "Test 2",
		UnixID:    "u2",
		ClassYear: lib.IntToPtr(3),
	}
	assert.NoError(db.Create(&u1).Create(&u2).Error)

	router := utils.SetupRouter(auth.ScopeUsers, auth.ScopeWriteSelf)
	utils.AddUserContexts(router, u1.ID)
	SetupRouter(router, db)

	postData, err := json.Marshal(map[string]interface{}{
		"visible":     false,
		"dormVisible": false,
		"homeVisible": true,
		"offCycle":    true,
		"pronoun":     "foobar",
		"name":        "baz",
	})
	assert.NoError(err)

	// Update test user 1 (expect success)
	w, err := utils.DoHTTPReq(router, http.MethodPatch, fmt.Sprintf("/%d", u1.ID), bytes.NewBuffer(postData))
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Get updated user
	res := models.User{}
	assert.NoError(db.First(&res, u1.ID).Error)

	// Check updated params
	assert.False(*res.Visible)
	assert.False(*res.DormVisible)
	assert.True(*res.HomeVisible)
	assert.True(*res.OffCycle)
	assert.Equal("foobar", *res.Pronoun)
	// Assert that name did not change
	assert.Equal(u1.Name, res.Name)

	// Update test user 2 (expect failure, unauthed)
	w, err = utils.DoHTTPReq(router, http.MethodPatch, fmt.Sprintf("/%d", u2.ID), bytes.NewBuffer(postData))
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusForbidden, w.Code)
}
