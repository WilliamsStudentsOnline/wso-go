package models

import "github.com/jinzhu/gorm"

// FactrakAgreement Model
type FactrakAgreementModel struct {
	*BaseModel
}

func NewFactrakAgreementModel(db *gorm.DB) *FactrakAgreementModel {
	return &FactrakAgreementModel{
		BaseModel: NewBaseModel(db),
	}
}
