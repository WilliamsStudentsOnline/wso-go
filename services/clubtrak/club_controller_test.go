package clubtrak_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
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

	//TODO: Discuss more efficicent use of User
	u1 := models.User{
		Type:       models.UserTypeStudent,
		Name:       "User 1",
		UnixID:     "u1",
		Visible:    lib.BoolToPtr(true),
		AtWilliams: lib.BoolToPtr(true),
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
