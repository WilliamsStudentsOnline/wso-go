package models

import "time"

type DailyMessage struct {
	BaseSchema
	MessageID        uint      `gorm:"unique" json:"id"`
	Date             time.Time `json:"date"`
	Title            string    `json:"title"`
	ShortDescription string    `json:"short_description"`
	Description      string    `json:"description"`
	Author           string    `json:"author"`
	AuthorEmail      string    `json:"author_email"`
	Department       string    `json:"dept"`
	Venue            *string   `json:"venue"`
	Category         string    `gorm:"index_category" json:"category"`
}

func (*DailyMessage) TableName() string {
	return "daily_messages"
}
