package models

import (
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

func (o *OrderModel) CreateOrder(userID uint, newOrder *Order) (err error) {
	db := o.DB
	newOrder.UserID = userID
	err = db.Create(newOrder).Error
	return
}

// admin & user function
func (o *OrderModel) GetOrder(orderID uint, order *Order) (err error) {
	db := o.DB
	// if there is an error finding the uint ID, return the error with nil
	err = db.First(order, orderID).Error
	return
}

// admin function
func (o *OrderModel) ListUserOrders(userID uint, orders *[]*Order) (err error) {
	db := o.DB
	err = db.Where("userID = ?", userID).Find(&orders).Error
	return
}

// // admin function
func (o *OrderModel) ListOrders(orders *[]*Order) (err error) {
	db := o.DB
	//get all rows in Order table and save in orders
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
