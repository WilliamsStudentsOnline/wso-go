package order

import (
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

type CreateOrderParams struct {
	ItemIDs       []uint `json: "itemIDS"`
	UserID        uint
	PhoneNumber   string    `json: "phoneNumber"`
	PreferredTime time.Time `json: "preferredTime"`
	Notes         string    `json: "notes"`
}

type UpdateOrderParams struct {
	OrderStatus   models.OrderStatus `json: "orderStatus"`
	AdminNotes    string             `json: "adminNotes"`
	EstimatedTime time.Time          `json: "estimatedTime"`
	Items         []models.MenuItem  `json: "items"`
}

// CreateOrder godoc
// @Summary Creates an order
// @Description
// @ID
// @Tags goodrich
// @Accept  json
// @Produce  json
// @Param createParams body order.OrderCreateParams true "Create Order Params"
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
	createParams.UserID = userID

	// phone number shouldn't be nil or blank
	if &createParams.PhoneNumber == nil || createParams.PhoneNumber == "" {
		t.RespondError(c, lib.ErrorGoodrichMissingPhoneNumber)
		return
	}

	// write order to DB
	var newOrder *models.Order
	err = t.orderModel.CreateOrder(userID, createParams, newOrder)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, newOrder)
}

// ListUserOrders godoc
// @Summary Lists orders
// @Description Returns a list of the user's order given a userID
// @ID
// @Tags goodrich
// @Accept  json
// @Produce  json
// @Param
// @Success 201 {object} models.MenuItem
// @Failure 400 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Failure 2155 {object} lib.APIError "the user ID given does not match any in our records"
// @Security Bearer
// @Router goodrich/user/orders [get]
func (t *Controller) ListUserOrders(c *gin.Context) {
	userID := services.GetUserID(c)

	var orders *[]*models.Order

	err := t.orderModel.ListUserOrders(userID, orders)
	if err != nil {
		t.RespondError(c, err)
	}

	t.RespondOK(c, orders)
}

// GetOrder godoc
// @Summary Gets an Order
// @Description Gets an Order using passed orderID
// @ID
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

	//orderID can't be nil or 0
	if err != nil || orderID == 0 {
		t.RespondError(c, lib.ErrorGoodrichMissingOrderID)
		return
	}

	var order *models.Order
	// get the order by ID using the OrderModel's function
	err = t.orderModel.GetOrder(orderID, userID, order)
	if err != nil {
		t.RespondError(c, err) //returns 404 if nothing is found
		return
	}

	t.RespondOK(c, order)
}

// GetOrder godoc
// @Summary Gets an Order
// @Description Gets an Order using passed orderID
// @ID
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
func (t *Controller) GetOrderAdmin(c *gin.Context) {
	orderID, err := services.GetUIntParam(c, "orderID")

	//orderID can't be or 0
	if err != nil || orderID == 0 {
		t.RespondError(c, lib.ErrorGoodrichMissingOrderID)
		return
	}

	var order *models.Order
	// get the order by ID using the OrderModel's function
	err = t.orderModel.GetOrderAdmin(orderID, order)
	if err != nil {
		t.RespondError(c, err) //returns 404 if nothing is found
		return
	}

	t.RespondOK(c, order)
}

// ListOrders godocturn
// @Summary Gets an Order
// @Description
// @ID
// @Tags goodrich
// @Accept  json
// @Produce  json
// @Param
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
	}

	t.RespondOK(c, orders)
}

// UpdateOrder godoc
// @Summary Gets an Order
// @Description
// @ID
// @Tags goodrich
// @Accept  json
// @Produce  json
// @Param
// @Success 201 {object} models.MenuItem
// @Failure 400 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router goodrich/orders/<order_id> [get]
func (t *Controller) UpdateOrder(c *gin.Context) {
	// updateParams := UpdateOrderParams{}
	// err := c.ShouldBindQuery(&updateParams)
	// if err != nil {
	// 	t.RespondError(c, err)
	// 	return
	// }

}
