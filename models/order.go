package models

import (
	"strconv"
	"time"

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

	var itemList string
	var totalPrice float64

	// iterate over ItemIDs in params
	for _, itemID := range params.ItemIDs {
		var item MenuItem
		// THIS DEPENDS ON MENU SERVICE FORMATTING
		db.Model(&MenuItem{}).Where("ID = ?", itemID).First(&item)
		// if the item is available, add it
		if item.Available {
			// add MenuItem to newOrder.Items
			(*newOrder) = append((*newOrder).Items, item)
			// CHECK FORMATTING -- EXTRA COMMA
			itemList += strconv.FormatUint(itemID, 10) + ","
			//adding the price for each available item to the total
			totalPrice += item.Price
		}
	}

	(*newOrder).ItemList = itemList
	(*newOrder).Status = OrderStatusPlaced
	(*newOrder).User = params.User
	(*newOrder).UserID = params.UserID
	(*newOrder).PhoneNumber = params.PhoneNumber
	(*newOrder).PreferredTime = params.PreferredTime

	// DUMMY NIL TIME for now
	var dummyTime *time.Time
	(*newOrder).EstimatedTime = dummyTime
	(*newOrder).Notes = params.Notes
	(*newOrder).TotalPrice = totalPrice
	// no AdminNotes are set when the user creates an order
	(*newOrder).AdminNotes = ""

	if err = db.Model(&Order{}).Create(newOrder).Error; err != nil {
		// log error in SugaredLogger
		o.log.Fatalln(err)
		(*newOrder) = nil
		return
	}
	return
}

func (o *OrderModel) ListOrders(orders *[]*Order) (err error) {
	db := o.DB

	//get all rows in Order table and save in orders
	if err = db.Model(&Order{}).Find(orders).Error; err != nil {
		//log in Sugaredlogger
		o.log.Fatalln(err)
		(*orders) = nil
		return
	}
	return
}

func (o *OrderModel) GetOrder(uint id, order *Order) (err error) {
	db := o.DB

	// if there is an error finding the uint ID, return the error with nil
	if err = db.Model(&Order{}).First(order, id).Error; err != nil {
		// log error in SugaredLogger
		o.log.Fatalln(err)
		(*order) = nil
		return
	}
	return
}
