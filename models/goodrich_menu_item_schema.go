package models

import (
	"strings"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

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
