package goodrich

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestController_ListMenu(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeGoodrich)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	m1 := models.GoodrichMenuItem{
		Title:       "Bagel",
		Description: "It's a bagel",
		Price:       1.69,
		Available:   true,
	}
	m2 := models.GoodrichMenuItem{
		Title:       "Coffee",
		Description: "you drink it",
		Price:       3.41,
		Available:   true,
	}
	m3 := models.GoodrichMenuItem{
		Title:       "Avocado",
		Description: "a rare specialty",
		Price:       4.29,
		Available:   false,
	}
	assert.NoError(db.Create(&m1).Create(&m2).Create(&m3).Error)

	/* Test 1: works normally */
	// Get test menu
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/menu", nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp []models.GoodrichMenuItem
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct menu items
	assert.Len(resp, 2)
	assert.Equal(m1.Title, resp[0].Title)
	assert.Equal(m2.Title, resp[1].Title)

	/* Test 2: ?all=true */
	// Get test menu
	w, err = utils.DoHTTPReq(router, http.MethodGet, "/menu?all=true", nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	resp = []models.GoodrichMenuItem{}
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct menu items
	assert.Len(resp, 3)
	assert.Equal(m3.Title, resp[0].Title) // (A)vocado
	assert.Equal(m1.Title, resp[1].Title) // (B)agel
	assert.Equal(m2.Title, resp[2].Title) // (C)offee
}

func TestController_CreateMenuItem(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeGoodrich, auth.ScopeGoodrichManager)
	cfg := utils.SetupConfig()

	u1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "User 1",
		UnixID: "u1",
	}
	u2 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "User 2",
		UnixID: "u2",
	}
	assert.NoError(db.Create(&u1).Create(&u2).Error)

	utils.AddUserContexts(router, u1.ID)
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	/* Create item with missing data (expect failure) */
	apiErr := lib.ErrorRequestDataValidationFailed
	w, err := utils.DoHTTPReq(router, http.MethodPost, "/menu", bytes.NewBufferString(`{"price": 4}`))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Create item with bad permissions (expect failure) */
	params := CreateMenuItemParams{
		Title:       "Hot Cocoa",
		Description: "It's tasty and warm",
		Price:       0.69,
		Available:   true,
	}
	// Setup bad student router
	r1 := utils.SetupRouter(auth.ScopeGoodrich)
	utils.AddUserContexts(r1, u2.ID)
	SetupRouter(r1, db, cfg, zaptest.NewLogger(t).Sugar())

	apiErr = lib.ErrorNoScopeAuthorization
	paramsData, err := json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(r1, http.MethodPost, "/menu", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Create item (expect success) */
	params = CreateMenuItemParams{
		Title:       "Hot Cocoa",
		Description: "It's tasty and warm",
		Price:       17.38,
		Available:   true,
	}
	paramsData, err = json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(router, http.MethodPost, "/menu", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(http.StatusCreated, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.GoodrichMenuItem
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Assert good response
	assert.NotZero(resp.ID)
	assert.Equal(params.Title, resp.Title)
	assert.Equal(params.Price, resp.Price)
	assert.Equal(params.Description, resp.Description)
	assert.Equal(params.Available, resp.Available)

	// Assert found in db
	var count int
	assert.NoError(db.Model(&models.GoodrichMenuItem{}).Where("id = ?", resp.ID).Count(&count).Error)
	assert.Equal(1, count)
}

func TestController_UpdateMenuItem(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeGoodrich, auth.ScopeGoodrichManager)
	cfg := utils.SetupConfig()

	u1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "User 1",
		UnixID: "u1",
	}
	u2 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "User 2",
		UnixID: "u2",
	}
	assert.NoError(db.Create(&u1).Create(&u2).Error)
	m1 := models.GoodrichMenuItem{
		Title:       "Bagel",
		Description: "It's a bagel",
		Price:       1.69,
		Available:   true,
	}
	assert.NoError(db.Create(&m1).Error)

	utils.AddUserContexts(router, u1.ID)
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	/* Update item with bad permissions (expect failure) */
	params := UpdateMenuItemParams{
		Price:     lib.Float64ToPtr(3.25),
		Available: lib.BoolToPtr(false),
	}
	// Setup bad student router
	r1 := utils.SetupRouter(auth.ScopeGoodrich)
	utils.AddUserContexts(r1, u2.ID)
	SetupRouter(r1, db, cfg, zaptest.NewLogger(t).Sugar())

	apiErr := lib.ErrorNoScopeAuthorization
	paramsData, err := json.Marshal(&params)
	assert.NoError(err)
	w, err := utils.DoHTTPReq(r1, http.MethodPatch, "/menu/1", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Update item (expect success) */
	params = UpdateMenuItemParams{
		Price:     lib.Float64ToPtr(3.25),
		Available: lib.BoolToPtr(false),
	}
	paramsData, err = json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(router, http.MethodPatch, "/menu/1", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.GoodrichMenuItem
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Assert good response
	assert.Equal(m1.ID, resp.ID)
	assert.Equal(*params.Price, resp.Price)
	assert.Equal(*params.Available, resp.Available)

	// Assert updated in db
	var respDB models.GoodrichMenuItem
	assert.NoError(db.First(&respDB, m1.ID).Error)
	assert.Equal(*params.Price, respDB.Price)
	assert.Equal(m1.Description, respDB.Description)
	assert.Equal(m1.Title, respDB.Title)
}
