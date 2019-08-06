package models

import "github.com/jinzhu/gorm"

const (
	UserTypeStudent   = "student"
	UserTypeAlum      = "alum"
	UserTypeProfessor = "professor"
	UserTypeStaff     = "staff"
)

// User Model Schema
type User struct {
	BaseSchema
	Type           string  `json:"type"`
	Name           string  `json:"name"`
	CellPhone      *string `json:"cellPhone"`
	CampusPhoneExt *string `json:"campusPhoneEXT"`
	UnixID         string  `gorm:"unique;not null;" json:"unixID"`
	WilliamsEmail  string  `json:"williamsEmail"`
	Title          *string `json:"title"`
	Visible        *bool   `gorm:"DEFAULT:true;not null" json:"visible"`
	ClassYear      *int    `gorm:"size:4" json:"classYear"`

	DormVisible *bool   `gorm:"DEFAULT:true;not null" json:"dormVisible"`
	HomeTown    *string `json:"homeTown"`
	HomeZip     *string `json:"homeZip"`
	HomePhone   *string `json:"homePhone"`
	HomeState   *string `json:"homeState"`
	HomeCountry *string `json:"homeCountry"`
	HomeVisible *bool   `gorm:"DEFAULT:true;not null" json:"homeVisible"`

	Major                     *string `json:"major"`
	SUBox                     *string `json:"suBox"`
	Entry                     *string `json:"entry"`
	Admin                     *bool   `gorm:"DEFAULT:false;not null" json:"admin"`
	FactrakAdmin              *bool   `gorm:"DEFAULT:false;not null" json:"factrakAdmin"`
	HasAcceptedFactrakPolicy  *bool   `gorm:"DEFAULT:false;not null" json:"hasAcceptedFactrakPolicy"`
	HasAcceptedDormtrakPolicy *bool   `gorm:"DEFAULT:false;not null" json:"hasAcceptedDormtrakPolicy"`

	// Belongs to Department iff professor
	DepartmentID *uint       `json:"departmentID"`
	Department   *Department `json:"department,omitempty"`

	// Belongs to Office iff staff/professor
	OfficeID *uint   `json:"officeID"`
	Office   *Office `json:"office,omitempty"`

	// Belongs to Dorm Room iff student
	DormRoomID *uint     `gorm:"index:index_rooms_on_dorm_room_id" json:"dormRoomID"`
	DormRoom   *DormRoom `json:"dormRoom,omitempty"`

	Pronoun              *string `json:"pronoun"`
	AtWilliams           *bool   `gorm:"DEFAULT:true;not null" json:"atWilliams"`
	OffCycle             *bool   `gorm:"DEFAULT:false;not null" json:"offCycle"`
	FactrakSurveyDeficit *int    `json:"factrakSurveyDeficit"`

	OptOutEphcatch      *bool `gorm:"DEFAULT:false;not null" json:"optOutEphcatch"`
	EphcatchEligibility *bool `gorm:"DEFAULT:false;not null" json:"ephcatchEligibility"`

	// Has many tags
	Tags []*Tag `gorm:"many2many:tags_users;" json:"tags,omitempty"`

	// These factrak survey fields are for GORM only: JSON will use the factrakSurveys field below.
	// Has many factrak surveys (if student)
	StudentFactrakSurveys []*FactrakSurvey `gorm:"foreignkey:UserID" json:"-"`
	// Has many factrak surveys (if professor)
	ProfessorFactrakSurveys []*FactrakSurvey `gorm:"foreignkey:ProfessorID" json:"-"`
	// As we cannot be both a student and a professor, this combines either a student or a professor's factrak survey.
	// We populate this field as a hook AfterFind.
	FactrakSurveys []*FactrakSurvey `gorm:"-" json:"factrakSurveys"`

	// Has many factrak agreements
	FactrakAgreements []*FactrakAgreement `json:"factrakAgreements,omitempty"`

	// Has many dormtrak reviews
	DormtrakReviews []*DormtrakReview `json:"dormtrakReviews,omitempty"`
}

func (*User) TableName() string {
	return "users"
}

func NewUserWithID(userID uint) User {
	return User{
		BaseSchema: BaseSchema{
			ID: userID,
		},
	}
}

func (u *User) IsStudent() bool {
	return u.Type == UserTypeStudent
}

func (u *User) IsAlum() bool {
	return u.Type == UserTypeAlum
}

func (u *User) IsProfessor() bool {
	return u.Type == UserTypeProfessor
}

func (u *User) IsStaff() bool {
	return u.Type == UserTypeStaff
}

func (u *User) Student() *Student {
	return &Student{
		User: u,
	}
}

// I hate hooks but I'm keeping this one here, as it is useful. Otherwise, put hooks in controllers.
func (u *User) AfterCreate(scope *gorm.Scope) (err error) {
	if u.IsStudent() {
		// Update survey deficit
		// This way, only student get a default, non-nil value for this
		userModel := NewUserModel(scope.DB())

		err = userModel.UpdateFactrakSurveyDeficit(u)
		if err != nil {
			return
		}
	}
	return
}

// Again, I hate hooks but this is the best way.
// This populates the factrak surveys field: please don't use this field for database updates.
// Factrak surveys are StudentFactrakSurveys if type=student, ProfessorFactrakSurveys if type=professor,
// and ProfessorFactrakSurveys if both StudentFactrakSurveys and ProfessorFactrakSurveys exist.
func (u *User) AfterFind() (err error) {
	if u.StudentFactrakSurveys != nil {
		u.FactrakSurveys = u.StudentFactrakSurveys
	}
	if u.ProfessorFactrakSurveys != nil {
		u.FactrakSurveys = u.ProfessorFactrakSurveys
	}
	return
}
