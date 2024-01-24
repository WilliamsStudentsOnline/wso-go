package models

// DiningHallReviewAgreement Schema
type DiningReviewAgreement struct {
	BaseSchema
	Agrees bool `json:"agrees"`

	// Belongs to survey
	DiningReviewID uint              `gorm:"index:index_dining_agreements_on_dining_review_survey_id;" json:"diningReviewID"`
	DiningReview   *DiningHallReview `json:"diningReview,omitempty"`

	// Belongs to user
	UserID uint  `gorm:"index:index_dining_agreements_on_user_id;" json:"userID"`
	User   *User `json:"user,omitempty"`
}

func (*DiningReviewAgreement) DiningAgreementTable() string {
	return "dining_review_agreements"
}
