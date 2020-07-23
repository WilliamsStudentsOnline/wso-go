package order

import (
	"encoding/json"
	"net/http"
	"testing"

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
		Type:   models.UserTypeStudent,
		Name:   "Student1",
		UnixID: "s1",
		ClassYear: lib.IntToPtr(2023)
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

	//dummy user creates dummy order
	o1 := models.Order{
		ItemList:    "banana, juice, ice-cream",
		Items:       [...]MenuItem{m1, m2, m3}, // FILL WITH MENU ITEMS
		User:        &s1,
		UserID:      s1.ID,
		PhoneNumber: "000-000-0000",
		Notes:       "very hungry",
	}

	//put dummy user, dummy menu-items, and dummy order into DB
	assert.NoError(db.Create(&s1).Error)
	assert.NoError(db.Create(&m1).Create(&m2).Create(&m3).Error)
	assert.NoError(db.Create(&o1).Error)

	//Setup router
	router := utils.SetupRouter(auth.ScopeGoodrichUser, auth.ScopeGoodrichAdmin)
	utils.AddUserContexts(router, s1.ID) //do we need this line
	cfg := utils.SetupConfig()
	logger := zap.S()
	SetupRouter(router, db, cfg, logger)

	//get dummy order from table
	o, err := utils.DoHTTPReq(router, http.MethodGet, "/orders", nil) //what to put as URL (3rd) argument indestad of '/orders'?
	assert.NoError(err)

	//Status should be okay
	assert.Equal(http.StatusOK, o.Code)

	//Decode response
	respData := utils.GetHTTPDataResp(assert, o.Body.Bytes())
	assert.Nil(respData.Error)
	var resp []models.Order
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	//Check for correctness
	assert.Len(resp, 1)
	assert.Equal(o1.UserID, resp[0].UserID)                 //check if same userID
	assert.Equal(o1.ItemList, resp[0].ItemList)             //check if same string item list
	assert.Equal(o1.PhoneNumber, resp[0].PhoneNumber)       //check if same phone number
	assert.True(resp[0].Items[0].Available)                 //check if banana is available
	assert.Equal(o1.Items[1].Price, resp[0].Items[1].Price) //check if juice has same price
	assert.True(!resp[0].Items[2].Available)                //check if ice-cream is unavailable

	//TEST 2: Create corrupt (missing fields) order and make sure not in db
	//TEST 3: Create order with unavailable MenuItems and see if correctly in db (price, availability)

	return nil
}

func TestController_ListUserOrders(t *testing.T) {
	// Setup
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeGoodrichUser, auth.ScopeGoodrichAdmin)
	cfg := utils.SetupConfig()
	logger := zap.S()
	SetupRouter(router, db, cfg, logger)

	return nil
}

func TestController_ListOrders(t *testing.T) {
	// Setup
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeGoodrichAdmin)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())
	return nil
}

func TestController_GetOrder(t *testing.T) {
	// Setup
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeGoodrichUser)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())
	return nil
}

func TestController_UpdateOrder(t *testing.T) {
	// Setup
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeGoodrichAdmin)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())
	return nil
}
