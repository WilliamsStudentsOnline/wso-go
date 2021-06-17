package ephmatch_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

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

	assert.NoError(db.Create(&models.EphmatchRelation{
		UserID:   s[0].ID,
		OtherID:  s[1].ID,
		Relation: "like",
	}).Error)
	assert.NoError(db.Create(&models.EphmatchRelation{
		UserID:   s[0].ID,
		OtherID:  s[3].ID,
		Relation: "like",
	}).Error)
	assert.NoError(db.Create(&models.EphmatchRelation{
		UserID:   s[0].ID,
		OtherID:  s[4].ID,
		Relation: "like",
	}).Error)
	assert.NoError(db.Create(&models.EphmatchRelation{
		UserID:   s[4].ID,
		OtherID:  s[0].ID,
		Relation: "like",
	}).Error)
	assert.NoError(db.Create(&models.EphmatchRelation{
		UserID:   s[1].ID,
		OtherID:  s[0].ID,
		Relation: "like",
	}).Error)
	assert.NoError(db.Create(&models.EphmatchRelation{
		UserID:   s[3].ID,
		OtherID:  s[4].ID,
		Relation: "like",
	}).Error)
	assert.NoError(db.Create(&models.EphmatchRelation{
		UserID:   s[4].ID,
		OtherID:  s[3].ID,
		Relation: "like",
	}).Error)
	// Match user 6 and 0 but delete user 6's profile and expect it not to be returned
	assert.NoError(db.Create(&models.EphmatchRelation{
		UserID:   s[0].ID,
		OtherID:  s[5].ID,
		Relation: "like",
	}).Error)
	assert.NoError(db.Create(&models.EphmatchRelation{
		UserID:   s[5].ID,
		OtherID:  s[0].ID,
		Relation: "like",
	}).Error)

	// Matches (0,4) (0,1) (0,5) (3,4)
	assert.NoError(db.Create(&models.EphmatchMatch{
		UserAID: s[0].ID,
		UserBID: s[4].ID,
		BaseSchema: models.BaseSchema{
			CreatedAt: time.Now().Add(-4 * time.Hour),
		},
	}).Error)
	assert.NoError(db.Create(&models.EphmatchMatch{
		UserAID: s[0].ID,
		UserBID: s[1].ID,
		BaseSchema: models.BaseSchema{
			CreatedAt: time.Now().Add(-3 * time.Hour),
		},
	}).Error)
	assert.NoError(db.Create(&models.EphmatchMatch{
		UserAID: s[0].ID,
		UserBID: s[5].ID,
		BaseSchema: models.BaseSchema{
			CreatedAt: time.Now().Add(-2 * time.Hour),
		},
	}).Error)
	assert.NoError(db.Create(&models.EphmatchMatch{
		UserAID: s[3].ID,
		UserBID: s[4].ID,
		BaseSchema: models.BaseSchema{
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
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

func TestController_ListMatches_Seen(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeEphmatch, auth.ScopeEphmatchMatches)
	cfg := utils.SetupConfig()

	srYear := (&models.StudentModel{}).SeniorYear()

	s := []*models.User{
		{
			EphmatchProfile: &models.EphmatchProfile{
				Description:  lib.StrToPtr("description1"),
				MatchMessage: lib.StrToPtr("self"),
			},
		},
		{
			EphmatchProfile: &models.EphmatchProfile{
				Description:  lib.StrToPtr("description2"),
				MatchMessage: lib.StrToPtr("match 1"),
			},
		},
		{
			EphmatchProfile: &models.EphmatchProfile{
				Description:  lib.StrToPtr("description5"),
				MatchMessage: lib.StrToPtr("match 2"),
			},
		},
		{
			EphmatchProfile: &models.EphmatchProfile{
				Description:  lib.StrToPtr("description4"),
				MatchMessage: lib.StrToPtr("matched later"),
			},
		},
	}
	for i, val := range s {
		val.Name = fmt.Sprintf("Student %d", i)
		val.ID = uint(i) + 1
		val.UnixID = fmt.Sprintf("s%d", i)
		val.Type = models.UserTypeStudent
		val.ClassYear = &srYear
		assert.NoError(db.Create(val).Error)
	}

	assert.NoError(db.Create(&models.EphmatchRelation{
		UserID:   s[0].ID,
		OtherID:  s[1].ID,
		Relation: "like",
	}).Error)
	assert.NoError(db.Create(&models.EphmatchRelation{
		UserID:   s[0].ID,
		OtherID:  s[2].ID,
		Relation: "like",
	}).Error)
	assert.NoError(db.Create(&models.EphmatchRelation{
		UserID:   s[0].ID,
		OtherID:  s[3].ID,
		Relation: "like",
	}).Error)
	assert.NoError(db.Create(&models.EphmatchRelation{
		UserID:   s[1].ID,
		OtherID:  s[0].ID,
		Relation: "like",
	}).Error)
	assert.NoError(db.Create(&models.EphmatchRelation{
		UserID:   s[2].ID,
		OtherID:  s[0].ID,
		Relation: "like",
	}).Error)

	// Matches (0,1) (0,2)
	assert.NoError(db.Create(&models.EphmatchMatch{
		BaseSchema: models.BaseSchema{
			CreatedAt: time.Now().Add(-3 * time.Hour),
			UpdatedAt: time.Now().Add(-3 * time.Hour),
		},
		UserAID: s[0].ID,
		UserBID: s[1].ID,
	}).Error)
	assert.NoError(db.Create(&models.EphmatchMatch{
		BaseSchema: models.BaseSchema{
			CreatedAt: time.Now().Add(-2 * time.Hour),
			UpdatedAt: time.Now().Add(-2 * time.Hour),
		},
		UserAID: s[0].ID,
		UserBID: s[2].ID,
	}).Error)

	utils.AddUserContexts(router, s[0].ID)
	SetupRouter(router, db, cfg, zap.S())

	/* Create ephmatch as duplicate (expect success) */
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/matches", nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp []models.EphmatchMatch
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct users
	assert.Len(resp, 2)
	assert.Equal(s[1].ID, resp[1].MatchedUser.ID)
	assert.Equal(s[2].ID, resp[0].MatchedUser.ID)
	assert.False(*resp[1].SeenBySelf)
	assert.False(*resp[0].SeenBySelf)

	// match next and get seen
	assert.NoError(db.Create(&models.EphmatchRelation{
		UserID:   s[3].ID,
		OtherID:  s[0].ID,
		Relation: "like",
	}).Error)
	assert.NoError(db.Create(&models.EphmatchMatch{
		UserAID: s[0].ID,
		UserBID: s[3].ID,
		BaseSchema: models.BaseSchema{
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
	}).Error)

	/* get ephmatches */
	w, err = utils.DoHTTPReq(router, http.MethodGet, "/matches", nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	resp = []models.EphmatchMatch{}
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct users
	assert.Len(resp, 3)
	assert.Equal(s[3].ID, resp[0].MatchedUser.ID)
	assert.Equal(s[2].ID, resp[1].MatchedUser.ID)
	assert.Equal(s[1].ID, resp[2].MatchedUser.ID)
	assert.False(*resp[0].SeenBySelf)
	assert.True(*resp[1].SeenBySelf)
	assert.True(*resp[2].SeenBySelf)
}

func TestController_CountMatches(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeEphmatch, auth.ScopeEphmatchMatches)
	cfg := utils.SetupConfig()

	srYear := (&models.StudentModel{}).SeniorYear()

	s := []*models.User{
		{
			EphmatchProfile: &models.EphmatchProfile{
				Description:  lib.StrToPtr("description1"),
				MatchMessage: lib.StrToPtr("self"),
			},
		},
		{
			EphmatchProfile: &models.EphmatchProfile{
				Description:  lib.StrToPtr("description2"),
				MatchMessage: lib.StrToPtr("match 1"),
			},
		},
		{
			EphmatchProfile: &models.EphmatchProfile{
				Description:  lib.StrToPtr("description5"),
				MatchMessage: lib.StrToPtr("match 2"),
			},
		},
		{
			EphmatchProfile: &models.EphmatchProfile{
				Description:  lib.StrToPtr("description4"),
				MatchMessage: lib.StrToPtr("matched later"),
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

	assert.NoError(db.Create(&models.EphmatchRelation{
		UserID:   s[0].ID,
		OtherID:  s[1].ID,
		Relation: "like",
	}).Error)
	assert.NoError(db.Create(&models.EphmatchRelation{
		UserID:   s[0].ID,
		OtherID:  s[2].ID,
		Relation: "like",
	}).Error)
	assert.NoError(db.Create(&models.EphmatchRelation{
		UserID:   s[0].ID,
		OtherID:  s[3].ID,
		Relation: "like",
	}).Error)
	assert.NoError(db.Create(&models.EphmatchRelation{
		UserID:   s[1].ID,
		OtherID:  s[0].ID,
		Relation: "like",
	}).Error)
	assert.NoError(db.Create(&models.EphmatchRelation{
		UserID:   s[2].ID,
		OtherID:  s[0].ID,
		Relation: "like",
	}).Error)

	// Matches (0,1) (0,2)
	assert.NoError(db.Create(&models.EphmatchMatch{
		UserAID: s[0].ID,
		UserBID: s[1].ID,
	}).Error)
	assert.NoError(db.Create(&models.EphmatchMatch{
		UserAID: s[0].ID,
		UserBID: s[2].ID,
	}).Error)

	utils.AddUserContexts(router, s[0].ID)
	SetupRouter(router, db, cfg, zap.S())

	/* Create ephmatch as duplicate (expect success) */
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/matches-count", nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp CountMatchesResponse
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct users
	assert.Equal(2, resp.Total)
	assert.Equal(2, resp.Unseen)

	/* get ephmatches */
	w, err = utils.DoHTTPReq(router, http.MethodGet, "/matches", nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// match next and get seen
	assert.NoError(db.Create(&models.EphmatchRelation{
		UserID:   s[3].ID,
		OtherID:  s[0].ID,
		Relation: "like",
	}).Error)
	assert.NoError(db.Create(&models.EphmatchMatch{
		UserAID: s[0].ID,
		UserBID: s[3].ID,
	}).Error)

	/* get ephmatches count */
	w, err = utils.DoHTTPReq(router, http.MethodGet, "/matches-count", nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	resp = CountMatchesResponse{}
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct users
	assert.Equal(3, resp.Total)
	assert.Equal(1, resp.Unseen)
}
