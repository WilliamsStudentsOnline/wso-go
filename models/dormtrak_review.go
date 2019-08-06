package models

import "github.com/jinzhu/gorm"

// DormtrakReview Model
type DormtrakReviewModel struct {
	*BaseModel
}

func NewDormtrakReviewModel(db *gorm.DB) *DormtrakReviewModel {
	return &DormtrakReviewModel{
		BaseModel: NewBaseModel(db),
	}
}
