package models

import "time"

// Post Schema
type Post struct {
	BaseSchema

	// Belongs to user
	UserID uint  `gorm:"index:index_posts_on_user_id;not null" json:"userID"`
	User   *User `json:"user,omitempty"`

	// Belongs to discussion
	DiscussionID uint        `gorm:"index:index_posts_on_discussion_id;not null;" json:"discussionID"`
	Discussion   *Discussion `json:"discussion,omitempty"`

	Content string `gorm:"size:65535;not null" json:"content"`

	// This is for when a user is deleted, but we still want to keep the info.
	// Ideally, we would only soft delete users but, we need to maintain the previous WSO's
	// standards.
	ExUserName string `json:"exUserName"`

	// Pass the created time: not looked at by GORM
	CreatedTime time.Time `gorm:"-" json:"createdTime"`
}

func (*Post) TableName() string {
	return "posts"
}

// Again, I hate hooks but this is the best way.
// This populates the createdTime field: please don't use this field for database updates.
func (m *Post) AfterFind() (err error) {
	m.CreatedTime = m.CreatedAt
	return
}
