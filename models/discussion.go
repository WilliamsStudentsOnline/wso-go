package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

// Discussion Model
type DiscussionModel struct {
	*BaseModel
}

func NewDiscussionModel(db *gorm.DB) *DiscussionModel {
	return &DiscussionModel{
		BaseModel: NewBaseModel(db),
	}
}

type GetAllDiscussionsOptions struct {
	// Pagination
	Offset *time.Time `json:"offset" form:"offset"`
	Limit  *uint      `json:"limit" form:"limit"`

	// What to preload (user, posts, postsUsers)
	Preload []string `json:"preload" form:"preload[]"`
}

func (o *GetAllDiscussionsOptions) Order(db *gorm.DB) *gorm.DB {
	return db.Order("discussions.last_active desc", true)
}

// Pagination starts at most recent and goes down from there
func (o *GetAllDiscussionsOptions) Paginate(db *gorm.DB) *gorm.DB {
	db = o.Order(db)
	if o.Offset != nil {
		db = db.Where("discussions.last_active < ?", *o.Offset)
	}
	if o.Limit != nil {
		db = db.Limit(*o.Limit)
	}
	return db
}

// Preload specifically allowed parts if requested
func (o *GetAllDiscussionsOptions) Preloader(db *gorm.DB) *gorm.DB {
	if o.Preload == nil {
		return db
	}

	if stringsContains(o.Preload, "user") {
		db = db.Preload("User")
	}
	if stringsContains(o.Preload, "posts") {
		db = db.Preload("Posts")
	}
	if stringsContains(o.Preload, "postsUsers") {
		db = db.Preload("Posts.User")
	}

	return db
}

func (o *GetAllDiscussionsOptions) Run(db *gorm.DB) *gorm.DB {
	db = o.Paginate(db)
	db = o.Preloader(db)

	return db
}

func (m *DiscussionModel) GetAllDiscussions(p *[]*Discussion, opts Options) (err error) {
	db := m.DB
	if opts != nil {
		db = opts.Run(db)
	}
	err = db.Find(p).Error
	return
}

func (m *DiscussionModel) GetDiscussionByID(id uint, p *Discussion) (err error) {
	err = m.DB.Preload("User").Preload("Posts").Preload("Posts.User").First(p, id).Error
	return
}

func (m *DiscussionModel) DoesDiscussionExist(id uint) (exists bool, err error) {
	var count int
	err = m.DB.Model(&Discussion{}).Where("id = ?", id).Count(&count).Error
	exists = count > 0
	return
}

func (m *DiscussionModel) CreateDiscussion(p *Discussion) (err error) {
	err = m.DB.Create(p).Error
	if err != nil {
		return err
	}

	err = m.DB.Find(p).Error
	return
}

// Deletes discussion and all of its posts (soft delete).
func (m *DiscussionModel) DeleteDiscussion(p *Discussion) (err error) {
	tx := m.DB.Begin()

	if err = tx.Error; err != nil {
		return
	}

	err = tx.Where("posts.discussion_id = ?", p.ID).Delete(&Post{}).Error
	if err != nil {
		tx.Rollback()
		return
	}

	err = tx.Delete(p).Error
	if err != nil {
		tx.Rollback()
		return
	}

	err = tx.Commit().Error
	return
}
