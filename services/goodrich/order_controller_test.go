package goodrich

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestController_ListUserOrders(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeGoodrich)
	cfg := utils.SetupConfig()

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

	o1 := models.GoodrichOrder{
		Status:        models.GoodrichOrderStatusPlaced,
		PhoneNumber:   "4131112222",
		TimeSlot:      "08:40",
		Date:          time.Now().Format(DateFormat),
		Notes:         "hi",
		TotalPrice:    5.1,
		ComboDeal:     lib.BoolToPtr(false),
		PaymentMethod: models.GoodrichPaymentMethodPoints,
		IDNumber:      lib.StrToPtr("5552223"),
		ItemList:      `[{"id": 1},{"id": 2}]`,
		UserID:        u1.ID,
		User:          &u1,
	}
	o2 := models.GoodrichOrder{
		Status:        models.GoodrichOrderStatusReady,
		PhoneNumber:   "4131112222",
		TimeSlot:      "08:30",
		Date:          time.Now().Format(DateFormat),
		Notes:         "old order",
		TotalPrice:    5,
		ComboDeal:     lib.BoolToPtr(true),
		PaymentMethod: models.GoodrichPaymentMethodSwipe,
		IDNumber:      lib.StrToPtr("5552223"),
		ItemList:      `[{"id": 1},{"id": 3}]`,
		UserID:        u1.ID,
		User:          &u1,
	}
	o3 := models.GoodrichOrder{
		Status:        models.GoodrichOrderStatusPaid,
		PhoneNumber:   "4135556666",
		TimeSlot:      "11:20",
		Date:          time.Now().Format(DateFormat),
		Notes:         "other order",
		TotalPrice:    7.7,
		ComboDeal:     lib.BoolToPtr(false),
		PaymentMethod: models.GoodrichPaymentMethodCreditCard,
		ItemList:      `[{"id": 2},{"id": 3}]`,
		UserID:        u2.ID,
		User:          &u2,
	}
	assert.NoError(db.Create(&o1).Create(&o2).Create(&o3).Error)

	utils.AddUserContexts(router, u1.ID)
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	/* Test 1: works normally */
	// Get test menu
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/user/orders", nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp []models.GoodrichOrder
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct orders
	assert.Len(resp, 2)
	assert.Equal(o1.Notes, resp[0].Notes)
	assert.Equal(o1.ID, resp[0].ID)
	assert.Equal(o2.Notes, resp[1].Notes)
	assert.Equal(o2.ID, resp[1].ID)
}
