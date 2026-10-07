package models

// BoardRideMeta holds ride-specific fields for board threads of type ride.
type BoardRideMeta struct {
	BoardBaseSchema

	ThreadID uint         `gorm:"unique_index:index_board_ride_meta_on_thread_id;not null" json:"threadID"`
	Thread   *BoardThread `json:"thread,omitempty"`

	OfferingRide bool   `gorm:"not null" json:"offeringRide"`
	Source       string `gorm:"not null" json:"source"`
	Destination  string `gorm:"not null" json:"destination"`
}

func (*BoardRideMeta) TableName() string {
	return "board_ride_meta"
}
