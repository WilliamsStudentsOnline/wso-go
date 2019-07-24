package models

type Department struct {
	BaseSchema
	Name string `gorm:"not null" json:"name"`

	// Has many professors & staff
	Users []*User `json:"users,omitempty"`

	// Has many areas of study
	AreasOfStudy []*AreaOfStudy `json:"areasOfStudy,omitempty"`
}

func (*Department) TableName() string {
	return "departments"
}
