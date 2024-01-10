package models

import "time"

const (
	DiningReviewDriscoll = "Driscoll"
	DiningReviewMission  = "Mission"
	DiningReviewWhitmans = "Whitman's"
)

// Dining Review Schema
type DiningReview struct {
	BaseSchema

	// Belongs to user (student)
	UserID uint  `gorm:"index:index_dining_review_on_user_id;not null" json:"userID"`
	User   *User `json:"user,omitempty"`

	// Belongs to dining hall
	DiningID uint  `gorm:"index:index_dining_reviews_on_dining_id;not null" json:"diningID"`
	Dining   *User `gorm:"foreignkey:DiningID" json:"dining,omitempty"`

	WouldRecommendFood *bool  `json:"wouldRecommendFood"`
	FoodQuality        *int   `json:"foodQuality"`
	WaitTime           *int   `json:"waitTime"`
	Comment            string `gorm:"size:65535" json:"comment"`
	Flagged            bool   `json:"flagged"`

	// Dining info data
	DiningHall *string `json:"diningHall"` //Driscoll, Mission, Whitman's

	// Pass the created time: not looked at by GORM
	CreatedTime time.Time `gorm:"-" json:"createdTime"`

	// Pass if the client agreed with the survey; not looked at by GORM.
	// True means user agreed, false means user disagreed, and null/missing means user does not have any
	// agreement/disagreement.
	ClientAgreement *bool `gorm:"-" json:"clientAgreement,omitempty"`
}

func (*DiningReview) TableName() string {
	return "dining_reviews"
}

func NewDiningReview(id uint) *DiningReview {
	return &DiningReview{
		BaseSchema: BaseSchema{
			ID: id,
		},
	}
}
