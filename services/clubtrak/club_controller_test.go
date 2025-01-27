package clubtrak_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/clubtrak"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestController_CreateClub(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeUsers)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	u1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "User 1",
		UnixID: "u1",
	}

	assert.NoError(db.Create(&u1).Error)
	utils.AddUserContexts(router, u1.ID)

	//Create Club Params
	params := ClubParams{
		Name:               "WSO",
		Subscribers:        30,
		MeetingDescription: "Sundays at 1pm in Wach B11",
		ClubID:             1,
		ClubAdminID:        &u1,
	}

	//Process Params
	paramsData, err := json.Marshal(&params)
	assert.NoError(err)

	// Test HTTP Request
	w, err := utils.DoHTTPReq(router, http.MethodPost, "/clubs", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(http.StatusCreated, w.Code)

}

func TestController_GetAllClubs(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeUsers)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	u1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "User 1",
		UnixID: "u1",
	}

	c1 := models.Club{
		Name:               "WSO",
		Category:           "STEM",
		ClubAdminID:        &u1,
		ClubDescription:    "Write code for WSO!",
		MeetingDescription: "Sundays 1-3pm in Wach B11",
	}

	assert.NoError(db.Create(&u1).Create(&c1).Error)

	c2 := models.Club{
		Name:               "Octet",
		Category:           "Performing Arts",
		ClubAdminID:        &u1,
		ClubDescription:    "Sing without instruments!",
		MeetingDescription: "Mondays 1-3pm in Bernhardt",
	}

	assert.NoError(db.Create(&c2).Error)

	c3 := models.Club{
		Name:               "Kusika",
		Category:           "Performing Arts",
		ClubAdminID:        &u1,
		ClubDescription:    "Learn African Drumming and Dance from across the diaspora!",
		MeetingDescription: "Mon Wed Fri 4-6pm in '62 Center Dance Studio",
	}

	assert.NoError(db.Create(&c3).Error)

	// Test Endpoint
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/clubs", nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp []models.Club
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct neighborhoods
	assert.Len(resp, 3)
	assert.Equal(c1.Name, resp[0].Name)
	assert.Equal(c2.Name, resp[1].Name)
	assert.Equal(c3.Name, resp[2].Name)
}
