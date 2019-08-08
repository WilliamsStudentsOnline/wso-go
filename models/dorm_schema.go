package models

type Dorm struct {
	BaseSchema

	// Belongs to neighborhood
	NeighborhoodID uint          `gorm:"not null" json:"neighborhoodID"`
	Neighborhood   *Neighborhood `json:"neighborhood,omitempty"`

	Name string `json:"name"`

	KeyOrCard         *string  `json:"keyOrCard"`
	Description       *string  `gorm:"size:65535" json:"description"`
	Built             *int     `json:"built"`
	Capacity          *int     `gorm:"DEFAULT:0;not null" json:"capacity"`
	NumberBathrooms   *int     `gorm:"DEFAULT:0;not null" json:"numberBathrooms"`
	NumberSingles     *int     `gorm:"DEFAULT:0;not null" json:"numberSingles"`
	NumberDoubles     *int     `gorm:"DEFAULT:0;not null" json:"numberDoubles"`
	NumberFlex        *int     `gorm:"DEFAULT:0;not null" json:"numberFlex"`
	NumberWashers     *int     `gorm:"DEFAULT:0;not null" json:"numberWashers"`
	BathroomRatio     *float64 `json:"bathroomRatio"`
	AverageSingleArea *int     `json:"averageSingleArea"`
	AverageDoubleArea *int     `json:"averageDoubleArea"`
	ModeSingleArea    *int     `json:"modeSingleArea"`
	ModeDoubleArea    *int     `json:"modeDoubleArea"`

	// These are average statistics from dorm reviews
	Comfort      *float64 `json:"comfort"`
	Loudness     *float64 `json:"loudness"`
	Convenience  *float64 `json:"convenience"`
	Wifi         *float64 `json:"wifi"`
	Location     *float64 `json:"location"`
	Satisfaction *float64 `json:"satisfaction"`

	// Has many dorm rooms
	DormRooms []*DormRoom `json:"dormRooms,omitempty"`
}

func (*Dorm) TableName() string {
	return "dorms"
}

// This updates the dorm before we save it. It sets capacity and bathroom size
func (r *Dorm) BeforeSave() (err error) {
	// Set capacity
	ns := 0
	if r.NumberSingles != nil {
		ns = *r.NumberSingles
	}
	nd := 0
	if r.NumberDoubles != nil {
		nd = *r.NumberDoubles
	}
	nf := 0
	if r.NumberFlex != nil {
		nf = *r.NumberFlex
	}

	capacity := ns + 2*nd + 2*nf
	r.Capacity = &capacity

	// Set bathroom ratio
	if r.NumberBathrooms != nil && *r.NumberBathrooms != 0 && r.Capacity != nil {
		ratio := float64(*r.Capacity) / float64(*r.NumberBathrooms)
		r.BathroomRatio = &ratio
	}
	return
}
