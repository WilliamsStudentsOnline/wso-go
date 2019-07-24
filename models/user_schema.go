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
	Visible        bool    `json:"visible"`
	ClassYear      *int    `gorm:"size:4" json:"classYear"`

	// Equivalent to belongs_to Department
	DepartmentID *uint       `json:"departmentID"`
	Department   *Department `json:"department,omitempty"`

	DormVisible bool    `gorm:"DEFAULT:true" json:"dormVisible"`
	HomeTown    *string `json:"homeTown"`
	HomeZip     *string `json:"homeZip"`
	HomePhone   *string `json:"homePhone"`
	HomeState   *string `json:"homeState"`
	HomeCountry *string `json:"homeCountry"`
	HomeVisible bool    `gorm:"DEFAULT:true" json:"homeVisible"`

	Major                     *string `json:"major"`
	SUBox                     *string `json:"suBox"`
	Entry                     *string `json:"entry"`
	Admin                     bool    `gorm:"DEFAULT:false" json:"admin"`
	FactrakAdmin              bool    `gorm:"DEFAULT:false" json:"factrakAdmin"`
	HasAcceptedFactrakPolicy  bool    `gorm:"DEFAULT:false" json:"hasAcceptedFactrakPolicy"`
	HasAcceptedDormtrakPolicy bool    `gorm:"DEFAULT:false" json:"hasAcceptedDormtrakPolicy"`

	// belongs_to Office
	OfficeID *uint   `json:"officeID"`
	Office   *Office `json:"office,omitempty"`

	// belongs_to Dorm Room
	DormRoomID *uint     `gorm:"index:index_rooms_on_dorm_room_id" json:"dormRoomID"`
	DormRoom   *DormRoom `json:"dormRoom,omitempty"`

	Pronoun              *string `json:"pronoun"`
	AtWilliams           bool    `gorm:"DEFAULT:true" json:"atWilliams"`
	OffCycle             bool    `gorm:"DEFAULT:false" json:"offCycle"`
	FactrakSurveyDeficit *int    `json:"factrakSurveyDeficit"`

	OptOutEphcatch      bool `gorm:"DEFAULT:false" json:"optOutEphcatch"`
	EphcatchEligibility bool `gorm:"DEFAULT:false" json:"ephcatchEligibility"`

	// Has many tags
	Tags []*Tag `gorm:"many2many:tags_users;" json:"tags,omitempty"`

	// Has many factrak surveys (if student)
	StudentFactrakSurveys []*FactrakSurvey `gorm:"foreignkey:UserID" json:"factrakSurveys,omitempty"`

	// Has many factrak surveys (if professor)
	ProfessorFactrakSurveys []*FactrakSurvey `gorm:"foreignkey:ProfessorID" json:"factrakSurveys,omitempty"`

	// Has many factrak agreements
	FactrakAgreements []*FactrakAgreement `json:"factrakAgreements,omitempty"`
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

func (u *User) AfterCreate(scope *gorm.Scope) (err error) {
	if u.IsStudent() {
		// Update survey deficit
		// This way, only student get a default, non-nil value for this
		userModel := NewUserModel(scope.DB())

		err = userModel.UpdateServerDeficit(u)
		if err != nil {
			return
		}
	}
	return
}
