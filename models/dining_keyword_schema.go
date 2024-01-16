package models

// Post Schema
type DiningKeyword struct {
	BaseSchema
	//set size limit to 30 to prevent any trolling
	Keyword string  `gorm:"size:30;not null" json:"keyword"`
	Users   []*User `gorm:"many2many:user_dining_keywords"`
}

func (*DiningKeyword) TableName() string {
	return "dining_keywords"
}
