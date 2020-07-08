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

type CreateOrderParams struct {
	Items         []MenuItem
	User          *User
	UserID        uint
	PhoneNumber   string
	PreferredTime time.Time
	Notes         string
}
func (*OrderModel) CreateOrder(params CreateOrderParams) Order {
	var newOrder Order

  var totalPrice float64
  for _, item := range params.Items {
    if item.Available {
      totalPrice += item.Price;
    }
  }


  newOrder = {
    "",
    params.Items,
    OrderStatusPlaced,
    params.User,
    params.PhoneNumber,
    params.PreferredTime,


  }
  order.ItemList = ""; // list of strings?

  order.Items = params.Items;
  order.Status = OrderStatusPlaced;
  order.User = params.User;
  order.UserID = params.
  order.PhoneNumber =
  order.PreferredTime =
  order.EstimatedTime =
  order.Notes =
  order.TotalPrice =
  order.AdminNotes =
  return newOrder
}

func (*OrderModel) ListOrders() []Order {

}

func (*OrderModel) GetOrder(uint id) Order {

}
