package models

// DormtrakReview Schema
type DormtrakReview struct {
	BaseSchema

	// Belongs to user (student)
	UserID uint  `gorm:"index:index_dormtrak_reviews_on_user_id;not null" json:"userID"`
	User   *User `json:"user,omitempty"`

	// Belongs to dorm room
	DormRoomID uint      `gorm:"index:index_dormtrak_reviews_on_dorm_room_id;not null" json:"dormRoomID"`
	DormRoom   *DormRoom `json:"dormRoom,omitempty"`

	Comment          *string `gorm:"size:65535" json:"comment"`
	LivedHere        *bool   `json:"livedHere"`
	Closet           *string `json:"closet"`
	ClosetDesc       *string `gorm:"size:65535" json:"closetDesc"`
	Flooring         *string `json:"flooring"`
	CommonRoomAccess *bool   `json:"commonRoomAccess"`
	CommonRoomDesc   *string `gorm:"size:65535" json:"commonRoomDesc"`
	ThermostatAccess *bool   `json:"thermostatAccess"`
	ThermostatDesc   *string `gorm:"size:65535" json:"thermostatDesc"`
	OutletsDesc      *string `gorm:"size:65535" json:"outletsDesc"`
	KeyOrCard        *string `json:"keyOrCard"`
	Noise            *string `gorm:"size:65535" json:"noise"`
	BedAdjustable    *bool   `json:"bedAdjustable"`
	Anonymous        *bool   `gorm:"not null" json:"anonymous"`
	Faces            *string `json:"faces"`
	PrivateBathroom  *bool   `json:"privateBathroom"`
	BathroomDesc     *string `gorm:"size:65535" json:"bathroomDesc"`
	Comfort          *int    `json:"comfort"`
	Loudness         *int    `json:"loudness"`
	Convenience      *int    `json:"convenience"`
	Wifi             *int    `json:"wifi"`
	Location         *int    `json:"location"`
	Satisfaction     *int    `json:"satisfaction"`
}

func (*DormtrakReview) TableName() string {
	return "dormtrak_reviews"
}
