package models

import (
	"strconv"
	"strings"

	"github.com/WilliamsStudentsOnline/wso-go/services/goodrich/order"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// Order Model
type OrderModel struct {
	*BaseModel
}

func NewOrderModel(db *gorm.DB, log *zap.SugaredLogger) *OrderModel {
	return &OrderModel{
		BaseModel: NewBaseModel(db, log),
	}
}

// Writes fully populated Order to db
func (o *OrderModel) CreateOrder(newOrder *Order) (err error) {
	db := o.DB
	err = db.Create(newOrder).Error
	if err != nil {
		return
	}
	err = db.Find(newOrder).Error
	return
}

// CreateOrderObject : populates an Order object, given CreateOrderParams
func (o *OrderModel) CreateOrderObject(userID uint, createParams order.CreateOrderParams, order *Order) (err error) {
	db := o.DB //maybe should be   db := t.menuModel.DB  ?
	var itemList []string
	var totalPrice float64

	// iterate over ItemIDs in params
	for _, itemID := range createParams.ItemIDs {
		var item MenuItem
		// THIS DEPENDS ON MENU SERVICE FORMATTING || WILL INFER THE MODEL
		err = db.First(&item, itemID).Error
		if err != nil {
			return
		}
		// if the item is available, add it
		if item.Available {
			// add MenuItem to newOrder.Items
			order = append(order.Items, item)
			// append itemID to itemList
			itemList = append(itemList, strconv.FormatUint(uint64(itemID), 10))
			//adding the price for each available item to the total
			totalPrice += item.Price
		}
	}
	order.ItemList = strings.Join(itemList, ",")
	order.Status = OrderStatusPlaced
	order.UserID = userID
	order.PhoneNumber = createParams.PhoneNumber
	order.PreferredTime = createParams.PreferredTime
	order.Notes = createParams.Notes
	order.TotalPrice = totalPrice

	return
}

// admin & user function
func (o *OrderModel) ListUserOrders(userID uint, orders *[]*Order) (err error) {
	db := o.DB
	err = db.Where("userID = ?", userID).Find(&orders).Error
	return
}

// admin & user function
func (o *OrderModel) GetOrder(orderID uint, userID uint, order *Order) (err error) {
	db := o.DB
	// if there is an error finding the uint ID, return the error with nil
	err = db.Where("user_id = ?", userID).First(order, orderID).Error
	return
}

// admin function
func (o *OrderModel) GetOrderAdmin(orderID uint, order *Order) (err error) {
	db := o.DB
	// if there is an error finding the uint ID, return the error with nil
	err = db.First(order, orderID).Error
	return
}

// admin function
func (o *OrderModel) ListOrders(orders *[]*Order) (err error) {
	db := o.DB
	// get all rows in Order table and save in orders
	err = db.Find(orders).Error
	return
}

// admin function
func (o *OrderModel) UpdateOrder(updatedOrder *Order) (err error) {
	db := o.DB
	// save newly updated order in DB
	err = db.Save(updatedOrder).Error
	return
}
c