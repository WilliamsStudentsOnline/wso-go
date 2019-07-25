package models

// Tag Schema
type Tag struct {
	BaseSchema
	Name  string  `gorm:"unique;" json:"name"`
	Users []*User `gorm:"many2many:tags_users;" json:"users,omitempty"`
}

func (*Tag) TableName() string {
	return "tags"
}
