package models

//CommunitySurveyResponse Schema
type CommunitySurveyResponse struct {
	BaseSchema

	// user info
	UserID uint  `gorm:"index:index_community_survey_response_on_user_id" json:"userID"`
	User   *User `json:"user,omitempty"`

	// survey info
	SurveyID uint             `gorm:"index;not null" json:"surveyID"`
	Survey   *CommunitySurvey `gorm:"foreignKey:SurveyID" json:"survey,omitempty"`

	// response info
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
