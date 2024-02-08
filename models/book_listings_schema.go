package models

import (
	"encoding/json"
	"errors"
)

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

	Condition   Condition `json:"condition" enums:",POOR,FAIR,GOOD,VERY_GOOD,LIKE_NEW,NEW"`
	Description *string   `gorm:"size:65535" json:"description"`

	ListingType ListingType `json:"listingType" enums:",BUY,SELL"`
}

func (*BookListing) TableName() string {
	return "book_listings"
}

func (condition *Condition) UnmarshalJSON(b []byte) error {
	// Define a secondary type to avoid ending up with a recursive call to json.Unmarshal
	type C Condition
	var r = (*C)(condition)
	err := json.Unmarshal(b, &r)
	if err != nil {
		panic(err)
	}
	switch *condition {
	case ConditionUndefined, ConditionPoor, ConditionFair, ConditionGood, ConditionVeryGood, ConditionLikeNew, ConditionNew:
		return nil
	}
	return errors.New("invalid condition")
}

func (listingType *ListingType) UnmarshalJSON(b []byte) error {
	// Define a secondary type to avoid ending up with a recursive call to json.Unmarshal
	type LT ListingType
	var r = (*LT)(listingType)
	err := json.Unmarshal(b, &r)
	if err != nil {
		panic(err)
	}
	switch *listingType {
	case ListingTypeUndefined, ListingTypeBuy, ListingTypeSell:
		return nil
	}
	return errors.New("invalid listing type")
}
