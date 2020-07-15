package models

type MenuItem struct {
	BaseSchema
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Available   bool    `json:"available"` // false if item is out of stock
}

func (*MenuItem) TableName() string { // maybe?
	return "goodrich_menu_items"
}
