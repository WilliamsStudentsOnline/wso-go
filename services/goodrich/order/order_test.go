package order

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestController_CreateOrder(t *testing.T) {
	// Setup
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	//TEST 1: Create valid order with invalid userID and check if it's successfully created in DB.
	//create dummy user
	s1 := models.User{
		Type:      models.UserTypeStudent,
		Name:      "Student1",
		UnixID:    "s1",
		ClassYear: lib.IntToPtr(2023),
	}

	m1 := models.MenuItem{
		Title:       "banana",
		Description: "i am healthy",
		Price:       1.75,
		Available:   true,
	}

	m2 := models.MenuItem{
		Title:       "juice",
		Description: "i am slurpy",
		Price:       2.00,
		Available:   true,
	}

	m3 := models.MenuItem{
		Title:       "ice-cream",
		Description: "i am cold",
		Price:       2.25,
		Available:   false,
	}

	orderParams := CreateOrderParams{
		ItemIDs:       []uint{m1.BaseSchema.ID, m2.BaseSchema.ID, m3.BaseSchema.ID},
		PhoneNumber:   "000-000-0000",
		PreferredTime: time.Now(),
		Notes:         "very hungry",
	}

	//put dummy user, dummy menu-items into DB
	assert.NoError(db.Create(&s1).Error)
	assert.NoError(db.Create(&m1).Create(&m2).Create(&m3).Error)

	//Setup router
	router := utils.SetupRouter(auth.ScopeGoodrichUser, auth.ScopeGoodrichAdmin)
	utils.AddUserContexts(router, s1.ID) //do we need this line ?
	cfg := utils.SetupConfig()
	logger := zap.S()
	SetupRouter(router, db, cfg, logger)

	//do http POST request to create order orderParams
	orderData, err := json.Marshal(&orderParams)
	assert.NoError(err)
	o, err := utils.DoHTTPReq(router, http.MethodPost, "/orders", bytes.NewBuffer(orderData))
	assert.NoError(err)

	//should be okay if nothing went wrong
	assert.Equal(http.StatusOK, o.Code)

	//get dummy order from table
	resp := models.Order{}
	err = db.Where(models.Order{UserID: s1.ID}).First(&resp).Error
	assert.NoError(err)

	//Check for correctness
	assert.Len(resp, 1)
	assert.Equal(s1.ID, resp.UserID)                                  //check if same userID
	assert.Equal(orderParams.ItemIDs[0], resp.Items[0].BaseSchema.ID) //check if same string item IDs
	assert.Equal(orderParams.PhoneNumber, resp.PhoneNumber)           //check for phone number
	assert.True(resp.Items[0].Available)                              //check if banana is available
	assert.Equal(m1.Price, resp.Items[0].Price)                       //check if banana has same price
	assert.True(resp.Items[1].Available)                              //check if juice is availble
	assert.Equal(m2.Price, resp.Items[1].Price)                       //check if juice has same price
	assert.True(!resp.Items[2].Available)                             //check if ice-cream is navailable
	assert.Equal(m3.Price, resp.Items[2].Price)                       //check if ice-cream has same price

	//TEST 2: Create corrupt (missing fields) order and make sure not in db
	orderParams = CreateOrderParams{
		ItemIDs:       []uint{m1.BaseSchema.ID, m2.BaseSchema.ID, m3.BaseSchema.ID},
		PhoneNumber:   "",
		PreferredTime: time.Now(),
		Notes:         "",
	}

	orderData, err = json.Marshal(&orderParams)
	assert.NoError(err)
	o, err = utils.DoHTTPReq(router, http.MethodPost, "/orders", bytes.NewBuffer(orderData))
	assert.NoError(err) //should this be deleted?
	apiError := lib.ErrorGoodrichMissingPhoneNumber
	assert.Equal(apiError.HTTPCode, o.Code)

	return
}

func TestController_ListUserOrders(t *testing.T) {
	// Setup
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeGoodrichUser, auth.ScopeGoodrichAdmin)
	cfg := utils.SetupConfig()
	logger := zap.S()
	SetupRouter(router, db, cfg, logger)

	return
}

func TestController_ListOrders(t *testing.T) {
	// Setup
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeGoodrichAdmin)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	return
}

func TestController_GetOrder(t *testing.T) {
	// Setup
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeGoodrichUser, auth.ScopeGoodrichAdmin)
	cfg := utils.SetupConfig()
	logger := zap.S()
	SetupRouter(router, db, cfg, logger)

	return
}

func TestController_UpdateOrder(t *testing.T) {
	// Setup
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeGoodrichUser, auth.ScopeGoodrichAdmin)
	cfg := utils.SetupConfig()
	logger := zap.S()
	SetupRouter(router, db, cfg, logger)

	return
}
