package models

import "time"

type BoardPoll struct {
	BaseSchema

	ThreadID uint       `gorm:"unique_index:index_board_polls_on_thread_id;not null" json:"threadID"`
	ClosesAt *time.Time `json:"closesAt"`

	Options []*BoardPollOption `gorm:"foreignkey:PollID;association_foreignkey:ID" json:"options,omitempty"`

	Closed     bool  `gorm:"-" json:"closed"`
	TotalVotes *int  `gorm:"-" json:"totalVotes,omitempty"`
	MyVote     *uint `gorm:"-" json:"myVote,omitempty"`
}

func (*BoardPoll) TableName() string {
	return "board_polls"
}

func (p *BoardPoll) AfterFind() (err error) {
	p.Closed = p.ClosesAt != nil && !p.ClosesAt.After(time.Now())
	return
}

type BoardPollOption struct {
	BaseSchema

	PollID   uint   `gorm:"index:index_board_poll_options_on_poll_id;not null" json:"-"`
	Position int    `gorm:"not null" json:"-"`
	Label    string `gorm:"not null" json:"label"`

	Votes *int `gorm:"-" json:"votes,omitempty"`
}

func (*BoardPollOption) TableName() string {
	return "board_poll_options"
}

type BoardPollVote struct {
	BaseSchema

	PollID   uint `gorm:"unique_index:index_board_poll_votes_on_poll_id_and_user_id;not null"`
	UserID   uint `gorm:"unique_index:index_board_poll_votes_on_poll_id_and_user_id;not null"`
	OptionID uint `gorm:"index:index_board_poll_votes_on_option_id;not null"`
}

func (*BoardPollVote) TableName() string {
	return "board_poll_votes"
}
