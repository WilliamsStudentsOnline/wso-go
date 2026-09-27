package migrations

import (
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

var CreateBoardTables20260927001906 = &gormigrate.Migration{
	ID: "20260927001906_create_board_tables",
	Migrate: func(tx *gorm.DB) error {
		type BoardThread struct {
			models.BoardBaseSchema

			Type  string `gorm:"index:index_board_threads_on_type;not null" json:"type"`
			Title string `gorm:"not null" json:"title"`

			UserID     uint   `gorm:"index:index_board_threads_on_user_id;not null" json:"userID"`
			ExUserName string `json:"exUserName"`

			RepliesEnabled bool       `gorm:"not null;default:true" json:"repliesEnabled"`
			Resolved       *bool      `json:"resolved"`
			StartsAt       *time.Time `json:"startsAt"`
			EndsAt         *time.Time `json:"endsAt"`

			LastActive time.Time `gorm:"index:index_board_threads_on_last_active;not null" json:"lastActive"`
			FlagCount  int       `gorm:"not null;default:0" json:"flagCount"`
			ReplyCount int       `gorm:"not null;default:0" json:"replyCount"`
		}

		type BoardPost struct {
			models.BoardBaseSchema

			ThreadID uint `gorm:"index:index_board_posts_on_thread_id;not null" json:"threadID"`

			UserID     uint   `gorm:"index:index_board_posts_on_user_id;not null" json:"userID"`
			ExUserName string `json:"exUserName"`

			Content   string `gorm:"size:65535;not null" json:"content"`
			IsOP      bool   `gorm:"column:is_op;not null;default:false" json:"isOP"`
			FlagCount int    `gorm:"not null;default:0" json:"flagCount"`
		}

		type BoardRideMeta struct {
			models.BoardBaseSchema

			ThreadID uint `gorm:"unique_index:index_board_ride_meta_on_thread_id;not null" json:"threadID"`

			OfferingRide bool   `gorm:"not null" json:"offeringRide"`
			Source       string `gorm:"not null" json:"source"`
			Destination  string `gorm:"not null" json:"destination"`
		}

		type Flag struct {
			models.BoardBaseSchema

			TargetType string `gorm:"unique_index:idx_flags_target_user;index:idx_flags_target;not null" json:"targetType"`
			TargetID   uint   `gorm:"unique_index:idx_flags_target_user;index:idx_flags_target;not null" json:"targetID"`

			UserID uint    `gorm:"unique_index:idx_flags_target_user;index:idx_flags_user_id;not null" json:"userID"`
			Reason *string `gorm:"size:1024" json:"reason,omitempty"`
		}

		type BoardBackfillMap struct {
			models.BoardBaseSchema

			SourceTable string `gorm:"unique_index:idx_board_backfill_source;not null" json:"sourceTable"`
			SourceID    uint   `gorm:"unique_index:idx_board_backfill_source;not null" json:"sourceID"`
			ThreadID    uint   `gorm:"index:idx_board_backfill_thread;not null" json:"threadID"`
		}

		return tx.AutoMigrate(
			&BoardThread{},
			&BoardPost{},
			&BoardRideMeta{},
			&Flag{},
			&BoardBackfillMap{},
		).Error
	},
	Rollback: func(tx *gorm.DB) error {
		if err := tx.DropTable("board_backfill_map").Error; err != nil {
			return err
		}
		if err := tx.DropTable("flags").Error; err != nil {
			return err
		}
		if err := tx.DropTable("board_ride_meta").Error; err != nil {
			return err
		}
		if err := tx.DropTable("board_posts").Error; err != nil {
			return err
		}
		return tx.DropTable("board_threads").Error
	},
}
