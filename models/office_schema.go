package models

type Office struct {
	BaseSchema
	Number string `gorm:"unique" json:"number"`
	Users  []User `json:"users,omitempty"`
}

func (*Office) TableName() string {
	return "offices"
}
