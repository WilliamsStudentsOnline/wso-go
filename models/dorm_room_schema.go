package models

import "github.com/jinzhu/gorm"

const (
	DormRoomTypeSingle = "s"
	DormRoomTypeDouble = "d"
	DormRoomTypeFlex   = "f"
)

type DormRoom struct {
	BaseSchema

	// Belongs to dorm
	DormID uint  `gorm:"index:index_dorm_rooms_on_dorm_id;not null" json:"dormID"`
	Dorm   *Dorm `json:"dorm,omitempty"`

	Number string `gorm:"not null" json:"number"`

	Closet           *string `json:"closet"`
	Flooring         *string `json:"flooring"`
	CommonRoomAccess *bool   `json:"commonRoomAccess"`
	CommonRoomDesc   *string `gorm:"size:65535" json:"commonRoomDesc"`
	ThermostatAccess *bool   `json:"thermostat_access"`
	ThermostatDesc   *string `gorm:"size:65535" json:"thermostatDesc"`
	OutletsDesc      *string `gorm:"size:65535" json:"outletsDesc"`
	KeyOrCard        *string `json:"keyOrCard"`
	Noise            *string `gorm:"size:65535" json:"noise"`
	BedAdjustable    *bool   `json:"bed_adjustable"`
	PrivateBathroom  *bool   `json:"privateBathroom"`
	BathroomDesc     *string `gorm:"size:65535" json:"bathroomDesc"`
	NumFlag          *bool   `json:"numFlag"` // No clue what this does
	Picture          *string `json:"picture"`

	RoomType    string  `gorm:"not null" json:"roomType"`
	Faces       *string `json:"faces"`
	HC          *bool   `json:"hc"`
	FloorNumber *int    `json:"floorNumber"`
	Area        *int    `json:"area"`
	Walkthrough *bool   `json:"walkthrough"`

	// Has many students
	Users []*User `json:"users,omitempty"`

	// Has many Dormtrak reviews
	DormtrakReviews []*DormtrakReview `json:"dormtrakReviews,omitempty"`

	// NOTE: I removed all ratings for dorm rooms (wifi, etc.), as there was no use for them. They were unused
	// in the Rails version as well. Now, we just track the ratings in the dorm, which we update every time a review
	// is changed.
}

func (*DormRoom) TableName() string {
	return "dorm_rooms"
}

// This updates the dorm after we change the dorm room
func (r *DormRoom) AfterSave(tx *gorm.DB) (err error) {
	err = NewDormModel(tx).UpdateDormFacts(r.DormID)
	return
}
