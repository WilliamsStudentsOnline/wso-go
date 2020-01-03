package models

import (
	"time"

	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// Discussion Model
type DiscussionModel struct {
	*BaseModel
}

func NewDiscussionModel(db *gorm.DB, log *zap.SugaredLogger) *DiscussionModel {
	return &DiscussionModel{
		BaseModel: NewBaseModel(db, log),
	}
}

type GetAllDiscussionsOptions struct {
	// Pagination
	Start *time.Time `json:"start" form:"start"`
	// Offset is ignored unless limit is supplied
	Offset *uint `json:"offset" form:"offset"`
	Limit  *uint `json:"limit" form:"limit"`

	// What to preload (user, posts, postsUsers)
	Preload []string `json:"preload" form:"preload[]"`

	// If true get the last/latest post of the discussion.
	GetLastPost *bool `json:"getLastPost" form:"getLastPost"`
}

func (o *GetAllDiscussionsOptions) Order(db *gorm.DB) *gorm.DB {
	return db.Order("discussions.last_active desc", true)
}

// Pagination starts at most recent and goes down from there
func (o *GetAllDiscussionsOptions) Paginate(db *gorm.DB) *gorm.DB {
	db = o.Order(db)
	if o.Start != nil {
		db = db.Where("discussions.last_active < ?", *o.Start)
	}
	if o.Limit != nil {
		db = db.Limit(*o.Limit)
		if o.Offset != nil {
			db = db.Offset(*o.Offset)
		}
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

func (o *GetAllDiscussionsOptions) Post(db *gorm.DB, d []*Discussion) error {
	// If getLastPost = true and we aren't preloading posts or postsUsers
	if o.GetLastPost != nil && *o.GetLastPost &&
		!stringsContains(o.Preload, "posts") && !stringsContains(o.Preload, "postsUsers") {
		for i := range d {
			p := Post{}
			err := db.New().Model(&Post{}).
				Where("discussion_id = ?", d[i].ID).
				Order("created_at desc").
				Preload("User").
				First(&p).Error
			// Only really error iff it is not a 404 error, otherwise just ignore it
			if err != nil {
				if gorm.IsRecordNotFoundError(err) {
					continue
				}
				return err
			}

			d[i].Posts = []*Post{&p}
		}
	}
	return nil
}

func (m *DiscussionModel) GetAllDiscussions(p *[]*Discussion, opts *GetAllDiscussionsOptions) (err error) {
	db := m.DB
	if opts != nil {
		db = opts.Run(db)
	}
	err = db.Find(p).Error
	if err != nil {
		return err
	}

	if opts != nil {
		err = opts.Post(db, *p)
	}
	return
}

func (m *DiscussionModel) CountAllDiscussions() (count int, err error) {
	db := m.DB.Model(&Discussion{})
	err = db.Count(&count).Error
	return
}

type GetDiscussionByIDOptions struct {
	// What to preload (user, posts, postsUsers)
	Preload []string `json:"preload" form:"preload[]"`
}

// Preload specifically allowed parts if requested
func (o *GetDiscussionByIDOptions) Preloader(db *gorm.DB) *gorm.DB {
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

func (o *GetDiscussionByIDOptions) Run(db *gorm.DB) *gorm.DB {
	db = o.Preloader(db)

	return db
}

func (m *DiscussionModel) GetDiscussionByID(id uint, p *Discussion, opts Options) (err error) {
	db := m.DB
	if opts != nil {
		db = opts.Run(db)
	}
	err = db.First(p, id).Error
	return
}

func (m *DiscussionModel) GetDiscussionByIDFullPreload(id uint, p *Discussion) (err error) {
	err = m.GetDiscussionByID(id, p, &GetDiscussionByIDOptions{
		Preload: []string{"user", "posts", "postsUsers"},
	})
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
