package response

import "github.com/WilliamsStudentsOnline/wso-go/models"

type GetUserResponse struct {
	ID             uint    `json:"id"`
	Type           string  `json:"type"`
	Name           string  `json:"name"`
	CellPhone      *string `json:"cellPhone"`
	CampusPhoneExt *string `json:"campusPhoneEXT"`
	UnixID         string  `json:"unixID"`
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
	DepartmentID *uint                      `json:"departmentID"`
	Department   *GetUserResponseDepartment `json:"department,omitempty"`

	// Belongs to Office iff staff/professor
	OfficeID *uint                  `json:"officeID"`
	Office   *GetUserResponseOffice `json:"office,omitempty"`

	// Belongs to Dorm Room iff student
	DormRoomID *uint                    `gorm:"index:index_rooms_on_dorm_room_id" json:"dormRoomID"`
	DormRoom   *GetUserResponseDormRoom `json:"dormRoom,omitempty"`

	Pronoun              *string `json:"pronoun"`
	AtWilliams           *bool   `gorm:"DEFAULT:true;not null" json:"atWilliams"`
	OffCycle             *bool   `gorm:"DEFAULT:false;not null" json:"offCycle"`
	FactrakSurveyDeficit *int    `json:"factrakSurveyDeficit"`

	OptOutEphcatch      *bool `gorm:"DEFAULT:false;not null" json:"optOutEphcatch"`
	EphcatchEligibility *bool `gorm:"DEFAULT:false;not null" json:"ephcatchEligibility"`

	// Has many tags
	Tags []*GetUserResponseTag `gorm:"many2many:tags_users;" json:"tags,omitempty"`
}

type GetUserResponseDepartment struct {
	ID   uint   `json:"id"`
	Name string `gorm:"not null" json:"name"`
}

type GetUserResponseOffice struct {
	ID     uint   `json:"id"`
	Number string `gorm:"unique" json:"number"`
}

type GetUserResponseDormRoom struct {
	ID uint `json:"id"`

	// Belongs to dorm
	DormID uint                         `gorm:"index:index_dorm_rooms_on_dorm_id;not null" json:"dormID"`
	Dorm   *GetUserResponseDormRoomDorm `json:"dorm,omitempty"`

	Number string `gorm:"not null" json:"number"`

	RoomType string `gorm:"not null" json:"roomType"`
}

type GetUserResponseDormRoomDorm struct {
	ID uint `json:"id"`

	// Belongs to neighborhood
	NeighborhoodID uint `gorm:"not null" json:"neighborhoodID"`

	Name string `json:"name"`
}

type GetUserResponseTag struct {
	ID   uint   `json:"id"`
	Name string `gorm:"unique;" json:"name"`
}

func ConvertGetUserResponse(m *models.User) *GetUserResponse {
	if m == nil {
		return nil
	}

	r := &GetUserResponse{}
	r.ID = m.ID
	r.Type = m.Type
	r.Name = m.Name
	r.CellPhone = m.CellPhone
	r.CampusPhoneExt = m.CampusPhoneExt
	r.UnixID = m.UnixID
	r.WilliamsEmail = m.WilliamsEmail
	r.Title = m.Title
	r.Visible = m.Visible
	r.ClassYear = m.ClassYear
	r.DormVisible = m.DormVisible
	r.HomeTown = m.HomeTown
	r.HomeZip = m.HomeZip
	r.HomePhone = m.HomePhone
	r.HomeState = m.HomeState
	r.HomeCountry = m.HomeCountry
	r.HomeVisible = m.HomeVisible
	r.Major = m.Major
	r.SUBox = m.SUBox
	r.Entry = m.Entry
	r.Admin = m.Admin
	r.FactrakAdmin = m.FactrakAdmin
	r.HasAcceptedFactrakPolicy = m.HasAcceptedFactrakPolicy
	r.HasAcceptedDormtrakPolicy = m.HasAcceptedDormtrakPolicy
	r.DepartmentID = m.DepartmentID
	r.OfficeID = m.OfficeID
	r.DormRoomID = m.DormRoomID
	r.Pronoun = m.Pronoun
	r.AtWilliams = m.AtWilliams
	r.OffCycle = m.OffCycle
	r.FactrakSurveyDeficit = m.FactrakSurveyDeficit
	r.OptOutEphcatch = m.OptOutEphcatch
	r.EphcatchEligibility = m.EphcatchEligibility

	r.Department = ConvertGetUserResponseDepartment(m.Department)

	r.Office = ConvertGetUserResponseOffice(m.Office)

	r.DormRoom = ConvertGetUserResponseDormRoom(m.DormRoom)

	r.Tags = make([]*GetUserResponseTag, len(m.Tags))
	for i := range m.Tags {
		r.Tags[i] = ConvertGetUserResponseTag(m.Tags[i])
	}

	return r
}

func ConvertGetUserResponseDepartment(m *models.Department) *GetUserResponseDepartment {
	if m == nil {
		return nil
	}

	r := &GetUserResponseDepartment{}
	r.ID = m.ID
	r.Name = m.Name

	return r
}

func ConvertGetUserResponseOffice(m *models.Office) *GetUserResponseOffice {
	if m == nil {
		return nil
	}

	r := &GetUserResponseOffice{}
	r.ID = m.ID
	r.Number = m.Number

	return r
}

func ConvertGetUserResponseDormRoom(m *models.DormRoom) *GetUserResponseDormRoom {
	if m == nil {
		return nil
	}

	r := &GetUserResponseDormRoom{}
	r.ID = m.ID
	r.Number = m.Number
	r.RoomType = m.RoomType
	r.DormID = m.DormID

	r.Dorm = ConvertGetUserResponseDormRoomDorm(m.Dorm)

	return r
}

func ConvertGetUserResponseDormRoomDorm(m *models.Dorm) *GetUserResponseDormRoomDorm {
	if m == nil {
		return nil
	}

	r := &GetUserResponseDormRoomDorm{}
	r.ID = m.ID
	r.Name = m.Name
	r.NeighborhoodID = m.NeighborhoodID

	return r
}

func ConvertGetUserResponseTag(m *models.Tag) *GetUserResponseTag {
	if m == nil {
		return nil
	}

	r := &GetUserResponseTag{}
	r.ID = m.ID
	r.Name = m.Name

	return r
}
