package models

type CommunitySurveyResponse struct {
	BaseSchema

	UserID uint  `gorm:"index:index_user_id;not null" json:"userID"`
	User   *User `json:"user"`

	SurveyID uint             `gorm:"index:index_survey_id;not null" json:"surveyID"`
	Survey   *CommunitySurvey `gorm:"foreignKey:SurveyID" json:"survey"`

	Response string `gorm:"size:65535" json:"response"`
}

func (*CommunitySurveyResponse) TableName() string {
	return "community_survey_responses"
}

func NewCommunitySurveyResponse(id uint) *CommunitySurveyResponse {
	return &CommunitySurveyResponse{
		BaseSchema: BaseSchema{
			ID: id,
		},
	}
}
