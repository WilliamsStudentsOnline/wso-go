package models

// ClubTrak Schema
type ClubTrak struct {
	BaseSchema

	NumMembers   string `gorm:"size:65535" json:"numMembers"`
	ClubLeaders  string `json:"clubLeaders"`
	Description  string `gorm:"not null;" json:"description"`
	MeetingTimes string `gorm:"not null;" json:"meetingTimes"`
	Events       string `gorm:"not null;" json:"events"`

	// Belongs to user Some Club Leader
	ClubID uint  `json:"clubID"`
	Club   *User `json:"club,omitempty"`
}

func (*ClubTrak) TableName() string {
	return "ClubTrak"
}
