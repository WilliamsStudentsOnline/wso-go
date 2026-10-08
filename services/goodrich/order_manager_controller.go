package goodrich

import (
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/sanitize"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

// ListOrders godoc
// @Summary List orders
// @Description lists all orders
// @ID listOrders
// @Tags goodrich
// @Accept  json
// @Produce  json
// @Param offset query int false "Offset Pagination"
// @Param limit query int false "Limit Pagination"
// @Param sort query string false "Sort"
// @Param date query string false "Date"
// @Param userID query uint false "User ID"
// @Param statuses query []string false "Allowed Status list"
// @Success 200 {array} models.GoodrichOrder
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /goodrich/orders [get]
func (t *Controller) ListOrders(c *gin.Context) {
	params := new(models.GetAllGoodrichOrdersOptions)

	err := c.ShouldBindQuery(params)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	var order []*models.GoodrichOrder
	err = t.orderModel.GetAllGoodrichOrders(&order, params)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, order)
}

// GetOrder godoc
// @Summary Get order
// @Description gets an order
// @ID getOrder
// @Tags goodrich
// @Accept  json
// @Produce  json
// @Param orderID path uint true "Order ID"
// @Success 200 {object} models.GoodrichOrder
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /goodrich/orders/{orderID} [get]
func (t *Controller) GetOrder(c *gin.Context) {
	// Decode orderID.
	orderID, err := services.GetUIntParam(c, "orderID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	order := models.GoodrichOrder{}
	err = t.orderModel.GetGoodrichOrder(&order, orderID, true)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// TODO[medium]: sanitize preloaded user better
	sanitize.User(order.User, c)

	t.RespondOK(c, order)
}

// UpdateOrderParams is a struct to hold the parameters used to update an order.
type UpdateOrderParams struct {
	Status     *models.GoodrichOrderStatus `json:"status"`
	AdminNotes *string                     `json:"adminNotes"`
}

// UpdateOrder godoc
// @Summary Update order
// @Description updates an order
// @ID updateOrder
// @Tags goodrich
// @Accept  json
// @Produce  json
// @Param updateParams body goodrich.UpdateOrderParams true "Update Order Params"
// @Param orderID path uint true "Order ID"
// @Success 200 {object} models.GoodrichOrder
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /goodrich/orders/{orderID} [patch]
func (t *Controller) UpdateOrder(c *gin.Context) {
	// Decode parameter
	orderID, err := services.GetUIntParam(c, "orderID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Bind update params
	updateData := UpdateOrderParams{}
	err = c.ShouldBind(&updateData)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	// Do database query to get bulletin
	var order models.GoodrichOrder
	err = t.orderModel.GetGoodrichOrder(&order, orderID, true)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Update fields: this is a bit long and verbose, but I don't want to mess with reflect
	if updateData.Status != nil {
		order.Status = *updateData.Status
	}
	order.AdminNotes = *lib.StrPtrDefaults(updateData.AdminNotes, &order.AdminNotes)

	if !models.ValidateGoodrichOrderStatus(order.Status) {
		t.RespondAPIError(c, lib.ErrorGoodrichInvalidOrderStatus)
		return
	}

	// Update the menu item in the db
	err = t.orderModel.UpdateOrder(&order)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// TODO[medium]: add notifications here

	// Return update menu item
	t.RespondOK(c, order)
}
