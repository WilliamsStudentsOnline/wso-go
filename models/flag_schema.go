package models

// Flag target types (shared across services).
const (
	FlagTargetBoardThread   = "board_thread"
	FlagTargetBoardPost     = "board_post"
	FlagTargetFactrakSurvey = "factrak_survey"
)

// Flag records a user flagging a piece of content.
type Flag struct {
	BoardBaseSchema

	TargetType string `gorm:"unique_index:idx_flags_target_user;index:idx_flags_target;not null" json:"targetType"`
	TargetID   uint   `gorm:"unique_index:idx_flags_target_user;index:idx_flags_target;not null" json:"targetID"`

	UserID uint  `gorm:"unique_index:idx_flags_target_user;index:idx_flags_user_id;not null" json:"userID"`
	User   *User `json:"user,omitempty"`

	Reason *string `gorm:"size:1024" json:"reason,omitempty"`
}

func (*Flag) TableName() string {
	return "flags"
}
