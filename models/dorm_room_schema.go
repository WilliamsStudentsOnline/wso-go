package models

const (
	DormRoomTypeSingle = "s"
	DormRoomTypeDouble = "d"
	DormRoomTypeFlex   = "f"
)

type DormRoom struct {
	BaseSchema
	DormID           uint     `gorm:"index:index_dorm_rooms_on_dorm_id" json:"dormID"`
	Dorm             Dorm     `json:"dorm,omitempty"`
	Number           string   `json:"number"`
	Closet           *string  `json:"closet"`
	Flooring         *string  `json:"flooring"`
	CommonRoomAccess *bool    `json:"commonRoomAccess"`
	CommonRoomDesc   *string  `gorm:"size:65535" json:"commonRoomDesc"`
	ThermostatAccess *bool    `json:"thermostat_access"`
	ThermostatDesc   *string  `gorm:"size:65535" json:"thermostatDesc"`
	OutletsDesc      *string  `gorm:"size:65535" json:"outletsDesc"`
	KeyOrCard        *string  `json:"keyOrCard"`
	Faces            *string  `json:"faces"`
	Noise            *string  `gorm:"size:65535" json:"noise"`
	BedAdjustable    *bool    `json:"bed_adjustable"`
	HC               *bool    `json:"hc"`
	PrivateBathroom  *bool    `json:"privateBathroom"`
	FloorNumber      *int     `json:"floorNumber"`
	Area             *int     `json:"area"`
	Walkthrough      *bool    `json:"walkthrough"`
	BathroomDesc     *string  `gorm:"size:65535" json:"bathroomDesc"`
	NumFlag          *bool    `json:"numFlag"`
	Picture          *string  `json:"picture"`
	Comfort          *float64 `json:"comfort"`
	Loudness         *float64 `json:"loudness"`
	Convenience      *float64 `json:"convenience"`
	Wifi             *float64 `json:"wifi"`
	Location         *float64 `json:"location"`
	Satisfaction     *float64 `json:"satisfaction"`
	RoomType         *string  `json:"roomType"`

	// Has many students
	Users []User `json:"users,omitempty"`
}

func (*DormRoom) TableName() string {
	return "dorm_rooms"
}
