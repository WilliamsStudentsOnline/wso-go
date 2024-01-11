package models

import (
	"encoding/json"
	"errors"
	"time"
)

type DiningHall string

const (
	DiningReviewDriscoll DiningHall = "Driscoll"
	DiningReviewMission  DiningHall = "Mission"
	DiningReviewWhitmans DiningHall = "Whitman's"
)

// Dining Review Schema
type DiningHallReview struct {
	BaseSchema

	// Belongs to user (student)
	UserID uint  `gorm:"index:index_dining_review_on_user_id;not null" json:"userID"`
	User   *User `json:"user,omitempty"`

	WouldRecommendFood *bool  `json:"wouldRecommendFood"`
	FoodQuality        *int   `json:"foodQuality"`
	WaitTime           *int   `json:"waitTime"`
	Comment            string `gorm:"size:65535" json:"comment"`
	Flagged            bool   `json:"flagged"`

	// Dining info data
	DiningHall DiningHall `json:"diningHall" enums:"Driscoll, Mission, Whitman's"` //Driscoll, Mission, Whitman's

	// Pass the created time: not looked at by GORM
	CreatedTime time.Time `gorm:"-" json:"createdTime"`

	// Pass if the client agreed with the survey; not looked at by GORM.
	// True means user agreed, false means user disagreed, and null/missing means user does not have any
	// agreement/disagreement.
	ClientAgreement *bool `gorm:"-" json:"clientAgreement,omitempty"`
}

func (*DiningHallReview) TableName() string {
	return "dining_reviews"
}

func NewDiningReview(id uint) *DiningHallReview {
	return &DiningHallReview{
		BaseSchema: BaseSchema{
			ID: id,
		},
	}
}

func (diningHall *DiningHall) UnmarshalJSON(b []byte) error {
	// Define a secondary type to avoid ending up with a recursive call to json.Unmarshal
	type DH DiningHall
	var r = (*DH)(diningHall)
	err := json.Unmarshal(b, &r)
	if err != nil {
		panic(err)
	}
	switch *diningHall {
	case DiningReviewDriscoll, DiningReviewMission, DiningReviewWhitmans:
		return nil
	}
	return errors.New("invalid dining hall")
}
