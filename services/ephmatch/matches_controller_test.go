package ephmatch_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

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
	router := utils.SetupRouter(auth.ScopeEphmatch)
	cfg := utils.SetupConfig()

	srYear := (&models.StudentModel{}).SeniorYear()

	s := []*models.User{
		{
			EphmatchProfile: &models.EphmatchProfile{
				Gender:      "he/him/his",
				Description: "description1",
			},
		},
		{
			EphmatchProfile: &models.EphmatchProfile{
				Gender:      "gender2",
				Description: "description2",
			},
		},
		{
			EphmatchProfile: &models.EphmatchProfile{
				Gender:      "gender3",
				Description: "description3",
			},
		},
		{
			EphmatchProfile: &models.EphmatchProfile{
				Gender:      "gender4",
				Description: "description4",
			},
		},
		{
			EphmatchProfile: &models.EphmatchProfile{
				Gender:      "gender5",
				Description: "description5",
			},
		},
		{
			EphmatchProfile: &models.EphmatchProfile{
				Gender:      "gender6",
				Description: "description6",
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

	assert.NoError(db.Create(&models.Ephmatch{
		UserID:  s[0].ID,
		OtherID: s[1].ID,
	}).Error)
	assert.NoError(db.Create(&models.Ephmatch{
		UserID:  s[0].ID,
		OtherID: s[3].ID,
	}).Error)
	assert.NoError(db.Create(&models.Ephmatch{
		UserID:  s[0].ID,
		OtherID: s[4].ID,
	}).Error)
	assert.NoError(db.Create(&models.Ephmatch{
		UserID:  s[4].ID,
		OtherID: s[0].ID,
	}).Error)
	assert.NoError(db.Create(&models.Ephmatch{
		UserID:  s[1].ID,
		OtherID: s[0].ID,
	}).Error)
	// Match user 6 and 0 but delete user 6's profile and expect it not to be returned
	assert.NoError(db.Create(&models.Ephmatch{
		UserID:  s[0].ID,
		OtherID: s[5].ID,
	}).Error)
	assert.NoError(db.Create(&models.Ephmatch{
		UserID:  s[5].ID,
		OtherID: s[0].ID,
	}).Error)
	// Delete user 6 and expect it not to be a response
	assert.NoError(db.Delete(s[5].EphmatchProfile).Error)

	utils.AddUserContexts(router, s[0].ID)
	SetupRouter(router, db, cfg, zap.S())

	/* Create ephmatch as duplicate (expect success) */
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/matches", nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp []models.Ephmatch
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct users
	assert.Len(resp, 2)
	assert.Equal(s[1].ID, resp[0].Other.ID)
	assert.Equal(s[4].ID, resp[1].Other.ID)
	assert.Equal(s[1].EphmatchProfile.Description, resp[0].Other.EphmatchProfile.Description)
	assert.Equal(s[1].EphmatchProfile.Gender, resp[0].Other.EphmatchProfile.Gender)
	assert.Equal(s[4].EphmatchProfile.Description, resp[1].Other.EphmatchProfile.Description)
	assert.Equal(s[4].EphmatchProfile.Gender, resp[1].Other.EphmatchProfile.Gender)
}
