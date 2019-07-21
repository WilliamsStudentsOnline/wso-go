package models

const (
	DormRoomTypeSingle = "s"
	DormRoomTypeDouble = "d"
	DormRoomTypeFlex = "f"
)

type DormRoom struct {
	BaseSchema
	DormID           int      `gorm:"index:index_dorm_rooms_on_dorm_id" json:"dorm_id"`
	Dorm             Dorm     `json:"-"`
	Number           string   `json:"number"`
	Closet           *string  `json:"closet"`
	Flooring         *string  `json:"flooring"`
	CommonRoomAccess *bool    `json:"common_room_access"`
	CommonRoomDesc   *string  `gorm:"size:65535" json:"common_room_desc"`
	ThermostatAccess *bool    `json:"thermostat_access"`
	ThermostatDesc   *string  `gorm:"size:65535" json:"thermostat_desc"`
	OutletsDesc      *string  `gorm:"size:65535" json:"outlets_desc"`
	KeyOrCard        *string  `json:"key_or_card"`
	Faces            *string  `json:"faces"`
	Noise            *string  `gorm:"size:65535" json:"noise"`
	BedAdjustable    *bool    `json:"bed_adjustable"`
	HC               *bool    `json:"hc"`
	PrivateBathroom  *bool    `json:"private_bathroom"`
	FloorNumber      *int     `json:"floor_number"`
	Area             *int     `json:"area"`
	Walkthrough      *bool    `json:"walkthrough"`
	BathroomDesc     *string  `gorm:"size:65535" json:"bathroom_desc"`
	NumFlag          *bool    `json:"num_flag"`
	Picture          *string  `json:"picture"`
	Comfort          *float64 `json:"comfort"`
	Loudness         *float64 `json:"loudness"`
	Convenience      *float64 `json:"convenience"`
	Wifi             *float64 `json:"wifi"`
	Location         *float64 `json:"location"`
	Satisfaction     *float64 `json:"satisfaction"`
	RoomType         *string  `json:"room_type"`
}

func (*DormRoom) TableName() string {
	return "dorm_rooms"
}
