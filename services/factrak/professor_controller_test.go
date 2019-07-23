package factrak_test

import (
	"encoding/json"
	"net/http"
	"testing"

	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/gin-gonic/gin"
	testify "github.com/stretchr/testify/assert"

	. "github.com/WilliamsStudentsOnline/wso-go/services/factrak"
)

func TestController_FetchAllProfessors(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := gin.Default()
	SetupRouter(router, db)

	// Insert test user into db
	p1 := models.User{
		Type: models.UserTypeProfessor,
		Name:      "Prof1",
		UnixID: "p1",
		Visible: true,
		AtWilliams: true,
	}
	p2 := models.User{
		Type: models.UserTypeProfessor,
		Name:      "Prof2",
		UnixID: "p2",
		Visible: true,
		AtWilliams: true,
	}
	// Should not show up
	s1 := models.User{
		Type: models.UserTypeStudent,
		Name:      "Student1",
		UnixID: "s1",
		Visible: true,
		AtWilliams: true,
	}
	// Not at williams
	p3 := models.User{
		Type: models.UserTypeProfessor,
		Name:      "Prof3",
		UnixID: "p3",
		Visible: true,
		AtWilliams: false,
	}
	err := db.Create(&p1).Create(&p2).Create(&s1).Create(&p3).Error
	assert.NoError(err)

	// Get test user
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/professors", nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp []models.User
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct users
	assert.Len(resp, 2)
	assert.Equal(p1.ID, resp[0].ID)
	assert.Equal(p1.UnixID, resp[0].UnixID)
	assert.Equal(p2.ID, resp[1].ID)
	assert.Equal(p2.UnixID, resp[1].UnixID)
}

func TestController_GetProfessor(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := gin.Default()
	SetupRouter(router, db)

	// Insert test user into db
	p1 := models.User{
		Type: models.UserTypeProfessor,
		Name:      "Prof1",
		UnixID: "p1",
		Visible: true,
		AtWilliams: true,
	}
	p2 := models.User{
		Type: models.UserTypeProfessor,
		Name:      "Prof2",
		UnixID: "p2",
		Visible: true,
		AtWilliams: true,
	}
	err := db.Create(&p1).Create(&p2).Error
	assert.NoError(err)

	// Get test user
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/professors", nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp []models.User
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct users
	assert.Len(resp, 2)
	assert.Equal(p1.ID, resp[0].ID)
	assert.Equal(p1.UnixID, resp[0].UnixID)
	assert.Equal(p2.ID, resp[1].ID)
	assert.Equal(p2.UnixID, resp[1].UnixID)
}