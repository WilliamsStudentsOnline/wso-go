package models

// BoardBackfillMap tracks which legacy rows have been copied into board_threads
// so the backfill job is idempotent.
type BoardBackfillMap struct {
	BoardBaseSchema

	SourceTable string `gorm:"unique_index:idx_board_backfill_source;not null" json:"sourceTable"`
	SourceID    uint   `gorm:"unique_index:idx_board_backfill_source;not null" json:"sourceID"`
	ThreadID    uint   `gorm:"index:idx_board_backfill_thread;not null" json:"threadID"`
}

func (*BoardBackfillMap) TableName() string {
	return "board_backfill_map"
}

const (
	BoardBackfillSourceDiscussions   = "discussions"
	BoardBackfillSourceBulletins     = "bulletins"
	BoardBackfillSourceBulletinRides = "bulletin_rides"
)
