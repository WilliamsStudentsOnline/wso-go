package models

type Dorm struct {
	BaseSchema
	NeighborhoodID    uint         `json:"neighborhoodID"`
	Neighborhood      Neighborhood `json:"neighborhood,omitempty"`
	Name              string       `json:"name"`
	KeyOrCard         *string      `json:"keyOrCard"`
	Description       *string      `gorm:"size:65535" json:"description"`
	Built             *int         `json:"built"`
	Capacity          *int         `json:"capacity"`
	NumberBathrooms   *int         `json:"numberBathrooms"`
	NumberSingles     *int         `json:"numberSingles"`
	NumberDoubles     *int         `json:"numberDoubles"`
	NumberFlex        *int         `json:"numberFlex"`
	NumberWashers     *int         `json:"numberWashers"`
	BathroomRatio     *float64     `json:"bathroomRatio"`
	Comfort           *int         `json:"comfort"`
	Loudness          *int         `json:"loudness"`
	Convenience       *int         `json:"convenience"`
	Wifi              *float64     `json:"wifi"`
	Location          *float64     `json:"location"`
	Satisfaction      *float64     `json:"satisfaction"`
	AverageSingleArea *int         `json:"averageSingleArea"`
	AverageDoubleArea *int         `json:"averageDoubleArea"`
	ModeSingleArea    *int         `json:"modeSingleArea"`
	ModeDoubleArea    *int         `json:"modeDoubleArea"`
	DormRooms         []DormRoom   `json:"dormRooms,omitempty"`
}

func (*Dorm) TableName() string {
	return "dorms"
}
