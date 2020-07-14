package order

import (
	"strconv"
	"strings"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
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

	// phone number shouldn't be nil or blank
	if &createParams.PhoneNumber == nil || createParams.PhoneNumber == "" {
		t.RespondAPIError(c, lib.ErrorMissingPhoneNumber)
		return
	}

	// if one nil itemID is found then return nil
	for _, itemID := range createParams.ItemIDs {
		if &itemID == nil {
			t.RespondError(c, lib.ErrorMissingItemID)
			return
		}
	}

	db := t.menuModel.DB
	var newOrder *models.Order
	var itemList []string
	var totalPrice float64

	// iterate over ItemIDs in params
	for _, itemID := range createParams.ItemIDs {
		var item models.MenuItem
		// THIS DEPENDS ON MENU SERVICE FORMATTING || WILL INFER THE MODEL
		err := db.First(&item, itemID).Error
		if err != nil {
			t.RespondAPIError(c, lib.ErrorUnknownItemID)
		}
		// if the item is available, add it
		if item.Available {
			// add MenuItem to newOrder.Items
			newOrder = append(newOrder.Items, item)
			// append itemID to itemList
			itemList = append(itemList, strconv.FormatUint(uint64(itemID), 10))
			//adding the price for each available item to the total
			totalPrice += item.Price
		}
	}
	newOrder.ItemList = strings.Join(itemList, ",")
	newOrder.Status = models.OrderStatusPlaced
	newOrder.PhoneNumber = createParams.PhoneNumber
	newOrder.PreferredTime = createParams.PreferredTime
	newOrder.Notes = createParams.Notes
	newOrder.TotalPrice = totalPrice

	// write order to DB
	err = t.orderModel.CreateOrder(userID, newOrder)
	if err != nil {
		// SPECIFIC ERROR TO RESPOND ?
		t.RespondAPIError(c, lib.ErrorCreateOrderFailed)
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
		t.RespondAPIError(c, lib.ErrorUserIDNotFound)
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
// @Failure 2154 {object} lib.APIError "the orderID given does not match any in our records"
// @Failure 2157 {object} lib.APIError "could not get an order that was not placed by the user"
// @Security Bearer
// @Router goodrich/orders/<order_id> [get]
func (t *Controller) GetOrder(c *gin.Context) {

	userIDin := services.GetUserID(c)
	orderID, err := services.GetUIntParam(c, "orderID")

	//orderID can't be nil or 0
	if err != nil || &orderID == nil || orderID == 0 {
		t.RespondAPIError(c, lib.ErrorMissingOrderID)
		return
	}

	var order *models.Order
	// get the order by ID using the OrderModel's function
	err = t.orderModel.GetOrder(orderID, order)
	if err != nil {
		t.RespondAPIError(c, lib.ErrorOrderIDNotFound)
		return
	}
	//Does this break if admin is trying to get an order?
	//how to check if admin, then ignore this checker
	if order.UserID != userIDin {
		t.RespondAPIError(c, lib.ErrorUserCannotAccessOrder)
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

	if gorm.IsRecordNotFoundError(err) {
		t.RespondAPIError(c, lib.ErrorUserIDNotFound)
	} else {
		t.RespondAPIError(c, lib.ErrorListOrdersFailed)
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
