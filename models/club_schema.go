package models

// Clubtrak Schema
type Club struct {
	BaseSchema
	Name               string `json:"name"`
	Category           string `json:"category"`
	Subscribers        int    `json:"subscribers"`
	MeetingDescription string `json:"meetingDescription"`
	ClubDescription    string `json:"clubDescription"`
	ClubPhoto          string `json:"clubPhoto"`

	// Belongs to some club leader
	ClubAdmin   uint  `json:"clubAdmin"`
	ClubAdminID *User `json:"clubAdminID,omitempty"`
}

func (*Club) TableName() string {
	return "clubs"
}
