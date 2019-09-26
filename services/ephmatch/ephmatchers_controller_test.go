package ephmatch_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/ephmatch"
	testify "github.com/stretchr/testify/assert"
)

func TestController_ListEphmatchers(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeEphmatch)
	cfg := utils.SetupConfig()

	srYear := (&models.StudentModel{}).SeniorYear()

	s := []*models.User{
		{},
		{},
		{},
		// Not student
		{Type: models.UserTypeProfessor, Name: "Professor 1", UnixID: "p1"},
		// Not visible
		{Visible: lib.BoolToPtr(false)},
		// Opted out
		{OptOutEphmatch: lib.BoolToPtr(true)},
	}
	for i, val := range s {
		if val.Name == "" {
			val.Name = fmt.Sprintf("Student %d", i)
		}
		if val.UnixID == "" {
			val.UnixID = fmt.Sprintf("s%d", i)
		}
		if val.Type == "" {
			val.Type = models.UserTypeStudent
		}
		if val.ClassYear == nil {
			val.ClassYear = &srYear
		}
		assert.NoError(db.Create(val).Error)
	}

	assert.NoError(db.Create(&models.Ephmatch{
		UserID:  s[0].ID,
		OtherID: s[2].ID,
	}).Error)

	utils.AddUserContexts(router, s[0].ID)
	SetupRouter(router, db, cfg)

	// Get test user
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/ephmatchers", nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp []models.Ephmatcher
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct users
	assert.Len(resp, 3)
	for i, exp := range []*models.User{s[0], s[1], s[2]} {
		assert.Equal(exp.ID, resp[i].ID)
	}
	assert.False(resp[0].Liked)
	assert.False(resp[1].Liked)
	assert.True(resp[2].Liked)
}

func TestController_GetEphmatcher(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeEphmatch)
	cfg := utils.SetupConfig()

	srYear := (&models.StudentModel{}).SeniorYear()

	s := []*models.User{
		{},
		{},
		{},
		// Not student
		{Type: models.UserTypeProfessor, Name: "Professor 1", UnixID: "p1"},
		// Not visible
		{Visible: lib.BoolToPtr(false)},
		// Opted out
		{OptOutEphmatch: lib.BoolToPtr(true)},
	}
	for i, val := range s {
		if val.Name == "" {
			val.Name = fmt.Sprintf("Student %d", i)
		}
		if val.UnixID == "" {
			val.UnixID = fmt.Sprintf("s%d", i)
		}
		if val.Type == "" {
			val.Type = models.UserTypeStudent
		}
		if val.ClassYear == nil {
			val.ClassYear = &srYear
		}
		assert.NoError(db.Create(val).Error)
	}

	assert.NoError(db.Create(&models.Ephmatch{
		UserID:  s[0].ID,
		OtherID: s[2].ID,
	}).Error)

	utils.AddUserContexts(router, s[0].ID)
	SetupRouter(router, db, cfg)

	// Get test ephmatcher
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/ephmatchers/%d", s[2].ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.Ephmatcher
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct users
	assert.Equal(s[2].ID, resp.ID)
	assert.True(resp.Liked)

	// Assert these fail
	for _, u := range []*models.User{s[3], s[4], s[5]} {
		/* Get test student 1 (expect failure) */
		w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/ephmatchers/%d", u.ID), nil)
		assert.NoError(err)

		// Status is not found
		assert.Equal(http.StatusNotFound, w.Code)
	}
}

func TestController_LikeEphmatcher(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeEphmatch)
	cfg := utils.SetupConfig()

	srYear := (&models.StudentModel{}).SeniorYear()

	s := []*models.User{
		{},
		{},
		{},
	}
	for i, val := range s {
		val.Name = fmt.Sprintf("Student %d", i)
		val.UnixID = fmt.Sprintf("s%d", i)
		val.Type = models.UserTypeStudent
		val.ClassYear = &srYear
		assert.NoError(db.Create(val).Error)
	}

	assert.NoError(db.Create(&models.Ephmatch{
		UserID:  s[0].ID,
		OtherID: s[2].ID,
	}).Error)

	utils.AddUserContexts(router, s[0].ID)
	SetupRouter(router, db, cfg)

	/* Create ephmatch with self (expect failure) */
	apiErr := lib.ErrorEphmatchLikeNoSelf
	w, err := utils.DoHTTPReq(router, http.MethodPost, fmt.Sprintf("/ephmatchers/%d/like", s[0].ID), nil)
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Create ephmatch with bad ephmatcher (expect failure) */
	apiErr = lib.ErrorEphmatcherNotFound
	w, err = utils.DoHTTPReq(router, http.MethodPost, fmt.Sprintf("/ephmatchers/%d/like", 42), nil)
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Create ephmatch as duplicate (expect failure) */
	apiErr = lib.ErrorEphmatchAlreadyExists
	w, err = utils.DoHTTPReq(router, http.MethodPost, fmt.Sprintf("/ephmatchers/%d/like", s[2].ID), nil)
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Create ephmatch as duplicate (expect success) */
	w, err = utils.DoHTTPReq(router, http.MethodPost, fmt.Sprintf("/ephmatchers/%d/like", s[1].ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusCreated, w.Code)

	var count int
	assert.NoError(db.Model(&models.Ephmatch{}).Where(&models.Ephmatch{
		UserID:  s[0].ID,
		OtherID: s[1].ID,
	}).Count(&count).Error)
	assert.Equal(1, count)
}

func TestController_UnlikeEphmatcher(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeEphmatch)
	cfg := utils.SetupConfig()

	srYear := (&models.StudentModel{}).SeniorYear()

	s := []*models.User{
		{},
		{},
		{},
	}
	for i, val := range s {
		val.Name = fmt.Sprintf("Student %d", i)
		val.UnixID = fmt.Sprintf("s%d", i)
		val.Type = models.UserTypeStudent
		val.ClassYear = &srYear
		assert.NoError(db.Create(val).Error)
	}

	assert.NoError(db.Create(&models.Ephmatch{
		UserID:  s[0].ID,
		OtherID: s[1].ID,
	}).Error)

	utils.AddUserContexts(router, s[0].ID)
	SetupRouter(router, db, cfg)

	/* Delete ephmatch with random user (expect failure) */
	apiErr := lib.ErrorEphmatchDoesNotExist
	w, err := utils.DoHTTPReq(router, http.MethodPost, fmt.Sprintf("/ephmatchers/%d/unlike", s[2].ID), nil)
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Delete ephmatch with bad user (expect failure) */
	apiErr = lib.ErrorEphmatcherNotFound
	w, err = utils.DoHTTPReq(router, http.MethodPost, fmt.Sprintf("/ephmatchers/%d/unlike", 42), nil)
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Create ephmatch as duplicate (expect success) */
	w, err = utils.DoHTTPReq(router, http.MethodPost, fmt.Sprintf("/ephmatchers/%d/unlike", s[1].ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	var count int
	assert.NoError(db.Model(&models.Ephmatch{}).Where(&models.Ephmatch{
		UserID:  s[0].ID,
		OtherID: s[1].ID,
	}).Count(&count).Error)
	assert.Equal(0, count)
}
