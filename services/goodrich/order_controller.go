package goodrich

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// ListUserOrders godoc
// @Summary List user orders
// @Description lists all user orders
// @ID goodrich-list-user-orders
// @Tags goodrich
// @Accept  json
// @Produce  json
// @Success 200 {array} models.GoodrichOrder
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /goodrich/user/orders [get]
func (t *Controller) ListUserOrders(c *gin.Context) {
	userID := services.GetUserID(c)

	var orders []*models.GoodrichOrder
	err := t.orderModel.GetGoodrichUserOrders(&orders, userID)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, orders)
}

// GetUserOrder godoc
// @Summary Get user order
// @Description gets a user order
// @ID goodrich-get-user-order
// @Tags goodrich
// @Accept  json
// @Produce  json
// @Param orderID path uint true "Order ID"
// @Success 200 {object} models.GoodrichOrder
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /goodrich/user/orders/{orderID} [get]
func (t *Controller) GetUserOrder(c *gin.Context) {
	userID := services.GetUserID(c)

	// Decode orderID.
	orderID, err := services.GetUIntParam(c, "orderID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	order := models.GoodrichOrder{}
	err = t.orderModel.GetGoodrichOrder(&order, orderID, false)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Must be owning your own order
	if order.UserID != userID {
		t.RespondError(c, lib.ErrorMustBeSelf)
		return
	}

	t.RespondOK(c, order)
}

// CreateOrderParams is a struct to hold the parameters used to create an order.
type CreateOrderParams struct {
	PhoneNumber   string                       `json:"phoneNumber"`
	PreferredTime time.Time                    `json:"preferredTime"`
	Notes         string                       `json:"notes"`
	ComboDeal     *bool                        `json:"comboDeal"`
	PaymentMethod models.GoodrichPaymentMethod `json:"paymentMethod"`
	IDNumber      *string                      `json:"idNumber"`
	ItemIDs       []uint                       `json:"itemIDs"`
}

// CreateOrder godoc
// @Summary Create order
// @Description creates an order
// @ID goodrich-create-order
// @Tags goodrich
// @Accept  json
// @Produce  json
// @Param createParams body goodrich.CreateOrderParams true "Create Order Params"
// @Success 201 {object} models.GoodrichOrder
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /goodrich/orders [post]
func (t *Controller) CreateOrder(c *gin.Context) {
	userID := services.GetUserID(c)

	// Bind update params
	createData := CreateOrderParams{}
	err := c.ShouldBind(&createData)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	if !models.ValidateGoodrichPaymentMethod(createData.PaymentMethod) {
		t.RespondAPIError(c, lib.ErrorGoodrichInvalidPaymentMethod)
		return
	}

	if createData.PreferredTime.Before(time.Now()) {
		t.RespondAPIError(c, lib.ErrorGoodrichPreferredTimeTooEarly)
		return
	}

	if createData.PaymentMethod == models.GoodrichPaymentMethodSwipe || createData.PaymentMethod == models.GoodrichPaymentMethodPoints {
		if createData.IDNumber == nil {
			t.RespondAPIError(c, lib.ErrorGoodrichMissingWilliamsID)
			return
		}
	}

	// TODO[high]: validate for goodrich open hours

	// TODO[high]: validate combo

	// TODO[high]: validate for too expensive on a swipe

	// TODO[medium]: validate phone #

	// Calculate price
	var itemIDsStr []string
	var totalPrice float64 = 0
	for _, itemID := range createData.ItemIDs {
		menuItem := &models.GoodrichMenuItem{}
		err = t.menuModel.GetMenuItemByID(itemID, menuItem)
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				t.RespondAPIError(c, lib.ErrorGoodrichUnknownMenuItem)
				return
			}
			t.RespondError(c, err)
			return
		}

		if !menuItem.Available {
			t.RespondAPIError(c, lib.ErrorGoodrichUnavailableMenuItem)
			return
		}

		totalPrice += menuItem.Price
		itemIDsStr = append(itemIDsStr, strconv.Itoa(int(itemID)))
	}

	itemIDList := strings.Join(itemIDsStr, ",")

	// Construct new bulletin
	order := models.GoodrichOrder{
		Status:        models.GoodrichOrderStatusPlaced,
		PhoneNumber:   createData.PhoneNumber,
		PreferredTime: createData.PreferredTime,
		Notes:         createData.Notes,
		TotalPrice:    totalPrice,
		ComboDeal:     createData.ComboDeal,
		PaymentMethod: createData.PaymentMethod,
		IDNumber:      createData.IDNumber,
		ItemList:      itemIDList,

		UserID: userID,
	}

	err = t.orderModel.CreateOrder(&order)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// TODO[medium]: add notifications here

	t.RespondCreated(c, order)
}
