package bulletin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"

	"github.com/gin-gonic/gin"
	testify "github.com/stretchr/testify/assert"
)

var (
	u1 = models.User{
		Name:        "Test 1",
		UnixID:      "u1",
		ClassYear:   lib.IntToPtr(3),
		Visible:     true,
		AtWilliams:  true,
		DormVisible: true,
		OffCycle:    false,
	}
	u2 = models.User{
		Name:        "Test 2",
		UnixID:      "u2",
		ClassYear:   lib.IntToPtr(2),
		Visible:     true,
		AtWilliams:  true,
		DormVisible: true,
		OffCycle:    false,
	}

	bulletin1 = models.Bulletin{
		BaseSchema: models.BaseSchema{
			ID: 1,
		},
		Title: "Test",
		Body:  "Hello WSO!",
		Type:  "lostAndFound",
		User:  &u1,
	}
	bulletin2 = models.Bulletin{
		BaseSchema: models.BaseSchema{
			ID: 2,
		},
		Title: "Test",
		Body:  "Hello WSO!",
		Type:  "job",
		User:  &u1,
	}
	bulletin3 = models.Bulletin{
		BaseSchema: models.BaseSchema{
			ID: 3,
		},
		Title: "Test",
		Body:  "Hello WSO!",
		Type:  "lostAndFound",
		User:  &u2,
	}
)

func TestController_CreateBulletin(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	// Insert test user into db
	assert.NoError(db.Create(&u1).Error)

	router := gin.Default()
	utils.AddUserContexts(router, u1.ID)
	SetupRouter(router, db)

	// First, run tests on validations

	// Test 1: Error on missing bulletin params
	params := BulletinCreateParams{Title: "hi"}
	// Expect ErrorMissingBulletinParams
	createBulletinExpectError(assert, router, params, 400, 1601)

	// Test 2: Missing start/end dates
	params = BulletinCreateParams{Title: "hi", Body: "WSO"}
	// Expect ErrorMissingBulletinDates
	createBulletinExpectError(assert, router, params, 400, 1602)

	// Test 3: Invalid start/end dates
	params = BulletinCreateParams{Title: "hi", Body: "WSO",
		StartDate: time.Date(2019, time.November, 10, 23, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2019, time.October, 10, 23, 0, 0, 0, time.UTC)}
	// Expect ErrorBulletinInvalidDates
	createBulletinExpectError(assert, router, params, 400, 1603)

	// Test 4: InvalidType
	params = BulletinCreateParams{Title: "hi", Body: "WSO", Type: "wrongType",
		StartDate: time.Date(2019, time.November, 10, 23, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2019, time.December, 10, 23, 0, 0, 0, time.UTC)}
	// Expect ErrorBulletinInvalidDates
	createBulletinExpectError(assert, router, params, 400, 1605)

	// Test 5: Success
	params = BulletinCreateParams{Title: "hi", Body: "WSO", Type: "lostAndFound",
		StartDate: time.Date(2019, time.November, 10, 23, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2019, time.December, 10, 23, 0, 0, 0, time.UTC)}
	// Expect Success
	resBulletin := createBulletinExpectSuccess(assert, router, params)

	// Assert these values:
	assert.Equal(params.Body, resBulletin.Body)
	assert.Equal(params.Title, resBulletin.Title)
	assert.Equal(params.Type, resBulletin.Type)
	assert.Equal(params.StartDate, resBulletin.StartDate)
	assert.Equal(params.EndDate, resBulletin.EndDate)
}

// TODO: Update with lib APIError
func createBulletinExpectError(assert *testify.Assertions, router *gin.Engine, params BulletinCreateParams, errHTTPCode, errCode int) {
	paramsData, err := json.Marshal(&params)
	assert.NoError(err)

	// Get bad survey (expect failure)
	w, err := utils.DoHTTPReq(router, http.MethodPost, "/bulletins", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(errHTTPCode, w.Code)
	// Assert correct error
	respErrCode := utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode
	assert.Equal(errCode, respErrCode)
}

func createBulletinExpectSuccess(assert *testify.Assertions, router *gin.Engine, params BulletinCreateParams) models.Bulletin {
	paramsData, err := json.Marshal(&params)
	assert.NoError(err)

	// Get bad survey (expect failure)
	w, err := utils.DoHTTPReq(router, http.MethodPost, "/bulletins", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(http.StatusCreated, w.Code)

	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)

	var resp models.Bulletin
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	return resp
}

func TestController_GetBulletinByID(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := gin.Default()
	SetupRouter(router, db)

	// Insert test bulletin into db
	err := db.FirstOrCreate(&bulletin1).Error
	assert.NoError(err)

	// Get test bulletin
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/1", nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	respBulletin := models.Bulletin{}
	err = json.Unmarshal(respData.Data, &respBulletin)
	assert.NoError(err)

	// Check if correct bulletin
	assert.Equal(bulletin1.ID, respBulletin.ID)
	assert.Equal(bulletin1.Title, respBulletin.Title)
}

func TestController_FetchAllBulletins(t *testing.T) {

	bulletins := []models.Bulletin{
		bulletin1, bulletin2, bulletin3,
	}

	testFetch := func(url string, expected []models.Bulletin) func(*testing.T) {
		return func(t *testing.T) {
			// Setup (can copy and paste this basically)
			assert := testify.New(t)
			db := utils.SetupServiceTest(assert)
			router := gin.Default()
			SetupRouter(router, db)

			// Insert test bulletin into db
			for i := 0; i < len(bulletins); i++ {
				err := db.FirstOrCreate(&bulletins[i]).Error
				assert.NoError(err)
			}

			// Get test bulletins
			w, err := utils.DoHTTPReq(router, http.MethodGet, url, nil)
			assert.NoError(err)

			// Status is okay
			assert.Equal(http.StatusOK, w.Code)

			// Decode response
			respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
			var respBulletin = []models.Bulletin{}
			err = json.Unmarshal(respData.Data, &respBulletin)
			assert.NoError(err)

			// Check if length of bulletin obtained matches expectations
			assert.Equal(len(respBulletin), len(expected))

			// Check if all the bulletins titles and bodies were obtained correctly
			for i := 0; i < len(respBulletin); i++ {
				assert.Equal(expected[i].ID, respBulletin[i].ID)
				assert.Equal(expected[i].Title, respBulletin[i].Title)
				assert.Equal(expected[i].Body, respBulletin[i].Body)
			}
		}
	}

	t.Run("Fetches all Bulletins", testFetch("/", []models.Bulletin{bulletin1, bulletin2, bulletin3}))
	t.Run("Fetches only the Lost and Found Bulletins", testFetch("/?type=lostAndFound",
		[]models.Bulletin{bulletin1, bulletin3}))

}

// Todo: Add test for admin scope when implemented
func TestController_DeleteBulletin(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := gin.Default()

	// Insert test user into db
	assert.NoError(db.Create(&u1).Error)

	utils.AddUserContexts(router, u1.ID)
	SetupRouter(router, db)

	// Insert test bulletin into db
	err := db.FirstOrCreate(&bulletin1).FirstOrCreate(&bulletin3).Error
	assert.NoError(err)

	// Get test bulletin
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/1", nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	respBulletin := models.Bulletin{}
	err = json.Unmarshal(respData.Data, &respBulletin)
	assert.NoError(err)

	// Check if correct bulletin
	assert.Equal(bulletin1.ID, respBulletin.ID)
	assert.Equal(bulletin1.Title, respBulletin.Title)

	// Delete test bulletin
	w, err = utils.DoHTTPReq(router, http.MethodDelete, "/1", nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Get test bulletin
	w, err = utils.DoHTTPReq(router, http.MethodGet, "/1", nil)
	assert.NoError(err)

	// Status not found because Bulletin is deleted
	assert.Equal(http.StatusNotFound, w.Code)

	// Get test bulletin
	w, err = utils.DoHTTPReq(router, http.MethodGet, "/3", nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	respBulletin = models.Bulletin{}
	err = json.Unmarshal(respData.Data, &respBulletin)
	assert.NoError(err)

	// Check if correct bulletin
	assert.Equal(bulletin3.ID, respBulletin.ID)
	assert.Equal(bulletin3.Title, respBulletin.Title)

	// Attempt to delete test bulletin
	w, err = utils.DoHTTPReq(router, http.MethodDelete, "/3", nil)
	assert.NoError(err)

	// Expect Forbidden since the user is incorrect
	assert.Equal(http.StatusForbidden, w.Code)

	// Get test bulletin
	w, err = utils.DoHTTPReq(router, http.MethodGet, "/3", nil)
	assert.NoError(err)

	// Status found because Bulletin was not deleted
	assert.Equal(http.StatusOK, w.Code)
}

func TestController_UpdateBulletin(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	// Insert test user into db
	assert.NoError(db.Create(&u1).Error)

	router := gin.Default()
	utils.AddUserContexts(router, u1.ID)
	SetupRouter(router, db)

	// Insert test bulletin into db
	err := db.FirstOrCreate(&bulletin1).Error
	assert.NoError(err)

	// Get test bulletin
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/1", nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	respBulletin := models.Bulletin{}
	err = json.Unmarshal(respData.Data, &respBulletin)
	assert.NoError(err)

	// Check if correct bulletin
	assert.Equal(bulletin1.ID, respBulletin.ID)
	assert.Equal(bulletin1.Title, respBulletin.Title)

	// Construct update params
	var jsonStr = []byte(`{"title":"HelloWSO.","type":"job"}`)

	// Update test bulletin
	w, err = utils.DoHTTPReq(router, http.MethodPut, "/1", bytes.NewBuffer(jsonStr))
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Get test bulletin
	w, err = utils.DoHTTPReq(router, http.MethodGet, "/1", nil)
	assert.NoError(err)

	// Decode response
	respData = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	respBulletin = models.Bulletin{}
	err = json.Unmarshal(respData.Data, &respBulletin)
	assert.NoError(err)

	// Check if bulletin title is updated
	assert.Equal("HelloWSO.", respBulletin.Title)

	// Check if bulletin type is not updated
	assert.Equal(bulletin1.Type, respBulletin.Type)
}
