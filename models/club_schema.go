package models

// Clubtrak Schema
type Club struct {
	BaseSchema
	Name               string `json:"name"`
	Subscribers        int    `json:"subscribers"`
	MeetingDescription string `json:"meetingDescription"`

	// Belongs to some club leader
	ClubAdmin   uint  `json:"clubAdmin"`
	ClubAdminID *User `json:"clubAdminID,omitempty"`
}

func (*Club) TableName() string {
	return "clubs"
}
