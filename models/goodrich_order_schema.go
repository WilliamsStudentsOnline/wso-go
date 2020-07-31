package models

import (
	"strconv"
	"strings"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
)

// GoodrichOrderStatus defines an enum of order statuses
type GoodrichOrderStatus int

const (
	// OrderStatusUnknown is default
	GoodrichOrderStatusUnknown GoodrichOrderStatus = iota
	GoodrichOrderStatusPlaced                      // Initial step: set when order is made
	GoodrichOrderStatusApproved
	GoodrichOrderStatusDenied
	GoodrichOrderStatusInProgress
	GoodrichOrderStatusCompleted
	GoodrichOrderStatusPickedUp
)

type Order struct {
	BaseSchema
	User          *User               `json:"user"`
	UserID        uint                `gorm:"index:index_goodrich_orders_on_user_id;not null;"  json:"userID"`
	ItemList      string              `json:"-"`
	Items         []*MenuItem         `gorm:"-" json:"items"`
	Status        GoodrichOrderStatus `json:"status"`
	PhoneNumber   string              `json:"phoneNumber"`
	PreferredTime time.Time           `json:"preferredTime"`
	EstimatedTime *time.Time          `json:"estimatedTime"`
	Notes         string              `json:"notes"`
	TotalPrice    float64             `json:"totalPrice"`
	AdminNotes    string              `json:"adminNotes"`
}

func GoodrichOrderFormatItemList(itemIDs []uint) string {
	lsStr := make([]string, len(itemIDs))
	for i := range itemIDs {
		lsStr[i] = strconv.Itoa(int(itemIDs[i]))
	}
	return strings.Join(lsStr, ",")
}

func GoodrichOrderValidatePreferredTime(preferredTime *time.Time) (err error) {
	now := time.Now()
	if !preferredTime.After(now) {
		return lib.ErrorGoodrichInvalidPreferredTime
	}
	return
}

func (*Order) TableName() string {
	return "goodrich_orders"
}
