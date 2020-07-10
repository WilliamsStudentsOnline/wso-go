package order

import (
	"strings"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
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
// @Failure 2151 {object} lib.APIError
// @Failure 2152 {object}	lib.APIError
// @Security Bearer
// @Router /goodrich/order [post]
func (t *Controller) CreateOrder(c *gin.Context) {

	userID := services.GetUserID(c)
	createParams := CreateOrderParams{}
	err := c.shouldBindQuery(&createParams)

	if err != nil {
		t.RespondError(c, err)
		return
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
			// CHECK FORMATTING -- EXTRA COMMA
			itemList = append(itemList, itemID)
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

	// check errors
	if createParams.PhoneNumber == nil {
		t.RespondError(c, lib.ErrorMissingPhoneNumber)
		return
	}
	// PhoneNumber cannot be blank
	if (createData.PhoneNumber != nil) && (*createData.PhoneNumber == "") {
		t.RespondError(c, lib.ErrorMissingPhoneNumber)
		return
	}
	// if one nil itemID is found then return nil
	for _, itemID := range createParams.ItemIDs {
		if itemID == nil {
			t.RespondError(c, lib.ErrorUnknownItemID)
			return
		}
	}

	err = t.CreateOrder(userID, newOrder)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, order)
}

// ListUserOrders godoc
// @Summary Lists orders
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
// @Router goodrich/user/orders [get]
func (t *Controller) ListUserOrders(c *gin.Context) {
	//// TODO:
	// userID := services.GetUserID(c)
	// somewhere here use c.shouldBind()
}

// GetOrder godoc
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
func (t *Controller) GetOrder(c *gin.Context) {
	userID := services.GetUserID(c)

	orderID, err := services.GetUIntParam(c, "orderID")

	if err != nil {
		t.RespondError(c, err)
		return
	}

	if orderID == nil {
		t.RespondError(c, lib.ErrorUnknownItemID)
		return
	}

	var order *Order
	// get the order by ID using the OrderModel's function
	err = t.OrderModel.GetOrder(orderID, order)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, order)
}

/*
# github.com/WilliamsStudentsOnline/wso-go/services/goodrich/order
./controller.go:13:27: undefined: models.OrderModel
./controller.go:20:19: undefined: models.NewOrderModel
./order_controller.go:21:16: undefined: OrderStatus
./order_controller.go:24:18: undefined: MenuItems
./order_controller.go:44:10: c.shouldBindQuery undefined (type *gin.Context has no field or method shouldBindQuery, but does have ShouldBindQuery)
./order_controller.go:52:16: undefined: Order
./order_controller.go:58:13: undefined: MenuItem
./order_controller.go:60:20: undefined: id
./order_controller.go:66:22: cannot use itemID (type uint) as type string in append
./order_controller.go:72:23: undefined: strings
./order_controller.go:72:23: too many errors
*/
