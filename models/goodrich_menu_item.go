package models

import (
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// GoodrichMenuItem Model
type GoodrichMenuItemModel struct {
	*BaseModel
}

func NewGoodrichMenuItemModel(db *gorm.DB, log *zap.SugaredLogger) *GoodrichMenuItemModel {
	return &GoodrichMenuItemModel{
		BaseModel: NewBaseModel(db, log),
	}
}

type GetAllGoodrichMenuItemsOptions struct {
	// Get all goodrich menu items, rather than just available ones.
	// By default, this is false.
	All bool `json:"all" form:"all"`
}

func (p *GetAllGoodrichMenuItemsOptions) Order(db *gorm.DB) *gorm.DB {
	return db.Order("goodrich_menu_items.title ASC, goodrich_menu_items.type ASC", true)
}

func (p *GetAllGoodrichMenuItemsOptions) Filter(db *gorm.DB) *gorm.DB {
	if !p.All {
		return db.Where("goodrich_menu_items.available = ?", true)
	}
	return db
}

func (p *GetAllGoodrichMenuItemsOptions) Run(db *gorm.DB) *gorm.DB {
	return p.Filter(p.Order(db))
}

// Get all Goodrich menu items.
func (m *GoodrichMenuItemModel) GetAllGoodrichMenuItems(g *[]*GoodrichMenuItem, opts Options) (err error) {
	// Get menu
	db := m.DB.Model(&GoodrichMenuItem{})

	if opts != nil {
		db = opts.Run(db)
	}

	return db.Find(g).Error
}

// CreateMenuItem creates a menu item
func (m *GoodrichMenuItemModel) CreateMenuItem(g *GoodrichMenuItem) (err error) {
	err = m.DB.Create(g).Error
	if err != nil {
		return err
	}

	err = m.DB.Find(g).Error
	return
}

// GetMenuItemByID retrieves a menu item by ID
func (m *GoodrichMenuItemModel) GetMenuItemByID(id uint, g *GoodrichMenuItem) (err error) {
	err = m.DB.Where("id = ?", id).First(g).Error
	return
}

// UpdateMenuItem Updates the menu item
func (m *GoodrichMenuItemModel) UpdateMenuItem(b *GoodrichMenuItem) (err error) {
	err = m.DB.Save(b).Error
	return
}
