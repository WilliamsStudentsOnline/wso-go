package goodrich

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/sanitize"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// ListOrders godoc
// @Summary List orders
// @Description lists all orders
// @ID goodrich-list-orders
// @Tags goodrich
// @Accept  json
// @Produce  json
// @Param offset query int false "Offset Pagination"
// @Param limit query int false "Limit Pagination"
// @Param type query string false "Sort"
// @Param type query uint false "User ID"
// @Param statuses query []string false "Allowed Status list"
// @Success 200 {array} models.GoodrichOrder
// @Failure 500 {object} lib.APIError
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
// @ID goodrich-get-order
// @Tags goodrich
// @Accept  json
// @Produce  json
// @Param orderID path uint true "Order ID"
// @Success 200 {object} models.GoodrichOrder
// @Failure 500 {object} lib.APIError
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
	PickupTime *time.Time                  `json:"pickupTime"`
	Status     *models.GoodrichOrderStatus `json:"status"`
	AdminNotes *string                     `json:"adminNotes"`
	ItemIDs    *[]uint                     `json:"itemIDs"`
}

// UpdateOrder godoc
// @Summary Update order
// @Description updates an order
// @ID goodrich-update-order
// @Tags goodrich
// @Accept  json
// @Produce  json
// @Param updateParams body goodrich.UpdateOrderParams true "Update Order Params"
// @Param orderID path uint true "Order ID"
// @Success 200 {object} models.GoodrichOrder
// @Failure 500 {object} lib.APIError
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

	// have this above to support changing order status with old pickup times
	if updateData.PickupTime.Before(time.Now()) {
		t.RespondAPIError(c, lib.ErrorGoodrichPickupTimeTooEarly)
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
	order.PickupTime = lib.TimePtrDefaults(updateData.PickupTime, order.PickupTime)
	if updateData.Status != nil {
		order.Status = *updateData.Status
	}
	order.AdminNotes = *lib.StrPtrDefaults(updateData.AdminNotes, &order.AdminNotes)

	if !models.ValidateGoodrichOrderStatus(order.Status) {
		t.RespondAPIError(c, lib.ErrorGoodrichInvalidOrderStatus)
		return
	}

	// TODO[high]: validate for goodrich open hours

	// TODO[high]: validate combo

	// TODO[high]: validate for too expensive on a swipe

	// If we changed items, do shit:
	// Calculate price
	if updateData.ItemIDs != nil {
		var itemIDsStr []string
		var totalPrice float64 = 0
		for _, itemID := range *updateData.ItemIDs {
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
		order.ItemList = itemIDList
		order.TotalPrice = totalPrice
	}

	// Update the menu item in the db
	err = t.orderModel.UpdateOrder(&order)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// TODO[high]: add notifications here

	// Return update menu item
	t.RespondOK(c, order)
}
