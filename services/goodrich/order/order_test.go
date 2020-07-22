package order

import (
	"testing"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/models"
)

func TestController_CreateOrder(t *testing.T) {
	// Setup
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeGoodrichUser, auth.ScopeGoodrichAdmin)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	//TEST 0: Create a completely valid order and check if in db
	
	//TEST 1: Create valid order with invalid userID and check if it's successfully created in DB.
	//create dummy user 
	s1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "Student1",
		UnixID: "s1",
	}
	//dummy user creates dummy order
	o1 := models.Order {
		ItemList: 		"banana, juice, ice-cream"
		Items:   		[]MenuItem // FILL WITH MENU ITEMS
		User: 			s1
		UserID: 		s1.UnixID
		PhoneNumber : 	"000-000-0000"
		Notes: 			"very hungry"
	}
	//put dummy order into DB 

	
	//see if dummy order is in db table


	//TEST 2: Create corrupt (not valid userID, missing fields) order and make sure not in db

	//TEST 3: Create order with unavailable MenuItems and see if correctly in db (price, availability)


	return nil
}

func TestController_ListUserOrders(t *testing.T) {
	// Setup
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeGoodrichUser, auth.ScopeGoodrichAdmin)
	cfg := utils.SetupConfig()
	
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	utils.AddUserContexts(router, )
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

func TestController_GetOrderAdmin(t *testing.T) {
	// Setup
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeGoodrichAdmin)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())
	return nil
}

func TestController_UpdateOrder(t *testing.T) {
	// Setup 
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter( auth.ScopeGoodrichAdmin)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())
	return nil
}
