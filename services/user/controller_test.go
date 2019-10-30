package user_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/user"
	"github.com/gin-gonic/gin"

	testify "github.com/stretchr/testify/assert"
)

func TestController_ListUsers(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeUsers, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg)

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

	// Test 1: get test user (expect success)
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

	// Test 2: get test user via search (expect success)
	// Update users to have search fields
	for _, u := range []models.User{u1, u2, u3, u4} {
		assert.NoError(models.NewUserModel(db).PopulateSearchFields(u.ID))
	}

	qs := url.Values{}
	qs.Add("q", "test 4")
	w, err = utils.DoHTTPReq(router, http.MethodGet, "/?"+qs.Encode(), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	resp = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(resp.Error)
	respUsers = []models.User{}
	err = json.Unmarshal(resp.Data, &respUsers)
	assert.NoError(err)

	// Check if correct user
	assert.Len(respUsers, 1)
	assert.Equal(u4.UnixID, respUsers[0].UnixID)
}

func TestController_ListUsers_Pagination(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeUsers, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg)

	users := []*models.User{
		{
			Name:   "Test A",
			UnixID: "a",
		},
		{
			Name:   "Test B",
			UnixID: "b",
		},
		{
			Name:   "Test C",
			UnixID: "c",
		},
		{
			Name:   "Test D",
			UnixID: "d",
		},
		{
			Name:   "Test E",
			UnixID: "e",
		},
	}
	for _, user := range users {
		assert.NoError(db.Create(user).Error)
	}
	// Update users to have search fields
	for _, u := range users {
		assert.NoError(models.NewUserModel(db).PopulateSearchFields(u.ID))
	}

	testCases := []struct {
		name     string
		query    string
		expected []*models.User
	}{
		{
			"all",
			"",
			users,
		},
		{
			"limit",
			"limit=2",
			users[0:2],
		},
		{
			"offset (ignored)",
			"offset=2",
			users,
		},
		{
			"start",
			"start=Test+B",
			users[2:],
		},
		{
			"offset and limit",
			"offset=2&limit=2",
			users[2:4],
		},
		{
			"start and limit",
			"start=Test+B&limit=2",
			users[2:4],
		},
		{
			"start and offset and limit",
			"start=Test+B&offset=1&limit=1",
			users[3:4],
		},
	}

	// Test for db query
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := testify.New(t)
			w, err := utils.DoHTTPReq(router, http.MethodGet, "/?"+tc.query, nil)
			a.NoError(err)
			a.Equal(http.StatusOK, w.Code)

			resp := utils.GetHTTPDataResp(assert, w.Body.Bytes())
			a.Nil(resp.Error)

			var res []*models.User
			err = json.Unmarshal(resp.Data, &res)
			a.NoError(err)

			a.Len(res, len(tc.expected))
			for i := range tc.expected {
				a.Equal(tc.expected[i].ID, res[i].ID)
			}
		})
	}

	// Test for search query
	for _, tc := range testCases {
		t.Run(tc.name+" [search]", func(t *testing.T) {
			a := testify.New(t)

			// Add an & b/c we are doing a query, but only if query has text
			if tc.query != "" {
				tc.query = "&" + tc.query
			}

			w, err := utils.DoHTTPReq(router, http.MethodGet, "/?q=test"+tc.query, nil)
			a.NoError(err)
			a.Equal(http.StatusOK, w.Code)

			resp := utils.GetHTTPDataResp(assert, w.Body.Bytes())
			a.Nil(resp.Error)

			var res []*models.User
			err = json.Unmarshal(resp.Data, &res)
			a.NoError(err)

			a.Len(res, len(tc.expected))
			for i := range tc.expected {
				a.Equal(tc.expected[i].ID, res[i].ID)
			}
		})
	}
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
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg)

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
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg)

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
