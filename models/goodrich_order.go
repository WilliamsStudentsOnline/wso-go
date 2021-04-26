package models

import (
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// GoodrichOrder Model
type GoodrichOrderModel struct {
	*BaseModel
}

func NewGoodrichOrderModel(db *gorm.DB, log *zap.SugaredLogger) *GoodrichOrderModel {
	return &GoodrichOrderModel{
		BaseModel: NewBaseModel(db, log),
	}
}

type GetAllGoodrichOrdersOptions struct {
	Statuses *[]GoodrichOrderStatus `json:"statuses" form:"statuses[]"`

	UserID *uint `json:"userID" form:"userID"`

	Sort *string `json:"sort" form:"sort"`

	// Pagination:
	// Offset is ignored unless limit is supplied
	Offset *uint `json:"offset" form:"offset"`
	Limit  *uint `json:"limit" form:"limit"`
}

func (o *GetAllGoodrichOrdersOptions) Order(db *gorm.DB) *gorm.DB {
	if o.Sort != nil {
		switch *o.Sort {
		case "created_at":
			return db.Order("goodrich_orders.created_at desc", true)
		case "preferred_time":
			return db.Order("goodrich_orders.preferred_time desc", true)
		case "estimated_time":
			return db.Order("goodrich_orders.estimated_time desc", true)
		}
	}
	return db.Order("goodrich_orders.created_at desc", true)
}

// Pagination is by offset :/
func (o *GetAllGoodrichOrdersOptions) Paginate(db *gorm.DB) *gorm.DB {
	db = o.Order(db)
	if o.Limit != nil {
		db = db.Limit(*o.Limit)
		if o.Offset != nil {
			db = db.Offset(*o.Offset)
		}
	}
	return db
}

// Preload specifically allowed parts if requested
func (o *GetAllGoodrichOrdersOptions) Preloader(db *gorm.DB) *gorm.DB {
	return db.Preload("User")
}

func (o *GetAllGoodrichOrdersOptions) Filter(db *gorm.DB) *gorm.DB {
	if o.Statuses != nil {
		for _, s := range *o.Statuses {
			db = db.Or("goodrich_orders.status = ?", s)
		}
	}

	if o.UserID != nil {
		db = db.Where("goodrich_orders.user_id = ?", *o.UserID)
	}

	return db
}

func (o *GetAllGoodrichOrdersOptions) Run(db *gorm.DB) *gorm.DB {
	db = o.Paginate(db)
	db = o.Preloader(db)
	db = o.Filter(db)

	return db
}

func (m *GoodrichOrderModel) GetAllGoodrichOrders(o *[]*GoodrichOrder, opts Options) (err error) {
	db := m.DB
	if opts != nil {
		db = opts.Run(db)
	}
	err = db.Find(o).Error
	return
}

// Get all Goodrich user orders.
func (m *GoodrichOrderModel) GetGoodrichUserOrders(o *[]*GoodrichOrder, userID uint) (err error) {
	db := m.DB.Model(&GoodrichOrder{})

	return db.Where("user_id = ?", userID).Find(o).Error
}

// Get Goodrich order
func (m *GoodrichOrderModel) GetGoodrichOrder(o *GoodrichOrder, id uint, preloadUser bool) (err error) {
	db := m.DB.Model(&GoodrichOrder{})

	if preloadUser {
		db.Preload("User")
	}

	return db.Where("id = ?", id).First(o).Error
}

// CreateOrder creates a menu item
func (m *GoodrichOrderModel) CreateOrder(o *GoodrichOrder) (err error) {
	err = m.DB.Create(o).Error
	if err != nil {
		return err
	}

	err = m.DB.Find(o).Error
	return
}

// UpdateOrder Updates the order
func (m *GoodrichOrderModel) UpdateOrder(o *GoodrichOrder) (err error) {
	err = m.DB.Save(o).Error
	return
}
