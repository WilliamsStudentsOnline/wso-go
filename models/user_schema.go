package models

import (
	"strconv"
	"strings"

	"github.com/jinzhu/gorm"
)

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
	Nickname       *string `json:"nickname"`

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

	// Keep this in here as long as we want to maintain this type of searching.
	SearchFields string `gorm:"default:'';not null'" json:"-"`

	// Has many tags
	Tags []*Tag `gorm:"many2many:tags_users;" json:"tags,omitempty"`

	// These factrak survey fields are for GORM only: JSON will use the factrakSurveys field below.
	// Has many factrak surveys (if student)
	StudentFactrakSurveys []*FactrakSurvey `gorm:"foreignkey:UserID" json:"-"`
	// Has many factrak surveys (if professor)
	ProfessorFactrakSurveys []*FactrakSurvey `gorm:"foreignkey:ProfessorID" json:"-"`
	// As we cannot be both a student and a professor, this combines either a student or a professor's factrak survey.
	// We populate this field as a hook AfterFind.
	FactrakSurveys []*FactrakSurvey `gorm:"-" json:"factrakSurveys,omitempty"`

	// Has many factrak agreements
	FactrakAgreements []*FactrakAgreement `json:"factrakAgreements,omitempty"`

	// Has many dormtrak reviews
	DormtrakReviews []*DormtrakReview `json:"dormtrakReviews,omitempty"`

	// Has many ephcatches (owner side)
	Ephcatches []*Ephcatch `gorm:"foreignkey:UserID" json:"ephcatches,omitempty"`
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

func (u *User) HomeAddress() string {
	var addressSlice []string

	if u.HomeTown != nil {
		addressSlice = append(addressSlice, *u.HomeTown)
	}
	if u.HomeState != nil {
		addressSlice = append(addressSlice, *u.HomeState)
	}
	if u.HomeCountry != nil && *u.HomeCountry != "United States" {
		addressSlice = append(addressSlice, *u.HomeCountry)
	}

	return strings.Join(addressSlice, ", ")
}

func (u *User) GenerateSearchFields() string {
	var searchFields []string

	if *u.Visible && *u.AtWilliams {
		searchFields = append(searchFields, u.Name, u.UnixID)
		if u.Title != nil {
			searchFields = append(searchFields, *u.Title)
		}
		if u.ClassYear != nil {
			searchFields = append(searchFields, strconv.Itoa(*u.ClassYear))
		}
		if u.Major != nil {
			searchFields = append(searchFields, *u.Major)
		}
		if u.SUBox != nil {
			searchFields = append(searchFields, *u.SUBox)
		}
		if u.Entry != nil {
			searchFields = append(searchFields, *u.Entry)
		}
		if *u.DormVisible && u.DormRoom != nil {
			searchFields = append(searchFields, u.DormRoom.Dorm.Name+" "+u.DormRoom.Number)
		} else if u.Office != nil {
			searchFields = append(searchFields, u.Office.Number)
		}
		if *u.HomeVisible && u.HomeTown != nil {
			searchFields = append(searchFields, u.HomeAddress())
		}

		for _, tag := range u.Tags {
			searchFields = append(searchFields, tag.Name)
		}
	}

	searchFieldsStr := strings.Join(searchFields, "#")
	searchFieldsStr = strings.ToLower(searchFieldsStr)
	return searchFieldsStr
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

// This updates the user search fields every time the user is updated.
// Again, I really hate hooks, and this one seems useless to me, but for the MVP we shall keep it.
// I think this is not a function we should always try to update, as searching can be over slightly old records.
// Ideally, this would run daily to update search fields. Actually, ideally, we would not use MySQL for search.
// TODO(aidan): Make this hook optional in config and/or remove it in production.
// NOTE: This might fail horribly if we try to update a user with only SOME data (not every field loaded).
// XXX: Removing this due to the fact it could catastrophically fail. Now we just generate search fields in specific
//  functions (and before create)
/*func (u *User) BeforeSave(scope *gorm.Scope) (err error) {
	userModel := NewUserModel(scope.DB())
	searchFields, err := userModel.generateSearchFieldsByUser(*u)
	if err != nil {
		return
	}
	u.SearchFields = searchFields
	return
}*/

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
