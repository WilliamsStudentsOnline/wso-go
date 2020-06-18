package ahe2nht1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestController_GetUserByUnix(t *testing.T) {
	/* SETUP */
	// (can copy and paste this basically)
	require := assert.New(t)
	db := utils.SetupServiceTest(require)

	// Insert test users into db
	u1 := models.User{
		Name:      "Test 1",
		UnixID:    "unix1",
		ClassYear: lib.IntToPtr(2022),
	}
	require.NoError(db.Create(&u1).Error)

	// Initialize Routing
	router := utils.SetupRouter(auth.ScopeUsers)
	utils.AddUserContexts(router, u1.ID)
	cfg := utils.SetupConfig()
	logger := zap.S()
	SetupRouter(router, db, cfg, logger)

	/* TEST 1: Get test user 1 (expect success) */
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/unix1", nil)
	require.NoError(err)

	// Status is okay
	fmt.Printf("W.code: %d", w.Code)
	require.Equal(http.StatusOK, w.Code)

	// Decode response
	// Ensure no errors from response
	resp := utils.GetHTTPDataResp(require, w.Body.Bytes())
	require.Nil(resp.Error)

	// Get the decoded response in respUser
	respUser := models.User{}
	err = json.Unmarshal(resp.Data, &respUser)
	require.NoError(err)

	// Ensure it is the correct user
	require.Equal(u1.ID, respUser.ID)
	require.Equal(u1.UnixID, respUser.UnixID)
	require.Equal(u1.Title, respUser.Title)

	// /* TEST 2: Get user with bad unix ID (expect failure) */
	// w, err = utils.DoHTTPReq(router, http.MethodGet, "fakeUnix", nil)
	// require.NoError(err)
	// // Error status
	// require.Equal(http.StatusNotFound, w.Code)
}
