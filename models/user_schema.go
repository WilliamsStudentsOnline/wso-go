package models

//User ...
type User struct {
	BaseSchema
	Type           string `json:"type"`
	Name           string `json:"name"`
	CellPhone      string `json:"cell_phone"`
	CampusPhoneExt string `json:"campus_phone_ext"`
	UnixID         string `json:"unix_id"`
	WilliamsEmail  string `json:"williams_email"`
	Title          string `json:"title"`
	Visible        bool   `json:"visible"`
	ClassYear      int    `gorm:"size:4" json:"class_year"`

	// Equivalent to belongs_to Department
	DepartmentID int        `json:"department_id"`
	Department   *Department `json:"department,omitempty"`

	DormVisible bool   `gorm:"DEFAULT:true" json:"dorm_visible"`
	HomeTown    string `json:"home_town"`
	HomeZip     string `json:"home_zip"`
	HomePhone   string `json:"home_phone"`
	HomeState   string `json:"home_state"`
	HomeCountry string `json:"home_country"`
	HomeVisible bool   `gorm:"DEFAULT:true" json:"home_visible"`

	Major                     string `json:"major"`
	SUBox                     string `json:"su_box"`
	Entry                     string `json:"entry"`
	Admin                     bool   `gorm:"DEFAULT:false" json:"admin"`
	FactrakAdmin              bool   `gorm:"DEFAULT:false" json:"factrak_admin"`
	HasAcceptedFactrakPolicy  bool   `gorm:"DEFAULT:false" json:"has_accepted_factrak_policy"`
	HasAcceptedDormtrakPolicy bool   `gorm:"DEFAULT:false" json:"has_accepted_dormtrak_policy"`

	// belongs_to Office

	// belongs_to Dorm Room

	Pronoun              string `json:"pronoun"`
	AtWilliams           bool   `gorm:"DEFAULT:true" json:"at_williams"`
	OffCycle             bool   `gorm:"DEFAULT:false" json:"off_cycle"`
	FactrakSurveyDeficit int    `json:"factrak_survey_deficit"`

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
