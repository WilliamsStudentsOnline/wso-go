package models

import (
	"strings"
)

// CommunitySurvey Schema
type CommunitySurvey struct {
	BaseSchema
	// survey info
	Question string `gorm:"size:65535" json:"question"`

	UserID uint  `gorm:"index:index_user_id;not null" json:"userID"`
	User   *User `gorm:"foreignKey:UserID" json:"user"`

	// options info
	NumOptions uint   `json:"numOptions"`
	Options    string `gorm:"size:65535" json:"Options"` // serialized,eg: option1;option2;option3

	// number of respondents
	AnswerCount uint `gorm:"DEFAULT:0;not null" json:"answerCount"`
}

func (s *CommunitySurvey) SerializeOptions(survey *CommunitySurvey, options []string) error {
	survey.Options = strings.Join(options, ";")
	survey.NumOptions = uint(len(options))
	return nil
}

func (s *CommunitySurvey) DeserializeOptions(survey *CommunitySurvey) ([]string, error) {
	options := strings.Split(survey.Options, ";")
	return options, nil
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
