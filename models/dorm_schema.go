package models

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib"
)

type Dorm struct {
	BaseSchema

	// Belongs to neighborhood
	NeighborhoodID uint          `gorm:"not null" json:"neighborhoodID"`
	Neighborhood   *Neighborhood `json:"neighborhood,omitempty"`

	Name string `json:"name"`

	KeyOrCard         *string  `json:"keyOrCard"`
	Description       *string  `gorm:"size:65535" json:"description"`
	Built             *int     `json:"built"`
	Capacity          *int     `json:"capacity"`
	NumberBathrooms   *int     `json:"numberBathrooms"`
	NumberSingles     *int     `json:"numberSingles"`
	NumberDoubles     *int     `json:"numberDoubles"`
	NumberFlex        *int     `json:"numberFlex"`
	NumberWashers     *int     `json:"numberWashers"`
	BathroomRatio     *float64 `json:"bathroomRatio"`
	Comfort           *int     `json:"comfort"`
	Loudness          *int     `json:"loudness"`
	Convenience       *int     `json:"convenience"`
	Wifi              *float64 `json:"wifi"`
	Location          *float64 `json:"location"`
	Satisfaction      *float64 `json:"satisfaction"`
	AverageSingleArea *int     `json:"averageSingleArea"`
	AverageDoubleArea *int     `json:"averageDoubleArea"`
	ModeSingleArea    *int     `json:"modeSingleArea"`
	ModeDoubleArea    *int     `json:"modeDoubleArea"`

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

	r.Capacity = lib.IntToPtr(ns + 2*nd + 2*nf)

	// Set bathroom ratio
	if r.NumberBathrooms != nil && *r.NumberBathrooms != 0 && r.Capacity != nil {
		ratio := float64(*r.Capacity) / float64(*r.NumberBathrooms)
		r.BathroomRatio = &ratio
	}
	return
}
