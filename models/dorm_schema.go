package models

type Dorm struct {
	BaseSchema
	NeighborhoodID    int          `json:"neighborhood_id"`
	Neighborhood      Neighborhood `json:"neighborhood,omitempty"`
	Name              string       `json:"name"`
	KeyOrCard         *string      `json:"key_or_card"`
	Description       *string      `gorm:"size:65535" json:"description"`
	Built             *int         `json:"built"`
	Capacity          *int         `json:"capacity"`
	NumberBathrooms   *int         `json:"number_bathrooms"`
	NumberSingles     *int         `json:"number_singles"`
	NumberDoubles     *int         `json:"number_doubles"`
	NumberFlex        *int         `json:"number_flex"`
	NumberWashers     *int         `json:"number_washers"`
	BathroomRatio     *float64     `json:"bathroom_ratio"`
	Comfort           *int         `json:"comfort"`
	Loudness          *int         `json:"loudness"`
	Convenience       *int         `json:"convenience"`
	Wifi              *float64     `json:"wifi"`
	Location          *float64     `json:"location"`
	Satisfaction      *float64     `json:"satisfaction"`
	AverageSingleArea *int         `json:"average_single_area"`
	AverageDoubleArea *int         `json:"average_double_area"`
	ModeSingleArea    *int         `json:"mode_single_area"`
	ModeDoubleArea    *int         `json:"mode_double_area"`
	DormRooms         []DormRoom   `json:"dorm_rooms,omitempty"`
}

func (*Dorm) TableName() string {
	return "dorms"
}
