package models

import (
	"time"
)

// CommunitySurvey Schema
type CommunitySurvey struct {
	BaseSchema
	// survey info
	Question  *string `gorm:"size:65535" json:"question"`
	Affiliate string  `json:"affiliate"`

	//Has many options
	Options []*Option `json:"options"`

	// number of respondents
	AnswerCount uint `gorm:"DEFAULT:0;not null" json:"answerCount"`

	//pass the created time: not looked at by gorm
	CreatedTime time.Time `gorm:"-" json:"createdTime"`
}

type Option struct {
	BaseSchema
	OptionText string `gorm:"size:65535" json:"optionText"`
	OptionType string `json:"optionType"` //open-ended, multiple choices, etc.
}

func (*CommunitySurvey) TableName() string {
	return "community_surveys"
}

func NewCommunitySurvey(id uint) *CommunitySurvey {
	return &CommunitySurvey{
		BaseSchema: BaseSchema{
			ID: id,
		},
	}
}

func (m *CommunitySurvey) AfterFind() (err error) {
	m.CreatedTime = m.CreatedAt
	return
}
