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
	
	o1 := models.Order {
		ItemList: 		"banana, juice, ice-cream"
		Items:   		[]MenuItem{m1, m2, m3} 
		User: 			s1
		UserID: 		s1.UnixID
		PhoneNumber : 	"000-000-0000"
		Notes: 			"very hungry"
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

	//dummy user
	s1 := models.User{
		Type:      models.UserTypeStudent,
		Name:      "Student2",
		UnixID:    "bb",
		ClassYear: lib.IntToPtr(2022),
	}
	
	//dummy menu items
	m1 := models.MenuItem{
		Title:       "bagel",
		Description: "i am round",
		Price:       1.25,
		Available:   true,
	}
	
	m2 := models.MenuItem{
		Title:       "orange",
		Description: "i am my name",
		Price:       0.75,
		Available:   false,
	}
	
	m3 := models.MenuItem{
		Title:       "coffee",
		Description: "i am cold or hot (exclusive-or)",
		Price:       1.75,
		Available:   true,
	}
	
	//dummy order
	o1 := models.Order {
		ItemList: 		"bagel, orange, coffee" 
		Items:   		[]MenuItem{m1, m2, m3} 
		User: 			s1
		UserID: 		s1.UnixID
		PhoneNumber : 	"999-999-9999"
		Notes: 			"too much hw for tomorrow, gonna be a long night..."
	}

	// put dummy order and and dumy user into db
	assert.NoError(db.Create(&s1).Error)
	assert.NoError(db.Create(&o1).Error)
	
	router := utils.SetupRouter(auth.ScopeGoodrichUser, auth.ScopeGoodrichAdmin)
	cfg := utils.SetupConfig()
	logger := zap.S()
	SetupRouter(router, db, cfg, logger)

	// TEST 1: Success testing
	//add contexts
	utils.AddUserContexts(router, s1.ID)

	// perform GET request to "/api/v2/goodrich/:user/orders"
	w, err := utils.DoHTTPReq(router, http.MethodGet, ":user/orders", nil) //is this correct like this??
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	//decode response
	resp := utils.GetHTTPDataResp(a, w.Body.Bytes())

	// make sure response yields same dummy orders

	// TEST 2: Test Failure

	// perform GET request to "/api/v2/goodrich/:user/orders" with invalid userID

	// ensure failure

	return
}

func TestController_ListOrders(t *testing.T) {
	// Setup
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeGoodrichAdmin)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())
	// TEST: Ensure No Errors when Listing All User Orders
	// create three dummy users
	//dummy user 1
	s1 := models.User{
		Type:      models.UserTypeStudent,
		Name:      "Student1",
		UnixID:    "s1",
		ClassYear: lib.IntToPtr(2022),
	}
	//dummy user 2
	s2 := models.User{
		Type:      models.UserTypeStudent,
		Name:      "Student2",
		UnixID:    "s2",
		ClassYear: lib.IntToPtr(2022),
	}
	//dummy user 3
	s3 := models.User{
		Type:      models.UserTypeStudent,
		Name:      "Student3",
		UnixID:    "s3",
		ClassYear: lib.IntToPtr(2022),
	}
	// create two dummy orders for each dummy user
	//dummy menu items
	m1 := models.MenuItem{
		Title:       "Duke Juice",
		Description: "best strawberry lemonade on campus",
		Price:       3.75,
		Available:   false,
	}
	
	m2 := models.MenuItem{
		Title:       "Happy Chick",
		Description: "chicken is abound",
		Price:       25.00,
		Available:   false,
	}
	
	m3 := models.MenuItem{
		Title:       "Gummy Bears",
		Description: "such stretchy bears indeed",
		Price:       2.25,
		Available:   true,
	}
	m4 := models.MenuItem{
		Title:       "Apple Pie Slice",
		Description: "best dessert ever",
		Price:       1.50,
		Available:   true,
	}
	
	//dummy orders 1 and 2 for s1
	o1 := models.Order {
		ItemList: 		m3.Title
		Items:   		[]MenuItem{m3} 
		User: 			s1
		UserID: 		s1.UnixID
		PhoneNumber : 	"010-101-0101"
		Notes: 			"gummy bears have never failed me"
	}
	o2 := models.Order {
		ItemList: 		m2.Title + "," + m2.Title
		Items:   		[]MenuItem{m3} 
		User: 			s1
		UserID: 		s1.UnixID
		PhoneNumber : 	"010-101-0101"
		Notes: 			"i've paid more for less"
	}

	// dummy orders 3 and 4 for s2
	o3 := models.Order {
		ItemList: 		m3.Title + "," + m4.Title
		Items:   		[]MenuItem{m3, m4} 
		User: 			s2
		UserID: 		s2.UnixID
		PhoneNumber : 	"101-010-1010"
		Notes: 			"i've paid more for less"
	}
	o4 := models.Order {
		ItemList: 		m2.Title + "," + m2.Title + "," + m4.Title
		Items:   		[]MenuItem{m2,m2,m4} 
		User: 			s2
		UserID: 		s2.UnixID
		PhoneNumber : 	"101-010-1010"
		Notes: 			"this might as well be 682..."
	}

	// dummy orders 5 and 6 for s3
	o5 := models.Order {
		ItemList: 		m2.Title + "," + m2.Title
		Items:   		[]MenuItem{m2,m2} 
		User: 			s3
		UserID: 		s3.UnixID
		PhoneNumber : 	"101-100-0100"
		Notes: 			"dr-xr--r--"
	}
	o6 := models.Order {
		ItemList: 		m1.Title + "," + m2.Title + m3.Title + "," + m4.Title
		Items:   		[]MenuItem{m1,m2,m3,m4} 
		User: 			s3
		UserID: 		s3.UnixID
		PhoneNumber : 	"101-100-0100"
		Notes: 			"everyone else gets to look but i'm hungry"
	}
	// put dummy users into db
	assert.NoError(db.Create(&s1).Error)
	assert.NoError(db.Create(&s2).Error)
	assert.NoError(db.Create(&s3).Error)
	// put dummy orders into db
	assert.NoError(db.Create(&o1).Error)
	assert.NoError(db.Create(&o2).Error)
	assert.NoError(db.Create(&o3).Error)
	assert.NoError(db.Create(&o4).Error)
	assert.NoError(db.Create(&o5).Error)
	assert.NoError(db.Create(&o6).Error)

	// perform GET request to "/api/v2/goodrich/orders"
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/orders", nil)
	assert.NoError(err)
	// put the bytes of data into a resp object
	resp := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Equal(http.StatusOK, w.Code)
	assert.Nil(resp.Error)
	// ensure correctness of response
	var res []models.Order
	err = json.Unmarshal(resp.Data, &res)
	assert.NoError(err)
	// make sure all fields of o1 match the response's 0th order
	assert.Equal(o1.ItemList, res[0].ItemList)
	assert.Equal(o1.Items, res[0].Items)
	assert.Equal(o1.User, res[0].User)
	assert.Equal(o1.UserID, res[0].UserID)
	assert.Equal(o1.PhoneNumber, res[0].PhoneNumber)
	assert.Equal(o1.Notes, res[0].Notes)
	// make sure all fields of o2 match the response's 1st order
	assert.Equal(o2.ItemList, res[1].ItemList)
	assert.Equal(o2.Items, res[1].Items)
	assert.Equal(o2.User, res[1].User)
	assert.Equal(o2.UserID, res[1].UserID)
	assert.Equal(o2.PhoneNumber, res[1].PhoneNumber)
	assert.Equal(o2.Notes, res[1].Notes)
	// make sure all fields of o3 match the response's 2nd order
	assert.Equal(o3.ItemList, res[2].ItemList)
	assert.Equal(o3.Items, res[2].Items)
	assert.Equal(o3.User, res[2].User)
	assert.Equal(o3.UserID, res[2].UserID)
	assert.Equal(o3.PhoneNumber, res[2].PhoneNumber)
	assert.Equal(o3.Notes, res[2].Notes)
	// make sure all fields of o4 match the response's 3rd order
	assert.Equal(o4.ItemList, res[3].ItemList)
	assert.Equal(o4.Items, res[3].Items)
	assert.Equal(o4.User, res[3].User)
	assert.Equal(o4.UserID, res[3].UserID)
	assert.Equal(o4.PhoneNumber, res[3].PhoneNumber)
	assert.Equal(o4.Notes, res[3].Notes)
	// make sure all fields of o5 match the response's 4th order
	assert.Equal(o5.ItemList, res[4].ItemList)
	assert.Equal(o5.Items, res[4].Items)
	assert.Equal(o5.User, res[4].User)
	assert.Equal(o5.UserID, res[4].UserID)
	assert.Equal(o5.PhoneNumber, res[4].PhoneNumber)
	assert.Equal(o5.Notes, res[4].Notes)
	// make sure all fields of o6 match the response's 5th order
	assert.Equal(o6.ItemList, res[5].ItemList)
	assert.Equal(o6.Items, res[5].Items)
	assert.Equal(o6.User, res[5].User)
	assert.Equal(o6.UserID, res[5].UserID)
	assert.Equal(o6.PhoneNumber, res[5].PhoneNumber)
	assert.Equal(o6.Notes, res[5].Notes)

	assert.Equal(len(res), 6) // make sure res only has 6 elements
	
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
	// TEST 1: Ensure Correctness for Dummy User and Dummy Order
	// make dummy user

	utils.AddOrderContexts(router, o1.ID) 	//Nathan, you might need this function to test GetOrder
											//see lib/test_utils.go for my rough implementation for now

	// put dummy order and user into db

	// create order based on orderParams (put into DB)

	// make GET request to "/api/v2/goodrich/orders/:orderID"

	// ensure correctness

	// TEST 2: Ensure Proper Scope
	// set up another router for a GoodrichUser

	// create second dummy user

	// try to get order of DummyUser1 using DummyUser2 auth and DummyUser1 orderID

	// ensure failure

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

	// create dummy user

	// create dummy orderParams for dummy user

	// put dummy order and user into db

	// TEST 1: Create good update Params for order and update

	// TEST 2: Create bad update Params for order and update

	// TEST 3: Create good update Params for wrong order and update

	return
}
