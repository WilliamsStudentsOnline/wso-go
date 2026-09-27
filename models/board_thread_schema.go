package models

import "time"

// Board thread types
const (
	BoardThreadTypeDiscussion   = "discussion"
	BoardThreadTypeAnnouncement = "announcement"
	BoardThreadTypeJob          = "job"
	BoardThreadTypeExchange     = "exchange"
	BoardThreadTypeLostAndFound = "lostAndFound"
	BoardThreadTypeRide         = "ride"
)

// BoardThread is a unified campus board thread (discussions, listings, rides, etc.).
type BoardThread struct {
	BoardBaseSchema

	Type  string `gorm:"index:index_board_threads_on_type;not null" json:"type"`
	Title string `gorm:"not null" json:"title"`

	UserID     uint   `gorm:"index:index_board_threads_on_user_id;not null" json:"userID"`
	User       *User  `json:"user,omitempty"`
	ExUserName string `json:"exUserName"`

	RepliesEnabled bool       `gorm:"not null;default:true" json:"repliesEnabled"`
	Resolved       *bool      `json:"resolved"`
	StartsAt       *time.Time `json:"startsAt"`
	EndsAt         *time.Time `json:"endsAt"`

	LastActive time.Time `gorm:"index:index_board_threads_on_last_active;not null" json:"lastActive"`
	FlagCount  int       `gorm:"not null;default:0" json:"flagCount"`
	ReplyCount int       `gorm:"not null;default:0" json:"replyCount"`

	// Associations
	Posts    []*BoardPost   `gorm:"foreignkey:ThreadID;association_foreignkey:ID" json:"posts,omitempty"`
	RideMeta *BoardRideMeta `gorm:"foreignkey:ThreadID;association_foreignkey:ID" json:"ride,omitempty"`
	Body     string         `gorm:"-" json:"body,omitempty"` // hydrated from OP for API convenience
}

func (*BoardThread) TableName() string {
	return "board_threads"
}

func NewBoardThreadByID(id uint) *BoardThread {
	return &BoardThread{
		BoardBaseSchema: BoardBaseSchema{ID: id},
	}
}

func (t *BoardThread) BeforeCreate() error {
	if t.LastActive.IsZero() {
		t.LastActive = time.Now()
	}
	return nil
}

func IsValidBoardThreadType(typ string) bool {
	switch typ {
	case BoardThreadTypeDiscussion, BoardThreadTypeAnnouncement, BoardThreadTypeJob,
		BoardThreadTypeExchange, BoardThreadTypeLostAndFound, BoardThreadTypeRide:
		return true
	default:
		return false
	}
}
