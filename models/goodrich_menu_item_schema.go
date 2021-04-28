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
}

func (*GoodrichMenuItem) TableName() string {
	return "goodrich_menu_items"
}
