package mgb4

import (
	"encoding/json"
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
	require := assert.New(t)              // Create a testing assertion
	db := utils.SetupServiceTest(require) // Initialize test DB

	//Insert test users into db
	u1 := models.User{
		Name:      "Test 1",
		UnixID:    "unix1",
		ClassYear: lib.IntToPtr(2022),
	}
	require.NoError(db.Create(&u1).Error) // Check insertion works

	// Initialize routing
	router := utils.SetupRouter(auth.ScopeUsers) // The users scope authorizes the test router
	utils.AddUserContexts(router, u1.ID)         // Send requests from user 1
	cfg := utils.SetupConfig()
	logger := zap.S()
	SetupRouter(router, db, cfg, logger)

	// Run tests
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/unix1", nil) // GET request for unix1
	require.NoError(err)

	// Decode the response
	resp := utils.GetHTTPDataResp(require, w.Body.Bytes())
	require.Nil(resp.Error)

	// Format the response data as a user
	respUser := models.User{}
	err = json.Unmarshal(resp.Data, &respUser)
	require.NoError(err)

	// Tests for valid user
	require.Equal(u1.UnixID, respUser.UnixID)
	require.Equal(u1.Name, respUser.Name)
	require.Equal(u1.ClassYear, respUser.ClassYear)
	require.Equal(u1.ID, respUser.ID)

	// Tests for invalid user
	w, err = utils.DoHTTPReq(router, http.MethodGet, "invalid", nil)
	require.NoError(err)
	require.Equal(http.StatusNotFound, w.Code) // Require 404 error
}
