package models

type MenuItem struct {
	BaseSchema
	Title       string
	Description string
	Price       float64
	Available   bool // false if item is out of stock
}

func (*MenuItem) TableName() string { // maybe?
	return "menu_items"
}
