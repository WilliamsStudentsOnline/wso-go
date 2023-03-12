package models

const (
	ConditionNew uint = iota
	ConditionLikeNew
	ConditionVeryGood
	ConditionGood
	ConditionFair
	ConditionPoor
	ConditionMAX
)

type BookListing struct {
	BaseSchema

	// Belongs to book
	BookID uint  `gorm:"index:index_book_listings_on_book_id;not null" json:"bookID"`
	Book   *Book `json:"books,omitempty"`

	// Belongs to user (student)
	UserID uint  `gorm:"index:index_book_listings_on_user_id;not null" json:"userID"`
	User   *User `json:"user,omitempty"`

	Condition   uint    `json:"condition"`
	Description *string `gorm:"size:65535" json:"description"`

	// True -> Offering to buy, False -> Offering to sell
	IsBuyListing bool `json:"-"`
}

func (*BookListing) TableName() string {
	return "book_listings"
}
