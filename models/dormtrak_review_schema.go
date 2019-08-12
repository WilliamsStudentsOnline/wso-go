package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

// DormtrakReview Schema
type DormtrakReview struct {
	BaseSchema

	// Belongs to user (student)
	UserID uint  `gorm:"index:index_dormtrak_reviews_on_user_id;not null" json:"userID"`
	User   *User `json:"user,omitempty"`

	// Belongs to dorm room
	DormRoomID uint      `gorm:"index:index_dormtrak_reviews_on_dorm_room_id;not null" json:"dormRoomID"`
	DormRoom   *DormRoom `json:"dormRoom,omitempty"`

	Comment *string `gorm:"size:65535" json:"comment"`

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
	PrivateBathroom  *bool   `json:"privateBathroom"`
	BathroomDesc     *string `gorm:"size:65535" json:"bathroomDesc"`
	Comfort          *int    `json:"comfort"`
	Loudness         *int    `json:"loudness"`
	Convenience      *int    `json:"convenience"`
	Wifi             *int    `json:"wifi"`
	Location         *int    `json:"location"`
	Satisfaction     *int    `json:"satisfaction"`

	// Removing the anonymous column  and just treating it as anonymous by default

	// Pass the created time: not looked at by GORM
	CreatedTime time.Time `gorm:"-" json:"createdTime"`
}

func (*DormtrakReview) TableName() string {
	return "dormtrak_reviews"
}

func NewDormtrakReview(id uint) *DormtrakReview {
	return &DormtrakReview{
		BaseSchema: BaseSchema{
			ID: id,
		},
	}
}

// Again, I hate hooks but this is the best way.
// This populates the createdTime field: please don't use this field for database updates.
func (m *DormtrakReview) AfterFind() (err error) {
	m.CreatedTime = m.CreatedAt
	return
}

// This updates the dorm room after we update the review. It also chains to update the dorm as well
func (r *DormtrakReview) AfterSave(tx *gorm.DB) (err error) {
	err = NewDormRoomModel(tx).ReloadStatistics(r.DormRoomID)
	return
}

// This updates the dorm room after we delete a review. It also chains to update the dorm as well
func (r *DormtrakReview) AfterDelete(tx *gorm.DB) (err error) {
	err = NewDormRoomModel(tx).ReloadStatistics(r.DormRoomID)
	return
}
