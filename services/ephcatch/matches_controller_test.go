package ephcatch_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/ephcatch"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestController_ListMatches(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeEphcatch)
	cfg := utils.SetupConfig()

	srYear := (&models.StudentModel{}).SeniorYear()

	s := []*models.User{
		{},
		{},
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

	assert.NoError(db.Create(&models.Ephcatch{
		UserID:  s[0].ID,
		OtherID: s[1].ID,
	}).Error)
	assert.NoError(db.Create(&models.Ephcatch{
		UserID:  s[0].ID,
		OtherID: s[3].ID,
	}).Error)
	assert.NoError(db.Create(&models.Ephcatch{
		UserID:  s[0].ID,
		OtherID: s[4].ID,
	}).Error)
	assert.NoError(db.Create(&models.Ephcatch{
		UserID:  s[4].ID,
		OtherID: s[0].ID,
	}).Error)
	assert.NoError(db.Create(&models.Ephcatch{
		UserID:  s[1].ID,
		OtherID: s[0].ID,
	}).Error)

	utils.AddUserContexts(router, s[0].ID)
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	/* Create ephcatch as duplicate (expect success) */
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/matches", nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp []models.User
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct users
	assert.Len(resp, 2)
	assert.Equal(s[1].ID, resp[0].ID)
	assert.Equal(s[4].ID, resp[1].ID)
}
