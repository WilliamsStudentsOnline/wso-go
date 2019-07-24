package models

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/jinzhu/gorm"
)

// FactrakSurvey Schema
type FactrakSurvey struct {
	BaseSchema

	// Belongs to user (student)
	UserID uint  `gorm:"index:index_factrak_surveys_on_user_id;not null" json:"userID"`
	User   *User `json:"user,omitempty"`

	// Belongs to professor
	ProfessorID uint  `gorm:"index:index_factrak_surveys_on_professor_id;not null" json:"professorID"`
	Professor   *User `gorm:"foreignkey:ProfessorID" json:"professor,omitempty"`

	// Belongs to course
	CourseID uint    `gorm:"index:index_factrak_surveys_on_course_id;not null" json:"courseID"`
	Course   *Course `json:"course,omitempty"`

	WouldRecommendCourse *bool   `json:"wouldRecommendCourse"`
	CourseWorkload       *int    `json:"courseWorkload"`
	CourseStimulating    *int    `json:"courseStimulating"`
	WouldTakeAnother     *bool   `json:"wouldTakeAnother"`
	Approachability      *int    `json:"approachability"`
	LeadLecture          *int    `json:"leadLecture"`
	PromoteDiscussion    *int    `json:"promoteDiscussion"`
	OutsideHelpfulness   *int    `json:"outsideHelpfulness"`
	Comment              string  `gorm:"size:65535" json:"comment"`
	Flagged              bool    `json:"flagged"`
	GradeReceived        *string `json:"gradeReceived"`

	// Has many agreements
	Agreements []*FactrakAgreement `json:"agreements,omitempty"`

	// Not looked at by GORM, just for returning in JSON
	TotalAgree    int `gorm:"-" json:"totalAgree,omitempty"`
	TotalDisagree int `gorm:"-" json:"totalDisagree,omitempty"`
}

func (*FactrakSurvey) TableName() string {
	return "factrak_surveys"
}

// TODO: add not_same_prof_and_course to controller
// TODO: make comment be min 100 in controller
// TODO: check for valid course in model
func (s *FactrakSurvey) BeforeCreate() (err error) {
	if !s.User.IsStudent() {
		err = lib.ErrorUserMustBeStudent
	}
	if s.User.Student().Prefrosh() {
		err = lib.ErrorUserCannotBePrefrosh
	}
	return
}

func (s *FactrakSurvey) AfterDelete(tx *gorm.DB) (err error) {
	// Delete (permanent) factrack agreements
	err = tx.Unscoped().Where(FactrakAgreement{FactrakSurveyID: s.ID}).Delete(&FactrakAgreement{}).Error
	return
}
