package models

// BoardPost is a post within a board thread (OP or reply).
type BoardPost struct {
	BoardBaseSchema

	ThreadID uint         `gorm:"index:index_board_posts_on_thread_id;not null" json:"threadID"`
	Thread   *BoardThread `json:"thread,omitempty"`

	UserID     uint   `gorm:"index:index_board_posts_on_user_id;not null" json:"userID"`
	User       *User  `json:"user,omitempty"`
	ExUserName string `json:"exUserName"`

	Content   string `gorm:"size:65535;not null" json:"content"`
	IsOP      bool   `gorm:"column:is_op;not null;default:false" json:"isOP"`
	FlagCount int    `gorm:"not null;default:0" json:"flagCount"`
}

func (*BoardPost) TableName() string {
	return "board_posts"
}

func NewBoardPostByID(id uint) *BoardPost {
	return &BoardPost{
		BoardBaseSchema: BoardBaseSchema{ID: id},
	}
}
