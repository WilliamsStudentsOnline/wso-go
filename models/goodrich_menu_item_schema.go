package models

// GoodrichMenuItem Schema
type GoodrichMenuItem struct {
	BaseSchema
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Type        string  `json:"type"`
	Category    string  `json:"category"`
	// false if item is out of stock
	Available bool `json:"available"`
	// Number of this menu item left; goes down with every order
	// quantity limit describe if there is a limit on quantity
	QuantityLimit bool `gorm:"DEFAULT:false;not null" json:"quantityLimit"`
	Quantity      *int `json:"quantity"`
}

func (*GoodrichMenuItem) TableName() string {
	return "goodrich_menu_items"
}
