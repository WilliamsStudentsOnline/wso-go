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

func (o *OrderModel) CreateOrder(userID uint, createParams order.CreateOrderParams, newOrder *Order) (err error) {
	db := o.DB //maybe should be   db := t.menuModel.DB  ?
	// Todo: l.26-52 should be in a separate function
	var itemList []string
	var totalPrice float64

	// iterate over ItemIDs in params
	for _, itemID := range createParams.ItemIDs {
		var item models.MenuItem
		// THIS DEPENDS ON MENU SERVICE FORMATTING || WILL INFER THE MODEL
		err = db.First(&item, itemID).Error
		if err != nil {
			return
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
	newOrder.Status = OrderStatusPlaced
	newOrder.PhoneNumber = createParams.PhoneNumber
	newOrder.PreferredTime = createParams.PreferredTime
	newOrder.Notes = createParams.Notes
	newOrder.TotalPrice = totalPrice

	err = db.Create(newOrder).Error
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
func (o *OrderModel) UpdateOrder(orderID uint, updatedOrder *Order) (err error) {
	db := o.DB
	// get order currently in the table by orderID and save in tempOrder
	var tempOrder *Order
	err = db.First(tempOrder, orderID).Error
	if err != nil {
		return
	}
	// update tempOrder to be equal to the updatedOrder
	*tempOrder = *updatedOrder
	// save newly updated order in DB
	err = db.Save(tempOrder).Error
	return
}
