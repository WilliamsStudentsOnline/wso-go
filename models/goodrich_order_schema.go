package models

import (
	"strings"

	"github.com/jinzhu/gorm"
)

type GoodrichOrderStatus int

const (
	GoodrichOrderStatusUnknown GoodrichOrderStatus = iota // Default is unknown
	GoodrichOrderStatusPlaced
	GoodrichOrderStatusReady
	GoodrichOrderStatusPaid
)

type GoodrichPaymentMethod int

const (
	GoodrichPaymentMethodUnknown GoodrichPaymentMethod = iota // Default is unknown
	GoodrichPaymentMethodSwipe
	GoodrichPaymentMethodPoints
	GoodrichPaymentMethodCreditCard
	GoodrichPaymentMethodCash
)

// GoodrichOrder Schema
type GoodrichOrder struct {
	BaseSchema

	Status      GoodrichOrderStatus `json:"status"`
	PhoneNumber string              `json:"phoneNumber"`
	Date        string              `json:"date"`     // Format: 2006-01-02
	TimeSlot    string              `json:"timeSlot"` // Format: 15:04
	Notes       string              `json:"notes"`
	TotalPrice  float64             `json:"totalPrice"`
	ComboDeal   *bool               `gorm:"DEFAULT:false;not null" json:"comboDeal"`
	AdminNotes  string              `json:"adminNotes"`

	PaymentMethod GoodrichPaymentMethod `json:"paymentMethod"`
	IDNumber      *string               `json:"idNumber"`

	// You will need to make a custom format to store these in SQL, b/c a join table would be absolutely massive. I suggest doing a comma separated format like "id,id,id"
	ItemList string `json:"-"`
	// this means this field will be ignored in the DB. Use this to pull the menu items from ItemList into actual objects in the Model/Controller side.
	Items []*GoodrichMenuItem `json:"items" gorm:"-"`

	// Belongs to user
	UserID uint  `json:"userID"`
	User   *User `json:"user,omitempty"`
}

func (*GoodrichOrder) TableName() string {
	return "goodrich_orders"
}

func (o *GoodrichOrder) AfterFind(tx *gorm.DB) (err error) {
	if o.ItemList != "" {
		itemIdStrs := strings.Split(o.ItemList, ",")

		var questionFmt []string
		var itemIDs []interface{}
		for _, itemID := range itemIdStrs {
			questionFmt = append(questionFmt, "?")
			itemIDs = append(itemIDs, itemID)
		}

		var items []*GoodrichMenuItem
		newTx := tx.New().Model(&GoodrichMenuItem{}).
			Where("id IN ("+strings.Join(questionFmt, ",")+")", itemIDs...)

		err = newTx.Find(&items).Error
		if err != nil {
			return err
		}

		o.Items = items
	}
	return
}

func ValidateGoodrichPaymentMethod(m GoodrichPaymentMethod) bool {
	switch m {
	case GoodrichPaymentMethodSwipe, GoodrichPaymentMethodCreditCard, GoodrichPaymentMethodCash:
		return true
	}
	return false
}

func ValidateGoodrichOrderStatus(m GoodrichOrderStatus) bool {
	switch m {
	case GoodrichOrderStatusPlaced,
		GoodrichOrderStatusReady,
		GoodrichOrderStatusPaid:
		return true
	}
	return false
}
