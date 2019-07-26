package user

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"

	testify "github.com/stretchr/testify/assert"
)

func TestController_GetUser(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeUsers, auth.ScopeWriteSelf)
	SetupRouter(router, db)

	// Insert test user into db
	user := models.User{
		BaseSchema: models.BaseSchema{
			ID: 1,
		},
		Name:      "Test",
		UnixID:    "u1",
		Visible:   true,
		ClassYear: lib.IntToPtr(3),
	}
	err := db.FirstOrCreate(&user).Error
	assert.NoError(err)

	// Get test user
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/1", nil)
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
	assert.Equal(user.ID, respUser.ID)
	assert.Equal(user.Title, respUser.Title)
}
