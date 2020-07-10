package order

import (
	"time"

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
// @Security Bearer
// @Router /goodrich/order [post]
func (t *Controller) CreateOrder(c *gin.Context) {

	params := CreateOrderParams{}
	err := c.shouldBindQuery(params)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	var order *Order
	err = t.CreateOrder(params, order)
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
// @Router goodrich/users/<user_id>/orders [get]
func (t *Controller) ListUserOrders(c *gin.Context) {
	//// TODO:
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

	orderID, err := services.GetUIntParam(c, "orderID") // TEMPORARY ORDER ID -- MAY CHANGE
	if err != nil {
		t.RespondError(c, err)
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
