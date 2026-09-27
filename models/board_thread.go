package models

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// BoardThreadModel models board threads.
type BoardThreadModel struct {
	*BaseModel
}

func NewBoardThreadModel(db *gorm.DB, log *zap.SugaredLogger) *BoardThreadModel {
	return &BoardThreadModel{BaseModel: NewBaseModel(db, log)}
}

// GetBoardThreadsOptions controls feed listing.
type GetBoardThreadsOptions struct {
	// Cursor pagination: opaque "sortValue|id" from a previous page.
	Cursor *string `json:"cursor" form:"cursor"`
	Limit  *uint   `json:"limit" form:"limit"`

	// Sort: activity (default), created, startsAt
	Sort string `json:"sort" form:"sort"`
	// IncludeNull only applies when Sort=startsAt (nulls last).
	IncludeNull bool `json:"includeNull" form:"includeNull"`

	// Filters
	Types        []string `json:"type" form:"type[]"`
	Resolved     *bool    `json:"resolved" form:"resolved"`
	OfferingRide *bool    `json:"offeringRide" form:"offeringRide"`
	UserID       *uint    `json:"userID" form:"userID"`

	Preload []string `json:"preload" form:"preload[]"`
}

func (o *GetBoardThreadsOptions) sortMode() string {
	switch o.Sort {
	case "created", "startsAt":
		return o.Sort
	default:
		return "activity"
	}
}

func (o *GetBoardThreadsOptions) Order(db *gorm.DB) *gorm.DB {
	switch o.sortMode() {
	case "created":
		return db.Order("board_threads.created_at desc, board_threads.id desc", true)
	case "startsAt":
		if o.IncludeNull {
			return db.Order("board_threads.starts_at is null asc, board_threads.starts_at desc, board_threads.id desc", true)
		}
		return db.Order("board_threads.starts_at desc, board_threads.id desc", true)
	default:
		return db.Order("board_threads.last_active desc, board_threads.id desc", true)
	}
}

func (o *GetBoardThreadsOptions) Paginate(db *gorm.DB) *gorm.DB {
	db = o.Order(db)
	if o.Cursor != nil && *o.Cursor != "" {
		sortVal, id, err := parseBoardCursor(*o.Cursor)
		if err == nil {
			switch o.sortMode() {
			case "created":
				db = db.Where(
					"(board_threads.created_at < ?) OR (board_threads.created_at = ? AND board_threads.id < ?)",
					sortVal, sortVal, id,
				)
			case "startsAt":
				db = db.Where(
					"(board_threads.starts_at < ?) OR (board_threads.starts_at = ? AND board_threads.id < ?)",
					sortVal, sortVal, id,
				)
			default:
				db = db.Where(
					"(board_threads.last_active < ?) OR (board_threads.last_active = ? AND board_threads.id < ?)",
					sortVal, sortVal, id,
				)
			}
		}
	}
	if o.Limit != nil {
		db = db.Limit(*o.Limit)
	} else {
		db = db.Limit(50)
	}
	return db
}

func (o *GetBoardThreadsOptions) Preloader(db *gorm.DB) *gorm.DB {
	if o.Preload == nil {
		return db
	}
	if stringsContains(o.Preload, "user") {
		db = db.Preload("User")
	}
	if stringsContains(o.Preload, "ride") {
		db = db.Preload("RideMeta")
	}
	if stringsContains(o.Preload, "posts") {
		db = db.Preload("Posts", func(db *gorm.DB) *gorm.DB {
			return db.Order("board_posts.created_at asc")
		})
	}
	if stringsContains(o.Preload, "postsUsers") {
		db = db.Preload("Posts.User")
	}
	return db
}

func (o *GetBoardThreadsOptions) Filter(db *gorm.DB) *gorm.DB {
	now := time.Now()
	// Always exclude not-yet-started and expired.
	db = db.Where("board_threads.starts_at IS NULL OR board_threads.starts_at <= ?", now)
	db = db.Where("board_threads.ends_at IS NULL OR board_threads.ends_at > ?", now)

	if len(o.Types) > 0 {
		valid := make([]string, 0, len(o.Types))
		for _, t := range o.Types {
			if IsValidBoardThreadType(t) {
				valid = append(valid, t)
			}
		}
		if len(valid) > 0 {
			db = db.Where("board_threads.type IN (?)", valid)
		}
	}

	if o.Resolved != nil {
		if *o.Resolved {
			db = db.Where("board_threads.resolved = ?", true)
		} else {
			db = db.Where("board_threads.resolved IS NULL OR board_threads.resolved = ?", false)
		}
	}

	if o.UserID != nil {
		db = db.Where("board_threads.user_id = ?", *o.UserID)
	}

	if o.OfferingRide != nil {
		db = db.Joins("JOIN board_ride_meta ON board_ride_meta.thread_id = board_threads.id AND board_ride_meta.deleted_at IS NULL").
			Where("board_ride_meta.offering_ride = ?", *o.OfferingRide)
	}

	if o.sortMode() == "startsAt" && !o.IncludeNull {
		db = db.Where("board_threads.starts_at IS NOT NULL")
	}

	return db
}

func (o *GetBoardThreadsOptions) Run(db *gorm.DB) *gorm.DB {
	db = o.Filter(db)
	db = o.Paginate(db)
	db = o.Preloader(db)
	return db
}

func parseBoardCursor(cursor string) (time.Time, uint, error) {
	parts := strings.SplitN(cursor, "|", 2)
	if len(parts) != 2 {
		return time.Time{}, 0, fmt.Errorf("invalid cursor")
	}
	t, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		t, err = time.Parse(time.RFC3339, parts[0])
		if err != nil {
			return time.Time{}, 0, err
		}
	}
	id64, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return time.Time{}, 0, err
	}
	return t, uint(id64), nil
}

// BoardCursorForThread builds the next-page cursor for a thread under the given sort.
func BoardCursorForThread(t *BoardThread, sort string) string {
	var sortVal time.Time
	switch sort {
	case "created":
		sortVal = t.CreatedAt
	case "startsAt":
		if t.StartsAt != nil {
			sortVal = *t.StartsAt
		}
	default:
		sortVal = t.LastActive
	}
	return sortVal.UTC().Format(time.RFC3339Nano) + "|" + strconv.FormatUint(uint64(t.ID), 10)
}

func (m *BoardThreadModel) GetThreads(threads *[]*BoardThread, opts *GetBoardThreadsOptions) error {
	db := m.DB
	if opts != nil {
		db = opts.Run(db)
	}
	return db.Find(threads).Error
}

func (m *BoardThreadModel) CountThreads(opts *GetBoardThreadsOptions) (int, error) {
	db := m.DB.Model(&BoardThread{})
	if opts != nil {
		db = opts.Filter(db)
	}
	var count int
	err := db.Count(&count).Error
	return count, err
}

type GetBoardThreadByIDOptions struct {
	Preload []string `json:"preload" form:"preload[]"`
}

func (o *GetBoardThreadByIDOptions) Preloader(db *gorm.DB) *gorm.DB {
	if o.Preload == nil {
		db = db.Preload("User").Preload("RideMeta").Preload("Posts", func(db *gorm.DB) *gorm.DB {
			return db.Order("board_posts.created_at asc")
		}).Preload("Posts.User")
		return db
	}
	if stringsContains(o.Preload, "user") {
		db = db.Preload("User")
	}
	if stringsContains(o.Preload, "ride") {
		db = db.Preload("RideMeta")
	}
	if stringsContains(o.Preload, "posts") {
		db = db.Preload("Posts", func(db *gorm.DB) *gorm.DB {
			return db.Order("board_posts.created_at asc")
		})
	}
	if stringsContains(o.Preload, "postsUsers") {
		db = db.Preload("Posts.User")
	}
	return db
}

func (o *GetBoardThreadByIDOptions) Run(db *gorm.DB) *gorm.DB {
	return o.Preloader(db)
}

func (m *BoardThreadModel) GetThreadByID(id uint, t *BoardThread, opts *GetBoardThreadByIDOptions) error {
	db := m.DB
	if opts == nil {
		opts = &GetBoardThreadByIDOptions{}
	}
	db = opts.Run(db)
	err := db.First(t, id).Error
	if err != nil {
		return err
	}
	hydrateThreadBody(t)
	return nil
}

func hydrateThreadBody(t *BoardThread) {
	for _, p := range t.Posts {
		if p.IsOP {
			t.Body = p.Content
			return
		}
	}
	if len(t.Posts) > 0 {
		t.Body = t.Posts[0].Content
	}
}

func (m *BoardThreadModel) DoesThreadExist(id uint) (bool, error) {
	var count int
	err := m.DB.Model(&BoardThread{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// CreateThread creates a thread with its OP post and optional ride meta in one transaction.
func (m *BoardThreadModel) CreateThread(t *BoardThread, body string, ride *BoardRideMeta) error {
	tx := m.DB.Begin()
	if err := tx.Error; err != nil {
		return err
	}

	if t.LastActive.IsZero() {
		t.LastActive = time.Now()
	}
	t.ReplyCount = 0
	// Capture before Create: GORM v1 omits false bools and may reload DB defaults into t.
	wantRepliesEnabled := t.RepliesEnabled

	if err := tx.Create(t).Error; err != nil {
		tx.Rollback()
		return err
	}
	if !wantRepliesEnabled {
		if err := tx.Exec("UPDATE board_threads SET replies_enabled = ? WHERE id = ?", false, t.ID).Error; err != nil {
			tx.Rollback()
			return err
		}
		t.RepliesEnabled = false
	}

	op := &BoardPost{
		ThreadID:   t.ID,
		UserID:     t.UserID,
		ExUserName: t.ExUserName,
		Content:    body,
		IsOP:       true,
	}
	if err := tx.Create(op).Error; err != nil {
		tx.Rollback()
		return err
	}

	if ride != nil {
		ride.ThreadID = t.ID
		if err := tx.Create(ride).Error; err != nil {
			tx.Rollback()
			return err
		}
		t.RideMeta = ride
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	t.Posts = []*BoardPost{op}
	t.Body = body
	return m.DB.Preload("User").Preload("RideMeta").First(t, t.ID).Error
}

func (m *BoardThreadModel) UpdateThread(t *BoardThread) error {
	return m.DB.Save(t).Error
}

func (m *BoardThreadModel) UpdateOPContent(threadID uint, content string) error {
	return m.DB.Model(&BoardPost{}).
		Where("thread_id = ? AND is_op = ?", threadID, true).
		Update("content", content).Error
}

// DeleteThread soft-deletes the thread, its posts, and ride meta.
func (m *BoardThreadModel) DeleteThread(t *BoardThread) error {
	tx := m.DB.Begin()
	if err := tx.Error; err != nil {
		return err
	}

	if err := tx.Where("thread_id = ?", t.ID).Delete(&BoardPost{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("thread_id = ?", t.ID).Delete(&BoardRideMeta{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Delete(t).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}
