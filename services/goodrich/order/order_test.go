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
		ItemIDs:       []uint{m1.ID, m2.ID, m3.ID},
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
	assert.Equal(orderParams.ItemIDs[0], resp.Items[0].ID) //check if same string item IDs
	assert.Equal(orderParams.PhoneNumber, resp.PhoneNumber)           //check for phone number
	assert.True(resp.Items[0].Available)                              //check if banana is available
	assert.Equal(m1.Price, resp.Items[0].Price)                       //check if banana has same price
	assert.True(resp.Items[1].Available)                              //check if juice is availble
	assert.Equal(m2.Price, resp.Items[1].Price)                       //check if juice has same price
	assert.True(!resp.Items[2].Available)                             //check if ice-cream is navailable
	assert.Equal(m3.Price, resp.Items[2].Price)                       //check if ice-cream has same price

	//TEST 2: Create corrupt (missing fields) order and make sure not in db
	orderParams = CreateOrderParams{
		ItemIDs:       []uint{m1.ID, m2.ID, m3.ID},
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
		Name:      "Student1",
		UnixID:    "s2",
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
	
	m4 := models.MenuItem{
		Title: 		"pizza",
		Description:"a need",
		Price:		6.00,
		Available:	true,
	}

	//dummy order 1
	o1 := models.Order {
		ItemList: 		"bagel, orange, coffee" 
		Items:   		[]MenuItem{m1, m2, m3} 
		User: 			s1
		UserID: 		s1.UnixID
		PhoneNumber : 	"999-999-9999"
		Notes: 			"too much hw for tomorrow, gonna be a long night..."
	}

	//dummy order 2
	o2 := models.Order {
		ItemList: 		"pizza" 
		Items:   		[]MenuItem{m4} 
		User: 			s1
		UserID: 		s1.UnixID
		PhoneNumber : 	"999-999-9999"
		Notes: 			"that's all i need to survive"
	}

	// put dummy order and and dumy user into db
	assert.NoError(db.Create(&s1).Error)
	assert.NoError(db.Create(&o1).Create(&o2).Error)
	
	router := utils.SetupRouter(auth.ScopeGoodrichUser, auth.ScopeGoodrichAdmin)
	cfg := utils.SetupConfig()
	logger := zap.S()
	SetupRouter(router, db, cfg, logger)

	// TEST 1: Success testing
	//add contexts
	utils.AddUserContexts(router, s1.ID)

	// perform GET request to "/api/v2/goodrich/:user/orders"
	w, err := utils.DoHTTPReq(router, http.MethodGet, o1 + "/orders", nil) //is this how to specify ":user" when making request?
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	//decode response
	resp := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(resp.Nil)

	//only two orders
	assert.Equal(len(resp), 2)

	// make sure response yields same dummy orders
	for i,_ := range len(resp){
		assert.Equal(o1.ItemList, res[i].ItemList)
		assert.Equal(o1.Items, res[i].Items)
		assert.Equal(o1.User, res[i].User)
		assert.Equal(o1.UserID, res[i].UserID)
		assert.Equal(o1.PhoneNumber, res[i].PhoneNumber)
		assert.Equal(o1.Notes, res[i].Notes)
	}

	// TEST 2: Test Failure

	//make new dummy user
	s2 := models.User{
		Type:		models.UserTypeStudent,
		Name: 		"Student2",
		UnixID: 	"s2",
		ClassYear: 	lib.IntToPtr(2021),
	}
	
	//setup new user router
	router2 := utils.SetupRouter(auth.ScopeGoodrichUser, auth.ScopeGoodrichAdmin)
	utils.AddUserContexts(router2, s2.ID)
	SetupRouter(router2, db, cfg, logger)

	//put new user in db
	assert.NoError(db.Create(&s2).Error)
	
	// perform GET request to "/api/v2/goodrich/:user/orders" with invalid userID
	w, err := utils.DoHTTPReq(router2, http.MethodGet, ":user/orders", nil)
	
	// ensure failure
	assert.NoError(err)
	assert.Equal(http.StatusNotFound, w.Code)	//what should the http error be here?

	return
}

func TestController_ListOrders(t *testing.T) {
	// Setup
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	
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

	//dummy admin 1
	a1 := {
		Type:      models.UserTypeStudent,
		Name:      "GoodrichAdmin1",
		UnixID:    "a1",
		ClassYear: lib.IntToPtr(2023),
		GoodrichAdmin: true,			//set as goodrich admin
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

	router := utils.SetupRouter(auth.ScopeGoodrichAdmin)
	utils.AddUserContexts(router, a1.ID)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

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
	assert.Equal(http.StatusOK, w.Code)
	// put the bytes of data into a resp object
	resp := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(resp.Error)
	
	var res []models.Order
	err = json.Unmarshal(resp.Data, &res)
	assert.NoError(err)

	/*
	Nathan, I don't think you need lines 383-385 because of the function GetHTTPDataResp that
	you called before. Here is its implementation in lib/utils.go
	
		func GetHTTPDataResp(assert *assert.Assertions, body []byte) APITestResp {
			resp := APITestResp{}
			err := json.Unmarshal(body, &resp)
			assert.NoError(err)

		return resp
	*/

	assert.Equal(len(res), 6) // make sure res only has 6 elements
	
	// ensure correctness of response (replaced with for-loop)
	for i,_ := range len(res) {
		assert.Equal(o1.ItemList, res[i].ItemList)
		assert.Equal(o1.Items, res[i].Items)
		assert.Equal(o1.User, res[i].User)
		assert.Equal(o1.UserID, res[i].UserID)
		assert.Equal(o1.PhoneNumber, res[i].PhoneNumber)
		assert.Equal(o1.Notes, res[i].Notes)
	}
	
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
	

	// create dummy user
	s1 := models.User{
		Type:      models.UserTypeStudent,
		Name:      "Student9",
		UnixID:    "e9",
		ClassYear: lib.IntToPtr(2022),
	}

	a1 := models.User {
		Type:      models.UserTypeStudent,
		Name:      "GoodrichAdmin1",
		UnixID:    "a1",
		ClassYear: lib.IntToPtr(2021),
		GoodrichAdmin: true,			//make a1 a Goodrich Admin 
	}

	//create dummy menu items
	m1 := models.MenuItem{
		Title:       "bagel",
		Description: "i am round",
		Price:       1.25,
		Available:   true,
	}
	m2 := models.MenuItem{
		Title: 		"supreme pizza",
		Description:"a need",
		Price:		6.75,
		Available:	false,
	}

	//create dummy order placed by s1
	o1 := models.Order {
		ItemList: 		m1.Title + "," + m2.Title
		Items:   		[]MenuItem{m1, m2} 
		User: 			s1
		UserID: 		s1.UnixID
		PreferredTime: 	time.Now.Add(1000),
		PhoneNumber : 	"010-909-0330"
		Notes: 			"please bake with <3, I'm quarantined :("
	}

	// put dummy order, user, and items into db
	assert.NoError(db.Create(&s1).Create(&a1).Error)
	assert.NoError(db.Create(&m1).Create(&m2).Error)
	assert.NoError(db.Create(&o1).Error)


	//routing setup
	router := utils.SetupRouter(auth.ScopeGoodrichAdmin)
	utils.AddUserContexts(router, a1.ID)
	cfg := utils.SetupConfig()
	logger := zap.S()
	SetupRouter(router, db, cfg, logger)

	// TEST 1: Create good update Params for order and update

	//create updateParams
	updateParams := models.UpdateOrderParams {
		OrderStatus:   	models.OrderStatusInProgress 	//should this be OrderStatusDenied because no pizza?
		AdminNotes:    	"sorry, no pizza's left for you :/",                     
		EstimatedTime 	time.Now.Add(10000),     		//check to see later if time gets updated          
		Items         	[]models.MenuItem{m1},       
	}

	//do http PATH request with a1 as admin (in user contexts)
	updateOrderData, err := json.Marshal(&updateParams)
	assert.NoError(err)
	w, err := utils.DoHTTPReq(router, http.MethodPatch, "/orders/" + o1.ID, bytes.NewBuffer(updateOrderData))		 //is this how to specify correctly orderID in URL? 
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	//decode updated entry
	resp := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(resp.Nil)

	//check for correctness
	assert.Equal(s1.ID, resp.UserID)                                  	//check if same userID

	total_price float64 = 0
	for i,_ := range len(resp.Items){
		assert.Equal(updateParams.Items[i].ID, resp.Items[i].ID)		//items have same itemID
		assert.True(resp.Items[i].Available)							//all items are available after admin update
		assert.Equal(updateParams.Items[i].Price, resp.Items[i].Price)	//each item has same price
		total_price += updateParams.Items[i].Price           			         
	}
	assert.Equal(total_price, resp.TotalPrice, )						//total price is updated for only available items
	assert.Equal(updateParams.EstimatedTime, resp.EstimatedTime)		//estimated time is updated
	assert.Equal(updateParams.OrderStatus, resp.OrderStatus)			//order status is updated from placed to in progress
	assert.Equal(updateParams.AdminNotes, respAdminNotes) 				//admin notes are the same too
	
	
	//TODO:
	// TEST 2: Create bad update Params for order and update
	// TEST 3: Create good update Params for wrong order and update

	return
}
