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
	CellPhone      *string `json:"cell_phone"`
	CampusPhoneExt *string `json:"campus_phone_ext"`
	UnixID         string  `json:"unix_id"`
	WilliamsEmail  string  `json:"williams_email"`
	Title          *string `json:"title"`
	Visible        bool    `json:"visible"`
	ClassYear      *int    `gorm:"size:4" json:"class_year"`

	// Equivalent to belongs_to Department
	DepartmentID *uint       `json:"department_id"`
	Department   *Department `json:"department,omitempty"`

	DormVisible bool    `gorm:"DEFAULT:true" json:"dorm_visible"`
	HomeTown    *string `json:"home_town"`
	HomeZip     *string `json:"home_zip"`
	HomePhone   *string `json:"home_phone"`
	HomeState   *string `json:"home_state"`
	HomeCountry *string `json:"home_country"`
	HomeVisible bool    `gorm:"DEFAULT:true" json:"home_visible"`

	Major                     *string `json:"major"`
	SUBox                     *string `json:"su_box"`
	Entry                     *string `json:"entry"`
	Admin                     bool    `gorm:"DEFAULT:false" json:"admin"`
	FactrakAdmin              bool    `gorm:"DEFAULT:false" json:"factrak_admin"`
	HasAcceptedFactrakPolicy  bool    `gorm:"DEFAULT:false" json:"has_accepted_factrak_policy"`
	HasAcceptedDormtrakPolicy bool    `gorm:"DEFAULT:false" json:"has_accepted_dormtrak_policy"`

	// belongs_to Office
	OfficeID *uint   `json:"office_id"`
	Office   *Office `json:"office,omitempty"`

	// belongs_to Dorm Room
	DormRoomID *uint     `gorm:"index_rooms_on_dorm_room_id" json:"dorm_room_id"`
	DormRoom   *DormRoom `json:"dorm_room,omitempty"`

	Pronoun              *string `json:"pronoun"`
	AtWilliams           bool    `gorm:"DEFAULT:true" json:"at_williams"`
	OffCycle             bool    `gorm:"DEFAULT:false" json:"off_cycle"`
	FactrakSurveyDeficit *int    `json:"factrak_survey_deficit"`

	OptOutEphcatch      bool `gorm:"DEFAULT:false" json:"opt_out_ephcatch"`
	EphcatchEligibility bool `gorm:"DEFAULT:false" json:"ephcatch_eligibility"`
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
		userModel := &UserModel{}
		userModel.DB = scope.DB()

		err = userModel.UpdateServerDeficit(u)
		if err != nil {
			return
		}
	}
	return
}
