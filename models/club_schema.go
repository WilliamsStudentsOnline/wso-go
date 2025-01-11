package models

// ClubTrak Schema
type Club struct {
	BaseSchema
	Name               string `json:"name"`
	Subscribers        int    `json:"Subscribers"`
	MeetingDescription string `json:"meetingDescription"`
	Events             string `json:"events"`

	// Belongs to user Some Club Leader
	ClubAdmin   uint  `json:"clubID"`
	ClubAdminID *User `json:"club,omitempty"`
}

func (*Club) TableName() string {
	return "clubs"
}
