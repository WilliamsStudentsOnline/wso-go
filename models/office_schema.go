package models

type Office struct {
	BaseSchema
	Number string `gorm:"unique" json:"number"`
}

func (*Office) TableName() string {
	return "offices"
}
