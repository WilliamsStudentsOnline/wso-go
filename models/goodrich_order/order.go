package goodrich_order

import (
	"strings"

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

func (o *OrderModel) CreateOrder(params CreateOrderParams, newOrder *Order) (err error) {
	db := o.DB

	var itemList []string
	var totalPrice float64

	// iterate over ItemIDs in params
	for _, itemID := range params.ItemIDs {
		var item MenuItem
		// THIS DEPENDS ON MENU SERVICE FORMATTING
		db.Model(&MenuItem{}).First(&item, id)
		// if the item is available, add it
		if item.Available {
			// add MenuItem to newOrder.Items
			newOrder = append(newOrder.Items, item)
			// CHECK FORMATTING -- EXTRA COMMA
			itemList = append(itemList, itemID)
			//adding the price for each available item to the total
			totalPrice += item.Price
		}
	}

	// Estimated time and AdminNotes are nil by default
	//newOrder.EstimatedTime = nil
	//newOrder.AdminNotes = nil
	newOrder.ItemList = strings.Join(itemList, ",")
	newOrder.Status = OrderStatusPlaced
	newOrder.User = params.User
	newOrder.UserID = params.UserID
	newOrder.PhoneNumber = params.PhoneNumber
	newOrder.PreferredTime = params.PreferredTime
	newOrder.Notes = params.Notes
	newOrder.TotalPrice = totalPrice

	err = db.Create(newOrder).Error
	return
}

func (o *OrderModel) ListOrders(orders *[]*Order) (err error) {
	db := o.DB
	//get all rows in Order table and save in orders
	err = db.Find(orders).Error
	return
}

func (o *OrderModel) GetOrder(orderID uint, order *Order) (err error) {
	db := o.DB
	// if there is an error finding the uint ID, return the error with nil
	err = db.First(order, orderID).Error
	return
}

//// TODO:

//func ListUserOrders(userID uint, orders *[]*Order) (err error)

//func UpdateOrder(orderID uint, params UpdateOrderParams) (err error, order Order)
