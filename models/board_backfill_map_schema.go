package models

// BoardBackfillMap tracks which legacy rows have been copied into board_threads
// so the backfill job is idempotent
type BoardBackfillMap struct {
	BaseSchema

	SourceTable string `gorm:"unique_index:idx_backfill_map_board_source;not null" json:"sourceTable"`
	SourceID    uint   `gorm:"unique_index:idx_backfill_map_board_source;not null" json:"sourceID"`
	ThreadID    uint   `gorm:"index:idx_backfill_map_board_thread;not null" json:"threadID"`
}

func (*BoardBackfillMap) TableName() string {
	return "backfill_map_board"
}

const (
	BoardBackfillSourceDiscussions   = "discussions"
	BoardBackfillSourceBulletins     = "bulletins"
	BoardBackfillSourceBulletinRides = "bulletin_rides"
	// Posts store the board_posts.id in ThreadID (no separate post_id column)
	BoardBackfillSourcePosts = "posts"
)
