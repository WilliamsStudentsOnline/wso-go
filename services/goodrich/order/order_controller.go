package order

import (
	"strconv"
	"strings"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

type CreateOrderParams struct {
	ItemIDs []uint `json: "itemIDS"`
	//User          *User
	//UserID        uint
	PhoneNumber   string    `json: "phoneNumber"`
	PreferredTime time.Time `json: "preferredTime"`
	Notes         string    `json: "notes"`
}

type UpdateOrderParams struct {
	OrderStatus   OrderStatus `json: "orderStatus"`
	AdminNotes    string      `json: "adminNotes"`
	EstimatedTime time.Time   `json: "estimatedTime"`
	Items         []MenuItems `json: "items"`
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

	// check errors
	if &createParams.PhoneNumber == nil {
		t.RespondError(c, lib.ErrorMissingPhoneNumber)
		return
	}
	// PhoneNumber cannot be blank
	if (&createParams.PhoneNumber != nil) && (createParams.PhoneNumber == "") {
		t.RespondError(c, lib.ErrorMissingPhoneNumber)
		return
	}
	// if one nil itemID is found then return nil
	for _, itemID := range createParams.ItemIDs {
		if &itemID == nil {
			t.RespondError(c, lib.ErrorUnknownItemID)
			return
		}
	}

	db := t.orderModel.DB
	var newOrder *Order
	var itemList []string
	var totalPrice float64

	// iterate over ItemIDs in params
	for _, itemID := range createParams.ItemIDs {
		var item MenuItem
		// THIS DEPENDS ON MENU SERVICE FORMATTING || WILL INFER THE MODEL
		db.First(&item, id)
		// if the item is available, add it
		if item.Available {
			// add MenuItem to newOrder.Items
			newOrder = append(newOrder.Items, item)
			// append itemID to itemList
			itemList = append(itemList, strconv.FormatUint(itemID))
			//adding the price for each available item to the total
			totalPrice += item.Price
		}
	}
	newOrder.ItemList = strings.Join(itemList, ",")
	newOrder.Status = OrderStatusPlaced
	newOrder.PhoneNumber = createParams.PhoneNumber
	newOrder.PreferredTime = createParams.PreferredTime
	newOrder.Notes = createParams.Notes
	newOrder.TotalPrice = totalPrice

	// write order to DB
	err = t.OrderModel.CreateOrder(userID, newOrder)
	if err != nil {
		// SPECIFIC ERROR TO RESPOND ?
		t.RespondError(c, lib.ErrorCreateOrderFailed)
		return
	}

	t.RespondOK(c, order)
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
// @Security Bearer
// @Router goodrich/user/orders [get]
func (t *Controller) ListUserOrders(c *gin.Context) {
	//// TODO:
	// userID := services.GetUserID(c)
	// somewhere here use c.ShouldBind(&)
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
// @Failure 2154 {object} lib.APIError "the orderID given does not match any in our records"
// @Security Bearer
// @Router goodrich/orders/<order_id> [get]
func (t *Controller) GetOrder(c *gin.Context) {

	userID := services.GetUserID(c)
	orderID, err := services.GetUIntParam(c, "orderID")

	if err != nil || orderID == nil {
		t.RespondError(c, lib.ErrorMissingOrderID)
		return
	}

	var order *Order
	// get the order by ID using the OrderModel's function
	err = t.OrderModel.GetOrder(orderID, order)
	if err != nil {
		t.RespondError(c, lib.ErrorOrderIDNotFound)
		return
	}

	t.RespondOK(c, order)
}

// ListOrders godoc
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
func (t *Controller) ListOrders(c *gin.Context) {

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

}
