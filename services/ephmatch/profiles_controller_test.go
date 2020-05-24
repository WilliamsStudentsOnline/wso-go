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
	"github.com/jinzhu/gorm"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestController_ListProfiles(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeEphmatch, auth.ScopeEphmatchMatches, auth.ScopeEphmatchProfiles)
	cfg := utils.SetupConfig()

	srYear := (&models.StudentModel{}).SeniorYear()

	s := []*models.User{
		{
			EphmatchProfile: &models.EphmatchProfile{
				Description: lib.StrToPtr("foobar"),
			},
		},
		{

			EphmatchProfile: &models.EphmatchProfile{
				Description:  lib.StrToPtr("test123"),
				MatchMessage: lib.StrToPtr("matched!"),
				LocationTown: lib.StrToPtr("Portola Valley"),
			},
			Tags: []*models.Tag{
				{Name: "WOC"},
			},
		},
		{

			EphmatchProfile: &models.EphmatchProfile{
				Description:     lib.StrToPtr("hello world"),
				LocationTown:    lib.StrToPtr("Serene Lakes"),
				LocationVisible: lib.FalsePtr(),
			},
		},
		// Not student
		{
			Type:   models.UserTypeProfessor,
			Name:   "Professor 1",
			UnixID: "p1",
			EphmatchProfile: &models.EphmatchProfile{
				Description: lib.StrToPtr("42"),
			},
		},
		// Not visible
		{
			Visible: lib.BoolToPtr(false),
			EphmatchProfile: &models.EphmatchProfile{
				Description: lib.StrToPtr("2"),
			},
		},
		// Opted out
		{
			EphmatchProfile: &models.EphmatchProfile{
				Description: lib.StrToPtr("desc"),
			},
		},
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

	assert.NoError(db.Create(&models.EphmatchLike{
		UserID:  s[0].ID,
		LikedID: s[2].ID,
	}).Error)
	assert.NoError(db.Create(&models.EphmatchLike{
		UserID:  s[2].ID,
		LikedID: s[0].ID,
	}).Error)
	assert.NoError(db.Create(&models.EphmatchMatch{
		UserAID: s[2].ID,
		UserBID: s[0].ID,
	}).Error)

	assert.NoError(db.Delete(s[5].EphmatchProfile).Error)

	utils.AddUserContexts(router, s[0].ID)
	SetupRouter(router, db, cfg, zap.S())

	// Get test user
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/profiles?preload[]=tags&preload[]=liked&preload[]=matched", nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp []models.EphmatchProfile
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct users. should not return self (s[0])
	assert.Len(resp, 3)
	for i, exp := range []*models.User{s[1], s[2], s[4]} {
		assert.Equal(exp.EphmatchProfile.ID, resp[i].ID)
		assert.Equal(exp.EphmatchProfile.Description, resp[i].Description)
		assert.Equal(exp.ID, resp[i].User.ID)
	}
	// Test Liked,Matched flags
	assert.False(*resp[0].Liked)
	assert.True(*resp[1].Liked)
	assert.False(*resp[0].Matched)
	assert.True(*resp[1].Matched)

	// Test location visibility
	assert.Equal(s[1].EphmatchProfile.LocationTown, resp[0].LocationTown)
	assert.Nil(resp[1].LocationTown)
	assert.False(*resp[1].LocationVisible)

	// Test match message nil when not matched
	assert.Nil(resp[0].MatchMessage)
	assert.Equal(resp[1].MatchMessage, s[2].EphmatchProfile.MatchMessage)
	assert.Nil(resp[2].MatchMessage)

	assert.Len(resp[0].User.Tags, 1)
	assert.Equal(s[1].Tags[0].Name, resp[0].User.Tags[0].Name)

	// Get test user (sorted)
	w, err = utils.DoHTTPReq(router, http.MethodGet, "/profiles?sort=new", nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	resp = []models.EphmatchProfile{}
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct users. should not return self (s[0])
	assert.Len(resp, 3)
	for i, exp := range []*models.User{s[4], s[2], s[1]} {
		assert.Equal(exp.EphmatchProfile.ID, resp[i].ID)
		assert.Equal(exp.EphmatchProfile.Description, resp[i].Description)
		assert.Equal(exp.ID, resp[i].User.ID)
		assert.Nil(resp[i].MatchMessage) // All should be null, as not getting matched flag
	}
}

func TestController_GetProfile(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeEphmatch, auth.ScopeEphmatchMatches, auth.ScopeEphmatchProfiles)
	cfg := utils.SetupConfig()

	srYear := (&models.StudentModel{}).SeniorYear()

	s := []*models.User{
		{
			EphmatchProfile: &models.EphmatchProfile{
				Description: lib.StrToPtr("foobar"),
			},
		},
		{

			EphmatchProfile: &models.EphmatchProfile{
				Description: lib.StrToPtr("test123"),
			},
		},
		{

			EphmatchProfile: &models.EphmatchProfile{
				Description:  lib.StrToPtr("hello world"),
				MatchMessage: lib.StrToPtr("matched!"),
			},
		},
		// Not student
		{
			Type:   models.UserTypeProfessor,
			Name:   "Professor 1",
			UnixID: "p1",
			EphmatchProfile: &models.EphmatchProfile{
				Description: lib.StrToPtr("42"),
			},
		},
		// Not visible
		{
			Visible: lib.BoolToPtr(false),
			EphmatchProfile: &models.EphmatchProfile{
				Description: lib.StrToPtr("2"),
			},
		},
		// Opted out
		{
			EphmatchProfile: &models.EphmatchProfile{
				Description: lib.StrToPtr("desc"),
			},
		},
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

	assert.NoError(db.Create(&models.EphmatchLike{
		UserID:  s[0].ID,
		LikedID: s[2].ID,
	}).Error)

	assert.NoError(db.Delete(s[5].EphmatchProfile).Error)

	utils.AddUserContexts(router, s[0].ID)
	SetupRouter(router, db, cfg, zap.S())

	// Get test profile
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/profiles/%d", s[2].ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.EphmatchProfile
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct users
	assert.Equal(s[2].EphmatchProfile.ID, resp.ID)
	assert.Equal(s[2].EphmatchProfile.Description, resp.Description)
	assert.True(*resp.Liked)
	assert.Nil(resp.MatchMessage)

	// Assert these fail
	for _, u := range []*models.User{s[3], s[5]} {
		/* Get test student 1 (expect failure) */
		w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/profiles/%d", u.ID), nil)
		assert.NoError(err)

		// Status is not found
		assert.Equal(http.StatusNotFound, w.Code)
	}

	// Assert these succeede
	for _, u := range []*models.User{s[4]} {
		/* Get test student 1 (expect failure) */
		w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/profiles/%d", u.ID), nil)
		assert.NoError(err)

		// Status is not found
		assert.Equal(http.StatusOK, w.Code)
	}
}

func TestController_LikeProfile(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeEphmatch, auth.ScopeEphmatchMatches, auth.ScopeEphmatchProfiles)
	cfg := utils.SetupConfig()

	srYear := (&models.StudentModel{}).SeniorYear()

	s := []*models.User{
		{
			EphmatchProfile: &models.EphmatchProfile{
				Description: lib.StrToPtr("foobar"),
			},
		},
		{

			EphmatchProfile: &models.EphmatchProfile{
				Description: lib.StrToPtr("test123"),
			},
		},
		{

			EphmatchProfile: &models.EphmatchProfile{
				Description: lib.StrToPtr("hello world"),
			},
		},
	}
	for i, val := range s {
		val.Name = fmt.Sprintf("Student %d", i)
		val.UnixID = fmt.Sprintf("s%d", i)
		val.Type = models.UserTypeStudent
		val.ClassYear = &srYear
		assert.NoError(db.Create(val).Error)
	}

	assert.NoError(db.Create(&models.EphmatchLike{
		UserID:  s[0].ID,
		LikedID: s[2].ID,
	}).Error)

	utils.AddUserContexts(router, s[0].ID)
	SetupRouter(router, db, cfg, zap.S())

	/* Create ephmatch with self (expect failure) */
	apiErr := lib.ErrorEphmatchLikeNoSelf
	w, err := utils.DoHTTPReq(router, http.MethodPost, fmt.Sprintf("/profiles/%d/like", s[0].ID), nil)
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Create ephmatch with bad profile (expect failure) */
	apiErr = lib.ErrorEphmatchProfileNotFound
	w, err = utils.DoHTTPReq(router, http.MethodPost, fmt.Sprintf("/profiles/%d/like", 42), nil)
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Create ephmatch as duplicate (expect failure) */
	apiErr = lib.ErrorEphmatchAlreadyExists
	w, err = utils.DoHTTPReq(router, http.MethodPost, fmt.Sprintf("/profiles/%d/like", s[2].ID), nil)
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Create ephmatch (expect success) */
	w, err = utils.DoHTTPReq(router, http.MethodPost, fmt.Sprintf("/profiles/%d/like", s[1].ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusCreated, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	resp := LikeProfileResp{}
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Assert did not match
	assert.False(resp.Matched)

	// Ensure one like was created
	var likedCount int
	assert.NoError(db.Model(&models.EphmatchLike{}).Where(&models.EphmatchLike{
		UserID:  s[0].ID,
		LikedID: s[1].ID,
	}).Count(&likedCount).Error)
	assert.Equal(1, likedCount)

	// Ensure a match wasn't created
	var matchedCount int
	assert.NoError(db.Model(&models.EphmatchMatch{}).Where(&models.EphmatchMatch{
		UserAID: s[0].ID,
		UserBID: s[1].ID,
	}).Or(&models.EphmatchMatch{
		UserAID: s[1].ID,
		UserBID: s[0].ID,
	}).Count(&matchedCount).Error)
	assert.Equal(0, matchedCount)
}

func TestController_LikeProfileAndCreateMatch(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeEphmatch, auth.ScopeEphmatchMatches, auth.ScopeEphmatchProfiles)
	cfg := utils.SetupConfig()

	srYear := (&models.StudentModel{}).SeniorYear()

	s := []*models.User{
		{
			EphmatchProfile: &models.EphmatchProfile{
				Description: lib.StrToPtr("foobar"),
			},
		},
		{

			EphmatchProfile: &models.EphmatchProfile{
				Description: lib.StrToPtr("test123"),
			},
		},
	}
	for i, val := range s {
		val.Name = fmt.Sprintf("Student %d", i)
		val.UnixID = fmt.Sprintf("s%d", i)
		val.Type = models.UserTypeStudent
		val.ClassYear = &srYear
		assert.NoError(db.Create(val).Error)
	}

	assert.NoError(db.Create(&models.EphmatchLike{
		UserID:  s[1].ID,
		LikedID: s[0].ID,
	}).Error)

	utils.AddUserContexts(router, s[0].ID)
	SetupRouter(router, db, cfg, zap.S())
	likeModel := models.NewEphmatchLikeModel(db, zap.S())
	matchModel := models.NewEphmatchMatchesModel(db, zap.S())

	// Ensure match, like doesnt exist
	likeExists, err := likeModel.DoesLikeExist(s[0].ID, s[1].ID)
	assert.NoError(err)
	assert.False(likeExists)

	matchExists, err := matchModel.IsMatching(s[0].ID, s[1].ID)
	assert.NoError(err)
	assert.False(matchExists)

	/* Create ephmatch (expect success) */
	w, err := utils.DoHTTPReq(router, http.MethodPost, fmt.Sprintf("/profiles/%d/like", s[1].ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusCreated, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	resp := LikeProfileResp{}
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Assert did match
	assert.True(resp.Matched)

	// Ensure one like was created
	var likedCount int
	assert.NoError(db.Model(&models.EphmatchLike{}).Where(&models.EphmatchLike{
		UserID:  s[0].ID,
		LikedID: s[1].ID,
	}).Count(&likedCount).Error)
	assert.Equal(1, likedCount)

	// Ensure a match was created
	var matchedCount int
	assert.NoError(db.Model(&models.EphmatchMatch{}).Where(&models.EphmatchMatch{
		UserAID: s[0].ID,
		UserBID: s[1].ID,
	}).Or(&models.EphmatchMatch{
		UserAID: s[1].ID,
		UserBID: s[0].ID,
	}).Count(&matchedCount).Error)
	assert.Equal(1, matchedCount)
}

func TestController_UnlikeProfile(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeEphmatch, auth.ScopeEphmatchMatches, auth.ScopeEphmatchProfiles)
	cfg := utils.SetupConfig()

	srYear := (&models.StudentModel{}).SeniorYear()

	s := []*models.User{
		{
			EphmatchProfile: &models.EphmatchProfile{
				Description: lib.StrToPtr("foobar"),
			},
		},
		{

			EphmatchProfile: &models.EphmatchProfile{
				Description: lib.StrToPtr("test123"),
			},
		},
		{

			EphmatchProfile: &models.EphmatchProfile{
				Description: lib.StrToPtr("hello world"),
			},
		},
	}
	for i, val := range s {
		val.Name = fmt.Sprintf("Student %d", i)
		val.UnixID = fmt.Sprintf("s%d", i)
		val.Type = models.UserTypeStudent
		val.ClassYear = &srYear
		assert.NoError(db.Create(val).Error)
	}

	assert.NoError(db.Create(&models.EphmatchLike{
		UserID:  s[0].ID,
		LikedID: s[1].ID,
	}).Error)

	utils.AddUserContexts(router, s[0].ID)
	SetupRouter(router, db, cfg, zap.S())

	/* Delete ephmatch with random user (expect failure) */
	apiErr := lib.ErrorEphmatchDoesNotExist
	w, err := utils.DoHTTPReq(router, http.MethodPost, fmt.Sprintf("/profiles/%d/unlike", s[2].ID), nil)
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Delete ephmatch with bad user (expect failure) */
	apiErr = lib.ErrorEphmatchProfileNotFound
	w, err = utils.DoHTTPReq(router, http.MethodPost, fmt.Sprintf("/profiles/%d/unlike", 42), nil)
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Delete ephmatch (expect success) */
	w, err = utils.DoHTTPReq(router, http.MethodPost, fmt.Sprintf("/profiles/%d/unlike", s[1].ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	var count int
	assert.NoError(db.Model(&models.EphmatchLike{}).Where(&models.EphmatchLike{
		UserID:  s[0].ID,
		LikedID: s[1].ID,
	}).Count(&count).Error)
	assert.Equal(0, count)

	// Ensure a match wasn't created
	var matchedCount int
	assert.NoError(db.Model(&models.EphmatchMatch{}).Where(&models.EphmatchMatch{
		UserAID: s[0].ID,
		UserBID: s[1].ID,
	}).Or(&models.EphmatchMatch{
		UserAID: s[1].ID,
		UserBID: s[0].ID,
	}).Count(&matchedCount).Error)
	assert.Equal(0, matchedCount)
}

func TestController_UnlikeProfileAndDeleteMatch(t *testing.T) {
	/*
		Flow:
		1. Create like, match by u0
		 - like created
		 - match created
		2. Delete like, match by u0
		 - like deleted
		 - match deleted
		 - match soft deleted
		3. Create like, match by u0
		 - like created
		 - match updated
		4. Delete like, match by u1
		 - like deleted
		 - match deleted
		 - match soft deleted
		5. Create like, match by u1
		 - like created
		 - match updated
	*/
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	routerU0 := utils.SetupRouter(auth.ScopeEphmatch, auth.ScopeEphmatchMatches, auth.ScopeEphmatchProfiles)
	routerU1 := utils.SetupRouter(auth.ScopeEphmatch, auth.ScopeEphmatchMatches, auth.ScopeEphmatchProfiles)
	cfg := utils.SetupConfig()

	srYear := (&models.StudentModel{}).SeniorYear()

	s := []*models.User{
		{
			EphmatchProfile: &models.EphmatchProfile{
				Description: lib.StrToPtr("foobar"),
			},
		},
		{

			EphmatchProfile: &models.EphmatchProfile{
				Description: lib.StrToPtr("test123"),
			},
		},
	}
	for i, val := range s {
		val.Name = fmt.Sprintf("Student %d", i)
		val.UnixID = fmt.Sprintf("s%d", i)
		val.Type = models.UserTypeStudent
		val.ClassYear = &srYear
		assert.NoError(db.Create(val).Error)
	}

	assert.NoError(db.Create(&models.EphmatchLike{
		UserID:  s[1].ID,
		LikedID: s[0].ID,
	}).Error)

	utils.AddUserContexts(routerU0, s[0].ID)
	SetupRouter(routerU0, db, cfg, zap.S())
	utils.AddUserContexts(routerU1, s[1].ID)
	SetupRouter(routerU1, db, cfg, zap.S())

	/* 1. Create like, match (expect success) */
	w, err := utils.DoHTTPReq(routerU0, http.MethodPost, fmt.Sprintf("/profiles/%d/like", s[1].ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusCreated, w.Code)

	// Ensure like was created
	assert.Equal(1, numLikes(db, assert, s[0].ID, s[1].ID))

	// Ensure a match was created
	assert.Equal(1, numMatches(db, assert, s[0].ID, s[1].ID))

	/* 2. Delete like, match (expect success) */
	w, err = utils.DoHTTPReq(routerU0, http.MethodPost, fmt.Sprintf("/profiles/%d/unlike", s[1].ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Ensure like was deleted
	assert.Equal(0, numLikes(db, assert, s[0].ID, s[1].ID))

	// Ensure match was deleted
	assert.Equal(0, numMatches(db, assert, s[0].ID, s[1].ID))

	// Ensure match was soft deleted
	assert.Equal(1, numMatchesUnscoped(db, assert, s[0].ID, s[1].ID))

	/* 3. Create like, match (expect success) */
	w, err = utils.DoHTTPReq(routerU0, http.MethodPost, fmt.Sprintf("/profiles/%d/like", s[1].ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusCreated, w.Code)

	// Ensure like was created
	assert.Equal(1, numLikes(db, assert, s[0].ID, s[1].ID))

	// Ensure a match was created
	assert.Equal(1, numMatches(db, assert, s[0].ID, s[1].ID))

	// Ensure a match was updated (only one match exists
	assert.Equal(1, numMatchesUnscoped(db, assert, s[0].ID, s[1].ID))

	/* 4. Delete like, match by u1 (expect success) */
	w, err = utils.DoHTTPReq(routerU1, http.MethodPost, fmt.Sprintf("/profiles/%d/unlike", s[0].ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Ensure like was deleted
	assert.Equal(0, numLikes(db, assert, s[1].ID, s[0].ID))

	// Ensure match was deleted
	assert.Equal(0, numMatches(db, assert, s[0].ID, s[1].ID))

	// Ensure match was soft deleted
	assert.Equal(1, numMatchesUnscoped(db, assert, s[0].ID, s[1].ID))

	/* 5. Create like, match by u1 (expect success) */
	w, err = utils.DoHTTPReq(routerU1, http.MethodPost, fmt.Sprintf("/profiles/%d/like", s[0].ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusCreated, w.Code)

	// Ensure like was created
	assert.Equal(1, numLikes(db, assert, s[1].ID, s[0].ID))

	// Ensure a match was created
	assert.Equal(1, numMatches(db, assert, s[0].ID, s[1].ID))

	// Ensure a match was updated (only one match exists)
	assert.Equal(1, numMatchesUnscoped(db, assert, s[0].ID, s[1].ID))
}

func numLikes(db *gorm.DB, assert *testify.Assertions, userID uint, likedID uint) (count int) {
	assert.NoError(db.Model(&models.EphmatchLike{}).Where(&models.EphmatchLike{
		UserID:  userID,
		LikedID: likedID,
	}).Count(&count).Error)
	return
}

func numMatches(db *gorm.DB, assert *testify.Assertions, userID uint, likedID uint) (count int) {
	s, l := orderUIntPair(userID, likedID)
	assert.NoError(db.Model(&models.EphmatchMatch{}).Where(&models.EphmatchMatch{
		UserAID: s,
		UserBID: l,
	}).Count(&count).Error)
	return
}

func numMatchesUnscoped(db *gorm.DB, assert *testify.Assertions, userID uint, likedID uint) (count int) {
	s, l := orderUIntPair(userID, likedID)
	assert.NoError(db.Unscoped().Model(&models.EphmatchMatch{}).Where(&models.EphmatchMatch{
		UserAID: s,
		UserBID: l,
	}).Count(&count).Error)
	return
}

func orderUIntPair(a uint, b uint) (smallest uint, largest uint) {
	smallest = a
	largest = b
	if smallest > largest {
		smallest = b
		largest = a
	}
	return
}
