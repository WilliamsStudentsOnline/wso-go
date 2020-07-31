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

// Writes fully populated Order to db
func (o *OrderModel) CreateOrder(newOrder *Order) (err error) {
	err = o.DB.Create(newOrder).Error
	if err != nil {
		return
	}
	err = o.DB.Find(newOrder).Error
	return
}

// Ensures menu items are all stocked and exist
func (o *OrderModel) ValidateMenuItems(menuItemIDs []uint) error {
	// TODO: fill this in once menu store is merged
	// in for loop, get current menuitem
	// CHECK IF ITEM EXISTS IN DB ?
	// get menu object from db and make sure it is avaialbe
	// break loop if one item isn't

	o.log.Warn("Calling incomplete function! DANGEROUS!")
	return nil
}

// Sums up menu items to get total price
func (o *OrderModel) GetMenuItemTotalPrice(menuItemIDs []uint) (float64, error) {
	// TODO: fill this in once menu store is merged
	// define totalPrice
	// get each menuItemID from db and sum total prices. return total price, nil
	o.log.Warn("Calling incomplete function! DANGEROUS!")
	return 0, nil
}

// admin & user function
func (o *OrderModel) ListUserOrders(userID uint, orders *[]*Order) (err error) {
	err = o.DB.Where("userID = ?", userID).Find(&orders).Error
	return
}

// admin & user function
func (o *OrderModel) GetOrder(orderID uint, userID uint, order *Order) (err error) {
	// if there is an error finding the uint ID, return the error with nil
	err = o.DB.Where("user_id = ?", userID).First(order, orderID).Error
	return
}

// admin function
func (o *OrderModel) GetOrderAdmin(orderID uint, order *Order) (err error) {
	// if there is an error finding the uint ID, return the error with nil
	err = o.DB.First(order, orderID).Error
	return
}

// admin function
func (o *OrderModel) ListOrders(orders *[]*Order) (err error) {
	// get all rows in Order table and save in orders
	err = o.DB.Find(orders).Error
	return
}

// admin function
func (o *OrderModel) UpdateOrder(updatedOrder *Order) (err error) {
	// save newly updated order in DB
	err = o.DB.Save(updatedOrder).Error
	return
}
