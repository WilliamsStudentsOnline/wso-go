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
	"go.uber.org/zap"
)

func TestController_ListMatches(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeEphmatch, auth.ScopeEphmatchMatches)
	cfg := utils.SetupConfig()

	srYear := (&models.StudentModel{}).SeniorYear()

	s := []*models.User{
		{
			EphmatchProfile: &models.EphmatchProfile{
				Description: lib.StrToPtr("description1"),
			},
		},
		{
			EphmatchProfile: &models.EphmatchProfile{
				Description:  lib.StrToPtr("description2"),
				MatchMessage: lib.StrToPtr("matched!"),
				LocationTown: lib.StrToPtr("Portola Valley"),
			},
			Tags: []*models.Tag{
				{Name: "WOC"},
			},
		},
		{
			EphmatchProfile: &models.EphmatchProfile{
				Description: lib.StrToPtr("description3"),
			},
		},
		{
			EphmatchProfile: &models.EphmatchProfile{
				Description: lib.StrToPtr("description4"),
			},
		},
		{
			EphmatchProfile: &models.EphmatchProfile{
				Description:     lib.StrToPtr("description5"),
				LocationTown:    lib.StrToPtr("Williamstown"),
				LocationVisible: lib.FalsePtr(),
			},
		},
		{
			EphmatchProfile: &models.EphmatchProfile{
				Description: lib.StrToPtr("description6"),
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
	assert.NoError(db.Create(&models.EphmatchLike{
		UserID:  s[0].ID,
		LikedID: s[3].ID,
	}).Error)
	assert.NoError(db.Create(&models.EphmatchLike{
		UserID:  s[0].ID,
		LikedID: s[4].ID,
	}).Error)
	assert.NoError(db.Create(&models.EphmatchLike{
		UserID:  s[4].ID,
		LikedID: s[0].ID,
	}).Error)
	assert.NoError(db.Create(&models.EphmatchLike{
		UserID:  s[1].ID,
		LikedID: s[0].ID,
	}).Error)
	assert.NoError(db.Create(&models.EphmatchLike{
		UserID:  s[3].ID,
		LikedID: s[4].ID,
	}).Error)
	assert.NoError(db.Create(&models.EphmatchLike{
		UserID:  s[4].ID,
		LikedID: s[3].ID,
	}).Error)
	// Match user 6 and 0 but delete user 6's profile and expect it not to be returned
	assert.NoError(db.Create(&models.EphmatchLike{
		UserID:  s[0].ID,
		LikedID: s[5].ID,
	}).Error)
	assert.NoError(db.Create(&models.EphmatchLike{
		UserID:  s[5].ID,
		LikedID: s[0].ID,
	}).Error)

	// Matches (0,4) (0,1) (0,5) (3,4)
	assert.NoError(db.Create(&models.EphmatchMatch{
		UserAID: s[4].ID,
		UserBID: s[0].ID,
	}).Error)
	assert.NoError(db.Create(&models.EphmatchMatch{
		UserAID: s[1].ID,
		UserBID: s[0].ID,
	}).Error)
	assert.NoError(db.Create(&models.EphmatchMatch{
		UserAID: s[5].ID,
		UserBID: s[0].ID,
	}).Error)
	assert.NoError(db.Create(&models.EphmatchMatch{
		UserAID: s[4].ID,
		UserBID: s[3].ID,
	}).Error)

	// Delete user 6 and expect it not to be a response
	assert.NoError(db.Delete(s[5].EphmatchProfile).Error)

	utils.AddUserContexts(router, s[0].ID)
	SetupRouter(router, db, cfg, zap.S())

	/* Create ephmatch as duplicate (expect success) */
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/matches?preload[]=tags", nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp []models.EphmatchMatch
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct users
	assert.Len(resp, 2)
	assert.Equal(s[1].ID, resp[0].MatchedUser.ID)
	assert.Equal(s[4].ID, resp[1].MatchedUser.ID)
	assert.Equal(s[1].EphmatchProfile.Description, resp[0].MatchedUser.EphmatchProfile.Description)
	assert.Equal(s[1].EphmatchProfile.MatchMessage, resp[0].MatchedUser.EphmatchProfile.MatchMessage)
	assert.Equal(s[1].EphmatchProfile.LocationTown, resp[0].MatchedUser.EphmatchProfile.LocationTown)
	assert.Equal(s[4].EphmatchProfile.Description, resp[1].MatchedUser.EphmatchProfile.Description)
	assert.Nil(resp[1].MatchedUser.EphmatchProfile.LocationTown)
	assert.False(*resp[1].MatchedUser.EphmatchProfile.LocationVisible)
	assert.Len(resp[0].MatchedUser.Tags, 1)
	assert.Equal(s[1].Tags[0].Name, resp[0].MatchedUser.Tags[0].Name)
}
