package models

// import "strings"

// CommunitySurvey Schema
type CommunitySurvey struct {
	BaseSchema
	// survey info
	Question  string `gorm:"size:65535" json:"question"`
	Affiliate string `json:"affiliate"`

	// options info
	NumOptions        uint   `json:"numOptions"`
	SerializedOptions string `gorm:"size:65535" json:"serializedOptions"` // eg: option1;option2;option3

	// number of respondents
	AnswerCount uint `gorm:"DEFAULT:0;not null" json:"answerCount"`
}

/* I wrote these two extra functions and wasn't sure if I need them or not so I left them as a comment

func addOptions (survey *CommunitySurvey, options [] string){
	survey.SerializedOptions = strings.Join(options, ";")
	survey.NumOptions = uint(len(options))
}

func getOptions(survey *CommunitySurvey) []string{
	return strings.Split(survey.SerializedOptions, ";")
}
*/

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
