package models

import "time"

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

	// Pass the created time: not looked at by GORM
	CreatedTime time.Time `gorm:"-" json:"createdTime"`
}

func (*FactrakSurvey) TableName() string {
	return "factrak_surveys"
}

func NewFactrakSurvey(id uint) *FactrakSurvey {
	return &FactrakSurvey{
		BaseSchema: BaseSchema{
			ID: id,
		},
	}
}

// Again, I hate hooks but this is the best way.
// This populates the createdTime field: please don't use this field for database updates.
func (m *FactrakSurvey) AfterFind() (err error) {
	m.CreatedTime = m.CreatedAt
	return
}
