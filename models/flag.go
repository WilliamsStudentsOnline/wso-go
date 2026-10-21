package models

import (
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// FlagModel models content flags.
type FlagModel struct {
	*BaseModel
}

func NewFlagModel(db *gorm.DB, log *zap.SugaredLogger) *FlagModel {
	return &FlagModel{BaseModel: NewBaseModel(db, log)}
}

type GetFlagsOptions struct {
	TargetType *string `json:"targetType" form:"targetType"`
	Limit      *uint   `json:"limit" form:"limit"`
	Offset     *uint   `json:"offset" form:"offset"`
}

func (o *GetFlagsOptions) Filter(db *gorm.DB) *gorm.DB {
	if o.TargetType != nil && *o.TargetType != "" {
		db = db.Where("flags.target_type = ?", *o.TargetType)
	}
	return db
}

func (o *GetFlagsOptions) Run(db *gorm.DB) *gorm.DB {
	db = o.Filter(db)
	db = db.Order("flags.created_at desc", true)
	if o.Limit != nil {
		db = db.Limit(*o.Limit)
		if o.Offset != nil {
			db = db.Offset(*o.Offset)
		}
	}
	return db.Preload("User")
}

func (m *FlagModel) ListFlags(flags *[]*Flag, opts *GetFlagsOptions) error {
	db := m.DB
	if opts != nil {
		db = opts.Run(db)
	}
	return db.Find(flags).Error
}

func (m *FlagModel) CountFlags(opts *GetFlagsOptions) (int, error) {
	db := m.DB.Model(&Flag{})
	if opts != nil {
		db = opts.Filter(db)
	}
	var count int
	err := db.Count(&count).Error
	return count, err
}

// CreateFlag inserts a flag and increments the target's flag_count.
// Returns false if a duplicate flag already exists.
func (m *FlagModel) CreateFlag(f *Flag) (created bool, err error) {
	var existing Flag
	err = m.DB.Where("target_type = ? AND target_id = ? AND user_id = ?", f.TargetType, f.TargetID, f.UserID).
		First(&existing).Error
	if err == nil {
		return false, nil
	}
	if !gorm.IsRecordNotFoundError(err) {
		return false, err
	}

	tx := m.DB.Begin()
	if err = tx.Error; err != nil {
		return false, err
	}

	if err = tx.Create(f).Error; err != nil {
		tx.Rollback()
		return false, err
	}

	if err = incrementFlagCount(tx, f.TargetType, f.TargetID, 1); err != nil {
		tx.Rollback()
		return false, err
	}

	if err = tx.Commit().Error; err != nil {
		return false, err
	}
	return true, nil
}

// ClearFlagsForTarget deletes all flags for a target and resets flag_count to 0.
func (m *FlagModel) ClearFlagsForTarget(targetType string, targetID uint) error {
	tx := m.DB.Begin()
	if err := tx.Error; err != nil {
		return err
	}

	if err := tx.Where("target_type = ? AND target_id = ?", targetType, targetID).
		Delete(&Flag{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := setFlagCount(tx, targetType, targetID, 0); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func incrementFlagCount(tx *gorm.DB, targetType string, targetID uint, delta int) error {
	switch targetType {
	case FlagTargetBoardThread:
		return tx.Model(NewBoardThreadByID(targetID)).
			Update("flag_count", gorm.Expr("flag_count + ?", delta)).Error
	case FlagTargetBoardPost:
		return tx.Model(NewBoardPostByID(targetID)).
			Update("flag_count", gorm.Expr("flag_count + ?", delta)).Error
	default:
		return nil
	}
}

func setFlagCount(tx *gorm.DB, targetType string, targetID uint, count int) error {
	switch targetType {
	case FlagTargetBoardThread:
		return tx.Model(NewBoardThreadByID(targetID)).Update("flag_count", count).Error
	case FlagTargetBoardPost:
		return tx.Model(NewBoardPostByID(targetID)).Update("flag_count", count).Error
	default:
		return nil
	}
}
