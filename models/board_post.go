package models

import (
	"time"

	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// BoardPostModel models board posts.
type BoardPostModel struct {
	*BaseModel
}

func NewBoardPostModel(db *gorm.DB, log *zap.SugaredLogger) *BoardPostModel {
	return &BoardPostModel{BaseModel: NewBaseModel(db, log)}
}

func (m *BoardPostModel) GetPostByID(id uint, p *BoardPost) error {
	return m.DB.Preload("User").Preload("Thread").First(p, id).Error
}

// CreateReply creates a reply and updates thread last_active / reply_count.
func (m *BoardPostModel) CreateReply(threadID uint, p *BoardPost) error {
	tx := m.DB.Begin()
	if err := tx.Error; err != nil {
		return err
	}

	p.ThreadID = threadID
	p.IsOP = false
	if err := tx.Create(p).Error; err != nil {
		tx.Rollback()
		return err
	}

	now := time.Now()
	if err := tx.Model(NewBoardThreadByID(threadID)).Updates(map[string]interface{}{
		"last_active": now,
		"reply_count": gorm.Expr("reply_count + 1"),
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}
	return m.DB.Preload("User").First(p, p.ID).Error
}

func (m *BoardPostModel) UpdatePost(p *BoardPost) error {
	return m.DB.Save(p).Error
}

// DeletePost soft-deletes a post. If it is the OP, cascades to the whole thread.
// Otherwise decrements reply_count and refreshes last_active.
func (m *BoardPostModel) DeletePost(p *BoardPost) error {
	if p.IsOP {
		tm := NewBoardThreadModel(m.DB, m.log)
		thread := NewBoardThreadByID(p.ThreadID)
		if err := m.DB.First(thread, p.ThreadID).Error; err != nil {
			return err
		}
		return tm.DeleteThread(thread)
	}

	tx := m.DB.Begin()
	if err := tx.Error; err != nil {
		return err
	}

	if err := tx.Delete(p).Error; err != nil {
		tx.Rollback()
		return err
	}

	var recent BoardPost
	err := tx.Where("thread_id = ?", p.ThreadID).
		Order("created_at desc").
		First(&recent).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		tx.Rollback()
		return err
	}

	updates := map[string]interface{}{
		"reply_count": gorm.Expr("CASE WHEN reply_count > 0 THEN reply_count - 1 ELSE 0 END"),
	}
	if recent.ID != 0 {
		updates["last_active"] = recent.CreatedAt
	}
	if err := tx.Model(NewBoardThreadByID(p.ThreadID)).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
