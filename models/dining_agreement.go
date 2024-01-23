package models

import (
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// DiningReviewAgreement Model
type DiningReviewAgreementModel struct {
	*BaseModel
}

func NewDiningReviewAgreementModel(db *gorm.DB, log *zap.SugaredLogger) *DiningReviewAgreementModel {
	return &DiningReviewAgreementModel{
		BaseModel: NewBaseModel(db, log),
	}
}

// Gets dining review agreement by its user id and survey id.
func (m *DiningReviewAgreementModel) GetAgreementByUserAndSurvey(userID uint, surveyID uint, p *DiningReviewAgreement) (err error) {
	err = m.DB.Where(&DiningReviewAgreement{
		UserID: userID, DiningReviewID: surveyID,
	}).First(p).Error
	return
}

// Checks if the dining review agreement by its user id and survey id exists.
func (m *DiningReviewAgreementModel) DoesAgreementByUserAndSurveyExist(userID uint, surveyID uint) (exists bool, err error) {
	var count int
	err = m.DB.Model(&DiningReviewAgreement{}).Where(&DiningReviewAgreement{
		UserID: userID, DiningReviewID: surveyID,
	}).Count(&count).Error
	exists = count > 0
	return
}

// Creates a new dining review agreement
func (m *DiningReviewAgreementModel) CreateDiningAgreement(agreement *DiningReviewAgreement) (err error) {
	err = m.DB.Create(agreement).Error
	return
}

// Updates a dining review agreement
func (m *DiningReviewAgreementModel) UpdateDiningAgreement(agreement *DiningReviewAgreement) (err error) {
	err = m.DB.Save(agreement).Error
	return
}

// Deletes a dining review permanently
func (m *DiningReviewAgreementModel) DeleteDiningAgreement(agreement *DiningReviewAgreement) (err error) {
	err = m.DB.Unscoped().Delete(agreement).Error
	return
}
