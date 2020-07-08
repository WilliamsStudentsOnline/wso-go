package models

import (
	"time"
)

type Order struct {
	BaseSchema // contains IDs and dates and meta info (updated, created, deleted At...)
	// You will need to make a custom format to store these in SQL, b/c a join table would be absolutely massive. I suggest doing a comma separated format like "id,id,id"
	ItemList string `json:"-"` // this means json will ignore this list. Use this for the "item_id,item_id,item_id" format that goes in the DB
	Items         []MenuItem `gorm:"-"` // this means this field will be ignored in the DB. Use this to pull the menu items from ItemList into actual objects in the Model/Controller side.
	Status        OrderStatus
	User          *User
	UserID        uint
	PhoneNumber   string
	PreferredTime time.Time
	EstimatedTime *time.Time
	Notes         string
	TotalPrice    float64
	AdminNotes    string
}


// func (*Order) TableName() string {
// 	return "order"
// }

// Define an enum of order statuses
type OrderStatus int const (
	OrderStatusUnknown    OrderStatus = iota // Default is unknown
	OrderStatusPlaced                 = 100
	OrderStatusApproved               = 200
	OrderStatusDenied                 = 300
	OrderStatusInProgress             = 400
	OrderStatusCompleted              = 500
	OrderStatusPickedUp               = 600
)
