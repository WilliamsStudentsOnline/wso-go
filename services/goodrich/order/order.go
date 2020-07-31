package order

import (
	"net/http"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

type CreateOrderParams struct {
	ItemIDs       []uint    `json:"itemIDs"`
	PhoneNumber   string    `json:"phoneNumber"`
	PreferredTime time.Time `json:"preferredTime"`
	Notes         string    `json:"notes"`
	TotalPrice    float64   `json:"totalPrice"`
}

// CreateOrder godoc
// @Summary Creates an order
// @Description
// @ID goodrich-order-create-order
// @Tags goodrich
// @Accept  json
// @Produce  json
// @Param createParams body order.CreateOrderParams true "Create Order Params"
// @Success 201 {object} models.MenuItem
// @Failure 400 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Failure 2150 {object}	lib.APIError "unknown item(s) id"
// @Failure 2151 {object} lib.APIError "missing phone number in create order params"
// @Failure 2153 {object} lib.APIError "order could not be created in table"
// @Security Bearer
// @Router /goodrich/order [post]
func (t *Controller) CreateOrder(c *gin.Context) {

	userID := services.GetUserID(c)
	createParams := CreateOrderParams{}
	err := c.ShouldBindQuery(&createParams)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// phone number shouldn't be nil or blank
	if createParams.PhoneNumber == "" {
		t.RespondError(c, lib.ErrorGoodrichMissingPhoneNumber)
		return
	}

	// Validates menu items
	if err := t.orderModel.ValidateMenuItems(createParams.ItemIDs); err != nil {
		t.RespondError(c, err)
		return
	}
	//reminder: supplement our own .lib error when item is not valids

	// Check estimated total price with client's passed total price to ensure they are the same
	estTotalPrice, err := t.orderModel.GetMenuItemTotalPrice(createParams.ItemIDs)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if estTotalPrice != createParams.TotalPrice {
		t.RespondAPIError(c, lib.ErrorGoodrichOrderPriceMismatch)
		return
	}

	// Create order struct
	order := &models.Order{
		UserID:        userID,
		Status:        models.GoodrichOrderStatusPlaced,
		PhoneNumber:   createParams.PhoneNumber,
		PreferredTime: createParams.PreferredTime, // TODO: validate time to ensure in future
		Notes:         createParams.Notes,
		ItemList:      models.GoodrichOrderFormatItemList(createParams.ItemIDs),
		TotalPrice:    estTotalPrice,
	}

	err = t.orderModel.CreateOrder(order)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	t.RespondOK(c, order)
}

// ListUserOrders godoc
// @Summary Lists orders
// @Description Returns a list of the user's order given a userID
// @ID goodrich-order-list-users-orders
// @Tags goodrich
// @Accept  json
// @Produce  json
// @Success 201 {object} models.MenuItem
// @Failure 400 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Failure 2155 {object} lib.APIError "the user ID given does not match any in our records"
// @Security Bearer
// @Router goodrich/user/orders [get]
func (t *Controller) ListUserOrders(c *gin.Context) {
	userID := services.GetUserID(c)

	// TODO: we probably only want active orders. Do this in another PR though.

	var orders []*models.Order
	err := t.orderModel.ListUserOrders(userID, &orders)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, orders)
}

// GetOrder godoc
// @Summary Gets an Order
// @Description Gets an Order using passed orderID
// @ID goodrich-order-get-order
// @Tags goodrich
// @Accept  json
// @Produce  json
// @Param	orderID path uint true "Order ID"
// @Param	userID body uint true "User ID"
// @Success 201 {object} models.OrderModel
// @Failure 400 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Failure 2152 {object} lib.APIError "missing order id in input"
// @Security Bearer
// @Router goodrich/orders/<order_id> [get]
func (t *Controller) GetOrder(c *gin.Context) {

	userID := services.GetUserID(c)
	orderID, err := services.GetUIntParam(c, "orderID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	var order *models.Order
	// get the order by ID using the OrderModel's function
	if auth.HasScope(c, auth.ScopeGoodrichAdmin) {
		err = t.orderModel.GetOrderAdmin(userID, order)
	} else {
		err = t.orderModel.GetOrder(orderID, userID, order)
	}
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, order)
}

// ListOrders godocturn
// @Summary Gets an Order
// @Description
// @ID goodrich-order-list-orders
// @Tags goodrich
// @Accept  json
// @Produce  json
// @Success 201 {object} models.MenuItem
// @Failure 400 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Failure 2155 {object} lib.APIError "the user ID given does not match any in our records"
// @Failure 2156 {object} lib.APIError "could not list orders in the table"
// @Security Bearer
// @Router goodrich/orders/<order_id> [get]
func (t *Controller) ListOrders(c *gin.Context) {
	var orders *[]*models.Order

	err := t.orderModel.ListOrders(orders)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, orders)
}

type UpdateOrderParams struct {
	OrderStatus   *models.GoodrichOrderStatus `json:"orderStatus"`
	AdminNotes    *string                     `json:"adminNotes"`
	EstimatedTime *time.Time                  `json:"estimatedTime"`
	ItemIDs       *[]uint                     `json:"items"`
}

// UpdateOrder godoc
// @Summary Gets an Order
// @Description
// @ID goodrich-order-update-order
// @Tags goodrich
// @Accept  json
// @Produce  json
// @Param updateParams body order.UpdateOrderParams true "Update Order Params"
// @Success 201 {object} models.MenuItem
// @Failure 400 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router goodrich/orders/<order_id> [patch]
func (t *Controller) UpdateOrder(c *gin.Context) {
	//get order id to find old order in db
	orderID, err := services.GetUIntParam(c, "orderID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	//bind new order params to update old order
	updateParams := UpdateOrderParams{}
	err = c.ShouldBindQuery(&updateParams)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	//get the order currently in the db using Admin function
	var currOrder models.Order
	err = t.orderModel.GetOrderAdmin(orderID, &currOrder)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	currOrder.AdminNotes = *lib.StrPtrDefaults(updateParams.AdminNotes, &currOrder.AdminNotes)
	currOrder.EstimatedTime = lib.TimePtrDefaults(updateParams.EstimatedTime, currOrder.EstimatedTime)
	if updateParams.OrderStatus != nil {
		currOrder.Status = *updateParams.OrderStatus
	}
	// If updating itemIDs, validate items, set items, and set new price
	if updateParams.ItemIDs != nil && len(*updateParams.ItemIDs) > 0 {
		// validate item ids here
		if err := t.orderModel.ValidateMenuItems(*updateParams.ItemIDs); err != nil {
			t.RespondError(c, err)
			return
		}

		// Check estimated total price with client's passed total price to ensure they are the same
		estTotalPrice, err := t.orderModel.GetMenuItemTotalPrice(*updateParams.ItemIDs)
		if err != nil {
			t.RespondError(c, err)
			return
		}

		currOrder.ItemList = models.GoodrichOrderFormatItemList(*updateParams.ItemIDs)
		currOrder.TotalPrice = estTotalPrice
	}

	// update order in the database
	err = t.orderModel.UpdateOrder(&currOrder)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, currOrder)
}
