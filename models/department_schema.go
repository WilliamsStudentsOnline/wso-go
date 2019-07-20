package models

type Department struct {
	BaseSchema
	Name string `json:"name"`
}

func (*Department) TableName() string {
	return "departments"
}
