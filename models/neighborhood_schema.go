package models

const (
	NeighborhoodFirstYear = "First-year"
	NeighborhoodCoop      = "Co-op"
)

type Neighborhood struct {
	BaseSchema

	Name    string `json:"name"`
	Trakked *bool  `gorm:"DEFAULT:true;not null" json:"trakked"`

	// Had many dorms
	Dorms []*Dorm `json:"dorms,omitempty"`
}

func (*Neighborhood) TableName() string {
	return "neighborhoods"
}
