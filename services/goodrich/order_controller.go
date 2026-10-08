package goodrich

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// ListUserOrders godoc
// @Summary List user orders
// @Description lists all user orders
// @ID listUserOrders
// @Tags goodrich
// @Accept  json
// @Produce  json
// @Success 200 {array} models.GoodrichOrder
// @Failure 500 {object} services.BaseErrorResponse
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
// @ID getUserOrder
// @Tags goodrich
// @Accept  json
// @Produce  json
// @Param orderID path uint true "Order ID"
// @Success 200 {object} models.GoodrichOrder
// @Failure 500 {object} services.BaseErrorResponse
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
	Items         []*models.GoodrichOrderItem  `json:"items"`
	LeaseID       string                       `json:"leaseID"`
}

// CreateOrder godoc
// @Summary Create order
// @Description creates an order
// @ID createOrder
// @Tags goodrich
// @Accept  json
// @Produce  json
// @Param createParams body goodrich.CreateOrderParams true "Create Order Params"
// @Success 201 {object} models.GoodrichOrder
// @Failure 500 {object} services.BaseErrorResponse
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

	if createData.LeaseID == "" {
		t.RespondAPIError(c, lib.ErrorGoodrichLeaseMissing)
		return
	}

	if !models.ValidateGoodrichPaymentMethod(createData.PaymentMethod) {
		t.RespondAPIError(c, lib.ErrorGoodrichInvalidPaymentMethod)
		return
	}

	// Make sure order has items
	if len(createData.Items) == 0 {
		t.RespondAPIError(c, lib.ErrorGoodrichOrderNoItems)
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
	validTimeSlots := t.generateTimeSlotsAfter(time.Now())
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
	if t.cfg.GoodrichSlotSpotSize-tsClosedSpots <= 0 {
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

	// Get menu items in an array
	var menuItems []*models.GoodrichMenuItem
	for i := range createData.Items {
		menuItem := &models.GoodrichMenuItem{}
		err = t.menuModel.GetMenuItemByID(createData.Items[i].ID, menuItem)
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

		if menuItem.QuantityLimit && menuItem.Quantity != nil && *menuItem.Quantity <= 0 {
			t.RespondAPIError(c, lib.ErrorGoodrichOutOfMenuItem)
			return
		}

		// Add to slice of menu items
		menuItems = append(menuItems, menuItem)

		// Sanitize create data items
		createData.Items[i].Item = nil
	}

	// Validate combo
	var numBagel, numSpread, numDrink int
	for _, item := range menuItems {
		switch item.Category {
		case "Bagel":
			numBagel++
		case "Spread":
			numSpread++
		case "Drink":
			numDrink++
		}
	}
	// Combo must have at least 1 drink, bagel, spread
	if createData.ComboDeal != nil && *createData.ComboDeal {
		if numBagel < 1 || numSpread < 1 || numDrink < 1 {
			t.RespondAPIError(c, lib.ErrorGoodrichComboDealInvalid)
			return
		}
	}

	// Calculate price
	var totalPrice float64 = 0

	// If in a combo deal, calculate price by finding the maximum price drink,
	// bagel, and spread in the order and set those total to $5.
	if createData.ComboDeal != nil && *createData.ComboDeal {

		// Get rid of max price combo item in each category
		menuItemsLessCombo := removeMaxItemFromSlice(menuItems, "Drink")
		if menuItemsLessCombo == nil {
			t.RespondAPIError(c, lib.ErrorGoodrichComboDealInvalid)
			return
		}
		menuItemsLessCombo = removeMaxItemFromSlice(menuItemsLessCombo, "Bagel")
		if menuItemsLessCombo == nil {
			t.RespondAPIError(c, lib.ErrorGoodrichComboDealInvalid)
			return
		}
		menuItemsLessCombo = removeMaxItemFromSlice(menuItemsLessCombo, "Spread")
		if menuItemsLessCombo == nil {
			t.RespondAPIError(c, lib.ErrorGoodrichComboDealInvalid)
			return
		}

		var priceLessCombo float64 = 0
		for _, item := range menuItemsLessCombo {
			priceLessCombo += item.Price
		}

		// Total price is combo price plus the prices of items not in combo
		totalPrice = 5.00 + priceLessCombo
	} else {
		// If not in a combo deal, calculate price by summing up the prices of
		// each item
		for _, item := range menuItems {
			totalPrice += item.Price
		}
	}

	itemListStr, err := json.Marshal(createData.Items)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// validate overswiping money
	if createData.PaymentMethod == models.GoodrichPaymentMethodSwipe && totalPrice > 5.0 {
		t.RespondAPIError(c, lib.ErrorGoodrichSwipeMaxedOut)
		return
	}

	leaseID, err := uuid.Parse(createData.LeaseID)
	if err != nil {
		t.RespondAPIError(c, lib.ErrorGoodrichLeaseMissing)
		return
	}

	// Ensure that we have a lease to order
	if !t.orderLessor.HasLease(leaseID) {
		t.RespondAPIError(c, lib.ErrorGoodrichLeaseExpired)
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
		ItemList:      string(itemListStr),

		UserID: userID,
	}

	err = t.orderModel.CreateOrder(&order)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Reduce items with quantity limit
	var decItemIDs []uint
	for _, item := range menuItems {
		if item.QuantityLimit {
			decItemIDs = append(decItemIDs, item.ID)
		}
	}
	if len(decItemIDs) > 0 {
		// On error, record but continue because we have already created the order
		err = t.menuModel.DecrementMenuItems(decItemIDs)
		if err != nil {
			t.Log.With(zap.Error(err)).Error("Decrement Menu Items error!")
			err = nil
		}
	}

	// End the lease
	t.orderLessor.EndLease(leaseID)

	// Notifications
	go func() {
		err := t.generateNotifEmail(order, userID)
		if err != nil {
			t.Log.Error("generate email error", err)
		}
	}()

	t.RespondCreated(c, order)
}

func removeMaxItemFromSlice(menuItems []*models.GoodrichMenuItem, category string) []*models.GoodrichMenuItem {
	var maxIdx = -1
	for i, item := range menuItems {
		if item.Category == category {
			if maxIdx == -1 || item.Price > menuItems[maxIdx].Price {
				maxIdx = i
			}
		}
	}

	if maxIdx == -1 {
		return nil
	}

	return append(menuItems[:maxIdx], menuItems[maxIdx+1:]...)
}

// ADMIN BACKDOOR
// AdminCreateOrderParams is a struct to hold the parameters used to create an order.
type AdminCreateOrderParams struct {
	Date          string                       `json:"date"`
	TimeSlot      string                       `json:"timeSlot"` // Format: 11:10 am
	Notes         string                       `json:"notes"`
	ComboDeal     *bool                        `json:"comboDeal"`
	PaymentMethod models.GoodrichPaymentMethod `json:"paymentMethod"`
	Items         []*models.GoodrichOrderItem  `json:"items"`
	//	UserID        uint                         `json:"userID" binding:"required"`
	UnixID string `json:"unixID" binding:"required"`
}

func (t *Controller) AdminCreateOrder(c *gin.Context) {
	// Bind update params
	createData := AdminCreateOrderParams{}
	err := c.ShouldBind(&createData)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	if !models.ValidateGoodrichPaymentMethod(createData.PaymentMethod) {
		t.RespondAPIError(c, lib.ErrorGoodrichInvalidPaymentMethod)
		return
	}

	// Make sure order has items
	if len(createData.Items) == 0 {
		t.RespondAPIError(c, lib.ErrorGoodrichOrderNoItems)
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
	validTimeSlots := t.generateAllDailyTimeSlots()
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
	if t.cfg.GoodrichSlotSpotSize-tsClosedSpots <= 0 {
		t.RespondAPIError(c, lib.ErrorGoodrichTimeFilled)
		return
	}

	// Get menu items in an array
	var menuItems []*models.GoodrichMenuItem
	for i := range createData.Items {
		menuItem := &models.GoodrichMenuItem{}
		err = t.menuModel.GetMenuItemByID(createData.Items[i].ID, menuItem)
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

		if menuItem.QuantityLimit && menuItem.Quantity != nil && *menuItem.Quantity <= 0 {
			t.RespondAPIError(c, lib.ErrorGoodrichOutOfMenuItem)
			return
		}

		// Add to slice of menu items
		menuItems = append(menuItems, menuItem)

		// Sanitize create data items
		createData.Items[i].Item = nil
	}

	// Validate combo
	var numBagel, numSpread, numDrink int
	for _, item := range menuItems {
		switch item.Category {
		case "Bagel":
			numBagel++
		case "Spread":
			numSpread++
		case "Drink":
			numDrink++
		}
	}
	// Combo must have at least 1 drink, bagel, spread
	if createData.ComboDeal != nil && *createData.ComboDeal {
		if numBagel < 1 || numSpread < 1 || numDrink < 1 {
			t.RespondAPIError(c, lib.ErrorGoodrichComboDealInvalid)
			return
		}
	}

	// Calculate price
	var totalPrice float64 = 0

	// If in a combo deal, calculate price by finding the maximum price drink,
	// bagel, and spread in the order and set those total to $5.
	if createData.ComboDeal != nil && *createData.ComboDeal {

		// Get rid of max price combo item in each category
		menuItemsLessCombo := removeMaxItemFromSlice(menuItems, "Drink")
		if menuItemsLessCombo == nil {
			t.RespondAPIError(c, lib.ErrorGoodrichComboDealInvalid)
			return
		}
		menuItemsLessCombo = removeMaxItemFromSlice(menuItemsLessCombo, "Bagel")
		if menuItemsLessCombo == nil {
			t.RespondAPIError(c, lib.ErrorGoodrichComboDealInvalid)
			return
		}
		menuItemsLessCombo = removeMaxItemFromSlice(menuItemsLessCombo, "Spread")
		if menuItemsLessCombo == nil {
			t.RespondAPIError(c, lib.ErrorGoodrichComboDealInvalid)
			return
		}

		var priceLessCombo float64 = 0
		for _, item := range menuItemsLessCombo {
			priceLessCombo += item.Price
		}

		// Total price is combo price plus the prices of items not in combo
		totalPrice = 5.00 + priceLessCombo
	} else {
		// If not in a combo deal, calculate price by summing up the prices of
		// each item
		for _, item := range menuItems {
			totalPrice += item.Price
		}
	}

	itemListStr, err := json.Marshal(createData.Items)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// validate overswiping money
	if createData.PaymentMethod == models.GoodrichPaymentMethodSwipe && totalPrice > 5.0 {
		t.RespondAPIError(c, lib.ErrorGoodrichSwipeMaxedOut)
		return
	}

	user := models.User{}
	err = t.userModel.GetUserByUnixID(createData.UnixID, &user)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	pn := ""
	if user.CellPhone != nil {
		pn = *user.CellPhone
	}

	// Construct new bulletin
	order := models.GoodrichOrder{
		Status:        models.GoodrichOrderStatusPlaced,
		PhoneNumber:   pn,
		TimeSlot:      createData.TimeSlot,
		Date:          createData.Date,
		Notes:         createData.Notes,
		TotalPrice:    totalPrice,
		ComboDeal:     createData.ComboDeal,
		PaymentMethod: createData.PaymentMethod,
		IDNumber:      &user.WilliamsID,
		ItemList:      string(itemListStr),

		UserID: user.ID,
	}

	err = t.orderModel.CreateOrder(&order)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Reduce items with quantity limit
	var decItemIDs []uint
	for _, item := range menuItems {
		if item.QuantityLimit {
			decItemIDs = append(decItemIDs, item.ID)
		}
	}
	if len(decItemIDs) > 0 {
		// On error, record but continue because we have already created the order
		err = t.menuModel.DecrementMenuItems(decItemIDs)
		if err != nil {
			t.Log.With(zap.Error(err)).Error("Decrement Menu Items error!")
			err = nil
		}
	}

	// Notifications
	go func() {
		err := t.generateNotifEmail(order, user.ID)
		if err != nil {
			t.Log.Error("generate email error", err)
		}
	}()

	t.RespondCreated(c, order)
}
