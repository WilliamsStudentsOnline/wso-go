package models

import (
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type FeedbackModel struct {
	*BaseModel
}

func NewFeedbackModel(db *gorm.DB, log *zap.SugaredLogger) *FeedbackModel {
	return &FeedbackModel{
		BaseModel: NewBaseModel(db, log),
	}
}

type GetAllFeedbackOptions struct {
	Offset *uint `json:"offset" form:"offset"`
	Limit  *uint `json:"limit" form:"limit"`

	Priority *FeedbackPriority `json:"priority" form:"priority" enums:",none,priority_low,priority_med,priority_high,resolved,dismissed"`

	Preload []string `json:"preload" form:"preload[]" enums:",user"`
}

func (o *GetAllFeedbackOptions) Order(db *gorm.DB) *gorm.DB {
	return db.Order("feedback.created_at DESC", true)
}

func (o *GetAllFeedbackOptions) Paginate(db *gorm.DB) *gorm.DB {
	db = o.Order(db)
	if o.Limit != nil {
		db = db.Limit(*o.Limit)
		if o.Offset != nil {
			db = db.Offset(*o.Offset)
		}
	}
	return db
}

func (o *GetAllFeedbackOptions) Preloader(db *gorm.DB) *gorm.DB {
	if o.Preload == nil {
		return db
	}
	m := NewFeedbackModel(nil, nil)
	if stringsContains(o.Preload, "user") {
		db = m.preloadUser(db)
	}
	return db
}

func (o *GetAllFeedbackOptions) Run(db *gorm.DB) *gorm.DB {
	db = o.Preloader(db)
	m := NewFeedbackModel(db.New(), nil)

	if o.Priority != nil {
		db = m.withPriority(*o.Priority)(db)
	}
	return db
}

func (m *FeedbackModel) preloadUser(db *gorm.DB) *gorm.DB {
	return db.Preload("User")
}

func (m *FeedbackModel) withPriority(priority FeedbackPriority) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("feedback.priority = ?", priority)
	}
}

func (m *FeedbackModel) GetAllFeedback(items *[]*Feedback, opts *GetAllFeedbackOptions) error {
	db := m.DB
	if opts != nil {
		db = opts.Paginate(db)
		db = opts.Run(db)
	} else {
		db = db.Order("feedback.created_at DESC", true)
	}
	return db.Find(items).Error
}

func (m *FeedbackModel) CountAllFeedback(opts *GetAllFeedbackOptions) (count int, err error) {
	db := m.DB.Model(&Feedback{})
	if opts != nil {
		db = opts.Run(db)
	}
	err = db.Count(&count).Error
	return
}

func (m *FeedbackModel) GetFeedbackByID(id uint, item *Feedback, preload ...string) error {
	opts := &GetAllFeedbackOptions{Preload: preload}
	return opts.Preloader(m.DB).First(item, id).Error
}

func (m *FeedbackModel) CreateFeedback(item *Feedback) error {
	if item.Priority == FeedbackPriorityUndefined {
		item.Priority = FeedbackPriorityNone
	}
	return m.DB.Create(item).Error
}

func (m *FeedbackModel) UpdateFeedback(id uint, updates map[string]interface{}) error {
	res := m.DB.Model(&Feedback{}).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (m *FeedbackModel) SoftDeleteFeedbackByID(id uint) error {
	res := m.DB.Where("id = ?", id).Delete(&Feedback{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
