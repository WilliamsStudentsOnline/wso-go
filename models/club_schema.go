package models

// ClubTrak Schema
type ClubTrak struct {
	BaseSchema

	NumMembers   string `gorm:"size:65535" json:"numMembers"`
	ClubLeaders  string `json:"clubLeaders"`
	Description  string `json:"description"`
	MeetingTimes string `json:"meetingTimes"`
	Events       string `json:"events"`

	// Belongs to user Some Club Leader
	ClubID uint   `json:"clubID"`
	Club   *User  `json:"club,omitempty"`
	Name   string `json:"name"`
}

func (*ClubTrak) TableName() string {
	return "clubs"
}
