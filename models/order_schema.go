package models

import (
	"time"
)

type Order struct {
	BaseSchema // contains IDs and dates and meta info (updated, created, deleted At...)
	// You will need to make a custom format to store these in SQL, b/c a join table would be absolutely massive. I suggest doing a comma separated format like "id,id,id"
	ItemList      string      `json:"-"`               // this means json will ignore this list. Use this for the "item_id,item_id,item_id" format that goes in the DB
	Items         []MenuItem  `gorm:"-" json: "items"` // this means this field will be ignored in the DB. Use this to pull the menu items from ItemList into actual objects in the Model/Controller side.
	Status        OrderStatus `json:"status"`
	User          *User       `json:"user"`
	UserID        uint        `json:"userID"`
	PhoneNumber   string      `json:"phoneNumber"`
	PreferredTime time.Time   `json:"preferredTime"`
	EstimatedTime *time.Time  `json:"estimatedTime"`
	Notes         string      `json:"notes"`
	TotalPrice    float64     `json:"totalPrice"`
	AdminNotes    string      `json:"adminNotes"`
}

func (*Order) TableName() string {
	return "orders"
}

// Define an enum of order statuses
type OrderStatus int

const (
	OrderStatusUnknown OrderStatus = iota // Default is unknown
	OrderStatusPlaced                     // Initial step: set when order is made
	OrderStatusApproved
	OrderStatusDenied
	OrderStatusInProgress
	OrderStatusCompleted
	OrderStatusPickedUp
)
