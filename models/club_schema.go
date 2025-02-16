package models

// Clubtrak Schema
type Club struct {
	BaseSchema
	Name               string `json:"name"`
	Category           string `json:"category"`
	Subscribers        int    `json:"subscribers"`
	MeetingDescription string `json:"meetingDescription"`
	ClubDescription    string `json:"clubDescription"`
	ClubPhotoFilePath  string `json:"clubPhotoFilePath"`
	ContactEmail       string `json:"contactEmail"`
	ContactPhoneNumber string `json:"contactPhoneNumber"`
	Website            string `json:"website"`

	// Club leader's DB ID
	ClubAdminID uint `json:"clubAdminID"`
}

func (*Club) TableName() string {
	return "clubs"
}
