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
	Date          string                       `json:"date"`
	TimeSlot      string                       `json:"timeSlot"` // Format: 11:10 am
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

	// validate for date exists
	if createData.Date != time.Now().Format(DateFormat) {
		t.RespondAPIError(c, lib.ErrorGoodrichTimeBadDay)
		return
	}

	// validate for date open
	foundValidDate := false
	for _, date := range t.cfg.GoodrichOpenDays {
		if createData.Date == date {
			foundValidDate = true
			break
		}
	}
	if !foundValidDate {
		t.RespondAPIError(c, lib.ErrorGoodrichDateClosed)
		return
	}

	// validate for timeslot exists
	foundValidTimeSlot := false
	validTimeSlots := generateTimeSlotsAfter(time.Now())
	for _, slot := range validTimeSlots {
		if createData.TimeSlot == slot.String() {
			foundValidTimeSlot = true
			break
		}
	}

	if !foundValidTimeSlot {
		t.RespondAPIError(c, lib.ErrorGoodrichTimeSlotInvalid)
		return
	}

	// validate for timeslot open
	tsAvailabilityMap, err := t.generateSlotAvailabilityMap(createData.Date)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	tsClosedSpots := tsAvailabilityMap[createData.TimeSlot]
	if goodrichSlotSpotSize-tsClosedSpots <= 0 {
		t.RespondAPIError(c, lib.ErrorGoodrichTimeFilled)
		return
	}

	if createData.PaymentMethod == models.GoodrichPaymentMethodSwipe || createData.PaymentMethod == models.GoodrichPaymentMethodPoints {
		if createData.IDNumber == nil {
			t.RespondAPIError(c, lib.ErrorGoodrichMissingWilliamsID)
			return
		}
	}

	// TODO[medium]: validate phone #

	// Calculate price
	var itemIDsStr []string
	var totalPrice float64 = 0
	// Validate combo
	var numBagel, numSpread, numDrink, numOther int
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

		switch menuItem.Category {
		case "Bagel":
			numBagel++
		case "Spread":
			numSpread++
		case "Drink":
			numDrink++
		default:
			numOther++
		}

		totalPrice += menuItem.Price
		itemIDsStr = append(itemIDsStr, strconv.Itoa(int(itemID)))
	}

	if createData.ComboDeal != nil && *createData.ComboDeal {
		if numBagel != 1 || numSpread != 1 || numDrink != 1 || numOther != 0 {
			t.RespondAPIError(c, lib.ErrorGoodrichComboDealInvalid)
			return
		}

		// if it is a valid combo deal, set total price to 5
		totalPrice = 5.0
	}

	itemIDList := strings.Join(itemIDsStr, ",")

	// validate overswiping money
	if createData.PaymentMethod == models.GoodrichPaymentMethodSwipe && totalPrice > 5.0 {
		t.RespondAPIError(c, lib.ErrorGoodrichSwipeMaxedOut)
		return
	}

	// Construct new bulletin
	order := models.GoodrichOrder{
		Status:        models.GoodrichOrderStatusPlaced,
		PhoneNumber:   createData.PhoneNumber,
		TimeSlot:      createData.TimeSlot,
		Date:          createData.Date,
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

	// Notifications
	go func() {
		err := t.generateNotifEmail(order, userID)
		if err != nil {
			t.Log.Error("generate email error", err)
		}
	}()

	t.RespondCreated(c, order)
}
