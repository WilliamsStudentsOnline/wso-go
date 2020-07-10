package goodrich_order

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
	newOrdder.userID = userID
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

// // TODO: user function
func ListUserOrders(userID uint, orders *[]*Order) (err error) {

}

// // admin function
func (o *OrderModel) ListOrders(orders *[]*Order) (err error) {
	db := o.DB
	//get all rows in Order table and save in orders
	err = db.Find(orders).Error
	return
}

// // TODO: admin function
func UpdateOrder(orderID uint, params UpdateOrderParams) (err error) {

}
