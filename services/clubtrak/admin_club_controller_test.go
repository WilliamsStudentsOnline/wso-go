package clubtrak_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/clubtrak"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestAdminController_CreateClub(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeAdminAll)
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
		ClubAdminID:        u1.ID,
	}

	//Process Params
	paramsData, err := json.Marshal(&params)
	assert.NoError(err)

	// Test HTTP Request
	w, err := utils.DoHTTPReq(router, http.MethodPost, "/admin/clubs", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(http.StatusCreated, w.Code)

	//Missing Fields Case
	params2 := ClubParams{
		Name:               "WSO",
		Subscribers:        30,
		MeetingDescription: "Sundays at 1pm in Wach B11",
	}

	//Process Params
	paramsData2, err := json.Marshal(&params2)
	assert.NoError(err)

	// Test HTTP Request
	w2, err2 := utils.DoHTTPReq(router, http.MethodPost, "/admin/clubs", bytes.NewBuffer(paramsData2))
	assert.NoError(err2)
	assert.Equal(http.StatusCreated, w2.Code)

}

func TestAdminController_GetAllClubs(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeAdminAll)
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
		ClubAdminID:        u1.ID,
		ClubDescription:    "Write code for WSO!",
		MeetingDescription: "Sundays 1-3pm in Wach B11",
	}

	assert.NoError(db.Create(&u1).Create(&c1).Error)

	c2 := models.Club{
		Name:               "Octet",
		Category:           "Performing Arts",
		ClubAdminID:        u1.ID,
		ClubDescription:    "Sing without instruments!",
		MeetingDescription: "Mondays 1-3pm in Bernhardt",
	}

	assert.NoError(db.Create(&c2).Error)

	c3 := models.Club{
		Name:               "Kusika",
		Category:           "Performing Arts",
		ClubAdminID:        u1.ID,
		ClubDescription:    "Learn African Drumming and Dance from across the diaspora!",
		MeetingDescription: "Mon Wed Fri 4-6pm in '62 Center Dance Studio",
	}

	assert.NoError(db.Create(&c3).Error)

	// Test Endpoint
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/admin/clubs", nil)
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

func TestAdminController_DeleteClubs(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeAdminAll)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	//Create Some Clubs
	u1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "User 1",
		UnixID: "u1",
	}

	c1 := models.Club{
		Name:               "WSO",
		Category:           "STEM",
		ClubAdminID:        u1.ID,
		ClubDescription:    "Write code for WSO!",
		MeetingDescription: "Sundays 1-3pm in Wach B11",
	}

	assert.NoError(db.Create(&u1).Create(&c1).Error)

	c2 := models.Club{
		Name:               "Octet",
		Category:           "Performing Arts",
		ClubAdminID:        u1.ID,
		ClubDescription:    "Sing without instruments!",
		MeetingDescription: "Mondays 1-3pm in Bernhardt",
	}

	assert.NoError(db.Create(&c2).Error)

	c3 := models.Club{
		Name:               "Kusika",
		Category:           "Performing Arts",
		ClubAdminID:        u1.ID,
		ClubDescription:    "Learn African Drumming and Dance from across the diaspora!",
		MeetingDescription: "Mon Wed Fri 4-6pm in '62 Center Dance Studio",
	}

	assert.NoError(db.Create(&c3).Error)

	//Grab the ID's of every club and convert to strings
	ID1 := c1.ID
	ID1str := strconv.FormatUint(uint64(ID1), 10)

	ID2 := c2.ID
	ID2str := strconv.FormatUint(uint64(ID2), 10)

	ID3 := c3.ID
	ID3str := strconv.FormatUint(uint64(ID3), 10)

	// Test Deleting c1 (WSO)
	w1, err := utils.DoHTTPReq(router, http.MethodDelete, "/admin/clubs/"+ID1str, nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w1.Code)

	// Decode response
	respData1 := utils.GetHTTPDataResp(assert, w1.Body.Bytes())
	assert.Nil(respData1.Error)
	var resp1 models.Club
	assert.NoError(json.Unmarshal(respData1.Data, &resp1))

	//Check that the clubs have been deleted with get request
	w2, err := utils.DoHTTPReq(router, http.MethodGet, "/admin/clubs", nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w2.Code)

	respData2 := utils.GetHTTPDataResp(assert, w2.Body.Bytes())
	assert.Nil(respData2.Error)
	var resp2 []models.Club
	assert.NoError(json.Unmarshal(respData2.Data, &resp2))

	//Number of clubs should be 2
	assert.Len(resp2, 2)

	// Test Deleting c2 (Octet)
	w3, err := utils.DoHTTPReq(router, http.MethodDelete, "/admin/clubs/"+ID2str, nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w3.Code)

	// Decode response
	respData3 := utils.GetHTTPDataResp(assert, w3.Body.Bytes())
	assert.Nil(respData3.Error)
	var resp3 models.Club
	assert.NoError(json.Unmarshal(respData3.Data, &resp3))

	//Check that the clubs have been deleted with get request
	w4, err := utils.DoHTTPReq(router, http.MethodGet, "/admin/clubs", nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w4.Code)

	respData4 := utils.GetHTTPDataResp(assert, w4.Body.Bytes())
	assert.Nil(respData1.Error)
	var resp4 []models.Club
	assert.NoError(json.Unmarshal(respData4.Data, &resp4))

	//Number of clubs should be 1
	assert.Len(resp4, 1)

	// Test Deleting c3 (Kusika)
	w5, err := utils.DoHTTPReq(router, http.MethodDelete, "/admin/clubs/"+ID3str, nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w5.Code)

	// Decode response
	respData5 := utils.GetHTTPDataResp(assert, w5.Body.Bytes())
	assert.Nil(respData3.Error)
	var resp5 models.Club
	assert.NoError(json.Unmarshal(respData5.Data, &resp5))

	//Check that the clubs have been deleted with get request
	w6, err := utils.DoHTTPReq(router, http.MethodGet, "/admin/clubs", nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w6.Code)

	respData6 := utils.GetHTTPDataResp(assert, w6.Body.Bytes())
	assert.Nil(respData6.Error)
	var resp6 []models.Club
	assert.NoError(json.Unmarshal(respData6.Data, &resp6))

	//Number of clubs should be zero
	assert.Len(resp6, 0)

}

func TestAdminController_UpdateClub(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeAdminAll)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	//Create Some Clubs
	u1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "User 1",
		UnixID: "u1",
	}

	c1 := models.Club{
		Name:               "WSO",
		Category:           "STEM",
		ClubAdminID:        u1.ID,
		ClubDescription:    "Write code for WSO!",
		MeetingDescription: "Sundays 1-3pm in Wach B11",
		Subscribers:        5,
		ContactEmail:       "coolhuman24@williams.edu",
		ContactPhoneNumber: "(123)-456-7890",
		Website:            "acoolwebsitethathopefullyisntreal.com",
	}

	assert.NoError(db.Create(&u1).Create(&c1).Error)

	//Create Updated Club Params
	params := ClubUpdateParams{
		Name:               "WSO",
		MeetingDescription: "Sundays at 1pm in Wach B12",
		Website:            "adifferentcoolwebistethathopefullyisntreal.com",
	}

	//Process Params
	paramsData, err := json.Marshal(&params)
	assert.NoError(err)

	// Test HTTP Request
	ID := c1.ID
	IDstr := strconv.FormatUint(uint64(ID), 10)
	w, err := utils.DoHTTPReq(router, http.MethodPatch, "/admin/clubs/"+IDstr, bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	//Use a get request to check that the clubs parameters have been updates
	w2, err := utils.DoHTTPReq(router, http.MethodGet, "/admin/clubs", nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w2.Code)

	respData2 := utils.GetHTTPDataResp(assert, w2.Body.Bytes())
	assert.Nil(respData2.Error)
	var resp2 []models.Club
	assert.NoError(json.Unmarshal(respData2.Data, &resp2))

	//Assert Proper updated values
	assert.Equal(resp2[0].MeetingDescription, params.MeetingDescription)
	assert.Equal(resp2[0].Website, params.Website)

}

func TestClubtrak_AdminAccessControl(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	cfg := utils.SetupConfig()

	//Setup routers for different roles
	userRouter := utils.SetupRouter(auth.ScopeUsers)
	adminRouter := utils.SetupRouter(auth.ScopeAdminAll)

	SetupRouter(userRouter, db, cfg, zaptest.NewLogger(t).Sugar())
	SetupRouter(adminRouter, db, cfg, zaptest.NewLogger(t).Sugar())

	u1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "User 1",
		UnixID: "u1",
	}
	assert.NoError(db.Create(&u1).Error)

	utils.AddUserContexts(userRouter, u1.ID)
	utils.AddUserContexts(adminRouter, u1.ID)

	params := ClubParams{
		Name:               "Test Club",
		Subscribers:        10,
		MeetingDescription: "Fridays 5-6pm",
		ClubAdminID:        u1.ID,
	}
	paramsData, err := json.Marshal(&params)
	assert.NoError(err)

	//Regular user tries to POST /admin/clubs
	w, err := utils.DoHTTPReq(userRouter, http.MethodPost, "/admin/clubs", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(http.StatusForbidden, w.Code)

	//Admin tries to POST /admin/clubs
	w2, err2 := utils.DoHTTPReq(adminRouter, http.MethodPost, "/admin/clubs", bytes.NewBuffer(paramsData))
	assert.NoError(err2)
	assert.Equal(http.StatusCreated, w2.Code)

	//Regular user GET /clubs
	w4, err4 := utils.DoHTTPReq(userRouter, http.MethodGet, "/clubs", nil)
	assert.NoError(err4)
	assert.Equal(http.StatusOK, w4.Code)
}
