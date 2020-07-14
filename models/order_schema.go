package models

import (
	"time"
)

// GoodrichOrderStatus defines an enum of order statuses
type GoodrichOrderStatus int

const (
	// OrderStatusUnknown is default
	OrderStatusUnknown GoodrichOrderStatus = iota
	OrderStatusPlaced                      // Initial step: set when order is made
	OrderStatusApproved
	OrderStatusDenied
	OrderStatusInProgress
	OrderStatusCompleted
	OrderStatusPickedUp
)

type Order struct {
	BaseSchema
	ItemList      string              `json:"-"`
	Items         []MenuItem          `gorm:"-" json: "items"`
	Status        GoodrichOrderStatus `json:"status"`
	User          *User               `json:"user"`
	UserID        uint                `gorm:"index:index_goodrich_orders_on_user_id;not null;"  json:"userID"`
	PhoneNumber   string              `json:"phoneNumber"`
	PreferredTime time.Time           `json:"preferredTime"`
	EstimatedTime *time.Time          `json:"estimatedTime"`
	Notes         string              `json:"notes"`
	TotalPrice    float64             `json:"totalPrice"`
	AdminNotes    string              `json:"adminNotes"`
}

func (*Order) TableName() string {
	return "goodrich_orders"
}
