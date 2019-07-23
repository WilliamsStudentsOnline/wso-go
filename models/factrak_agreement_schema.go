package models

import "github.com/WilliamsStudentsOnline/wso-go/lib"

// FactrakAgreement Schema
type FactrakAgreement struct {
	BaseSchema
	Agrees bool `json:"agrees"`

	// Belongs to survey
	FactrakSurveyID uint `gorm:"index:index_factrak_agreements_on_factrak_survey_id;" json:"factrakSurveyID"`
	FactrakSurvey *FactrakSurvey `json:"factrakSurvey"`

	// Belongs to user
	UserID uint `gorm:"index:index_factrak_agreements_on_user_id;" json:"userID"`
	User *User `json:"user"`
}

func (*FactrakAgreement) TableName() string {
	return "factrak_agreements"
}

func (s *FactrakAgreement) BeforeCreate() (err error) {
	if !s.User.IsStudent() {
		err = lib.ErrorUserMustBeStudent
	}

	return
}

