package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

// Post Model
type PostModel struct {
	*BaseModel
}

func NewPostModel(db *gorm.DB) *PostModel {
	return &PostModel{
		BaseModel: NewBaseModel(db),
	}
}

type GetPostsByDiscussionOptions struct {
	// Pagination
	Start  *time.Time `json:"start" form:"start"`
	Offset *uint      `json:"offset" form:"offset"`
	Limit  *uint      `json:"limit" form:"limit"`

	// What to preload (user, discussion)
	Preload []string `json:"preload" form:"preload[]"`
}

func (p *GetPostsByDiscussionOptions) Order(db *gorm.DB) *gorm.DB {
	return db.Order("posts.created_at desc", true)
}

// Pagination starts at most recent and goes down from there
func (p *GetPostsByDiscussionOptions) Paginate(db *gorm.DB) *gorm.DB {
	db = p.Order(db)
	if p.Start != nil {
		db = db.Where("posts.created_at < ?", *p.Start)
	}
	if p.Offset != nil {
		db = db.Offset(*p.Offset)
	}
	if p.Limit != nil {
		db = db.Limit(*p.Limit)
	}
	return db
}

// Preload specifically allowed parts if requested
func (p *GetPostsByDiscussionOptions) Preloader(db *gorm.DB) *gorm.DB {
	if p.Preload == nil {
		return db
	}

	if stringsContains(p.Preload, "user") {
		db = db.Preload("User")
	}
	if stringsContains(p.Preload, "discussion") {
		db = db.Preload("Discussion")
	}

	return db
}

func (p *GetPostsByDiscussionOptions) Run(db *gorm.DB) *gorm.DB {
	db = p.Paginate(db)
	db = p.Preloader(db)

	return db
}

func (m *PostModel) GetPostsByDiscussion(discussionID uint, p *[]*Post, opts Options) (err error) {
	db := m.DB
	if opts != nil {
		db = opts.Run(db)
	}
	err = db.Where("posts.discussion_id = ?", discussionID).Find(p).Error
	return
}

func (m *PostModel) GetPostByID(id uint, p *Post) (err error) {
	err = m.DB.Preload("User").Preload("Discussion").First(p, id).Error
	return
}

func (m *PostModel) CreatePostByDiscussion(discussionID uint, p *Post) (err error) {
	tx := m.DB.Begin()

	if err = tx.Error; err != nil {
		return
	}

	// Create post
	err = tx.Create(p).Error
	if err != nil {
		tx.Rollback()
		return
	}

	// Update discussion to have new last active
	err = tx.Model(NewDiscussionByID(discussionID)).Update("last_active", time.Now()).Error
	if err != nil {
		tx.Rollback()
		return
	}

	// Commit changes
	if err = tx.Commit().Error; err != nil {
		return
	}

	err = m.DB.Find(p).Error
	return
}

func (m *PostModel) UpdatePost(p *Post) (err error) {
	err = m.DB.Save(p).Error
	return
}

func (m *PostModel) DeletePost(p *Post) (err error) {
	// First, get the new most recent post (if it doesn't exist, do something else):
	var recentPost Post
	err = m.DB.
		Where("posts.discussion_id = ?", p.DiscussionID).
		Not("posts.id = ?", p.ID).
		Order("posts.created_at").
		First(&recentPost).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return
	}

	// Do the database modifications
	tx := m.DB.Begin()

	if err = tx.Error; err != nil {
		return
	}

	// Delete post
	err = tx.Delete(p).Error
	if err != nil {
		tx.Rollback()
		return
	}

	// Update discussion to have old last active if it exists
	if recentPost.ID != 0 {
		err = tx.Model(NewDiscussionByID(p.DiscussionID)).Update("last_active", recentPost.CreatedAt).Error
	} else {
		// If post was the last one in the discussion, delete the discussion.
		err = tx.Delete(NewDiscussionByID(p.DiscussionID)).Error
	}
	if err != nil {
		tx.Rollback()
		return
	}

	// Commit changes
	err = tx.Commit().Error
	return
}
