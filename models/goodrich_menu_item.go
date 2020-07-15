package models

import (
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)


// menu item model
type MenuItemModel struct {
	*BaseModel
}

// new menu item model
func NewMenuItemModel(db *gorm.DB, log *zap.SugaredLogger) *MenuItemModel {
	return &MenuItemModel{
		BaseModel: NewBaseModel(db, log),
	}
}


// returns err after passing bool and menu slice
func (m *MenuItemModel) ListMenuItems(includeUnavailable bool, menu *[]*MenuItem) (err error) {
	//var menu []MenuItem
	//var err error
	db := m.DB
	if includeUnavailable {
		// includeUnavailable = true, returns items available and unavailable
		err = db.Find(menu).Error

	} else {
		// includeUnavailable = false, return only items that are available
		err = db.Where("goodrich_menu_items.available = ?", true).Find(menu).Error
	}
	return
}

// get menu item by item ID by passing item ID and var menuItem
func (m *MenuItemModel) GetMenuItem(itemID uint, menuItem *MenuItem) (err error) {
	err = m.DB.First(menuItem, itemID).Error
	return
}


func (m *MenuItemModel) CreateMenuItem(item *MenuItem) (err error) {
	db := m.DB
	err = db.Create(item).Error
	return
}

func (m *MenuItemModel) UpdateMenuItem(itemID uint, item *MenuItem) (err error) {
	db := m.DB
	err = db.Save(item).Error
	return
}
