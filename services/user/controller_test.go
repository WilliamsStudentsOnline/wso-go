package user

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"

	"github.com/gin-gonic/gin"
	testify "github.com/stretchr/testify/assert"
)

func TestController_GetUser(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := gin.Default()
	SetupRouter(router, db)

	// Insert test user into db
	user := models.User{
		BaseSchema: models.BaseSchema{
			ID: 1,
		},
		Name:      "Test",
		ClassYear: lib.IntToPtr(3),
	}
	err := db.FirstOrCreate(&user).Error
	assert.NoError(err)

	// Get test user
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/1", nil)
	assert.NoError(err)

	// Status is okay
	t.Log(utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error)
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes()).Data
	respUser := models.User{}
	err = json.Unmarshal(respData, &respUser)
	assert.NoError(err)

	// Check if correct user
	assert.Equal(user.ID, respUser.ID)
	assert.Equal(user.Title, respUser.Title)
}
