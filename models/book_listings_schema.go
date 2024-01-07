package models

type Condition string
type ListingType string

const (
	ConditionUndefined Condition = ""
	ConditionPoor      Condition = "POOR"
	ConditionFair      Condition = "FAIR"
	ConditionGood      Condition = "GOOD"
	ConditionVeryGood  Condition = "VERY_GOOD"
	ConditionLikeNew   Condition = "LIKE_NEW"
	ConditionNew       Condition = "NEW"
)

const (
	ListingTypeUndefined ListingType = ""
	ListingTypeBuy       ListingType = "BUY"
	ListingTypeSell      ListingType = "SELL"
)

type BookListing struct {
	BaseSchema

	// Belongs to book (FK)
	BookID uint  `gorm:"index:index_book_listings_on_book_id;not null" json:"bookID"`
	Book   *Book `json:"book,omitempty"`

	// Belongs to user (FK)
	UserID uint  `gorm:"index:index_book_listings_on_user_id;not null" json:"userID"`
	User   *User `json:"user,omitempty"`

	Condition   Condition `json:"condition"`
	Description *string   `gorm:"size:65535" json:"description"`

	ListingType ListingType `json:"listingType"`
}

func (*BookListing) TableName() string {
	return "book_listings"
}
