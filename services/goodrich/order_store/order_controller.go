package order_store

import (
	"time"

	"github.com/gin-gonic/gin"
)

type CreateOrderParams struct {
	Items         []MenuItem
	User          *User
	UserID        uint
	PhoneNumber   string
	PreferredTime time.Time
	Notes         string
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
	//// TODO:
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
	//// TODO:
}
