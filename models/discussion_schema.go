package models

import "time"

// Discussion Schema
type Discussion struct {
	BaseSchema

	// Belongs to user
	UserID uint  `gorm:"index:index_discussions_on_user_id;not null" json:"userID"`
	User   *User `json:"user,omitempty"`

	LastActive time.Time `json:"lastActive"`
	Title      string    `json:"title"`

	// This is for when a user is deleted, but we still want to keep the info.
	// Ideally, we would only soft delete users but, we need to maintain the previous WSO's
	// standards.
	ExUserName string `json:"exUserName"`

	// Has many posts
	Posts []*Post `json:"posts,omitempty"`

	// Pass the created time: not looked at by GORM
	CreatedTime time.Time `gorm:"-" json:"createdTime"`
}

func (*Discussion) TableName() string {
	return "discussions"
}

func NewDiscussionByID(id uint) *Discussion {
	return &Discussion{
		BaseSchema: BaseSchema{
			ID: id,
		},
	}
}

func (d *Discussion) BeforeCreate() error {
	if d.LastActive.IsZero() {
		d.LastActive = time.Now()
	}

	return nil
}

// Again, I hate hooks but this is the best way.
// This populates the createdTime field: please don't use this field for database updates.
func (d *Discussion) AfterFind() (err error) {
	d.CreatedTime = d.CreatedAt
	return
}
