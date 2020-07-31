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

// Ensure menu items are all stocked and exist
func (o *OrderModel) ValidateMenuItems(menuItemIDs []uint) error {
	for _ , itemID := range menuItemIDs{
		item MenuItem
		err := o.DB.First(itemID, &item).Error
		if err != nil || !item.Available {
			return error
		}
	}
	return nil
}

// Sums up menu items to get total price
func (o *OrderModel) GetMenuItemTotalPrice(menuItemIDs []uint) (float64, error) {
	total_price float64 := 0.0
	for _ , itemID := range menuItemIDs{
		item MenuItem
		err := o.DB.First(itemID, &item).Error
		if err != nil {
			return -1, error
		}
		total_price += item.Price
	}
	return total_price, nil
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
