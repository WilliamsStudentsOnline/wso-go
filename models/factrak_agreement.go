package models

import (
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// FactrakAgreement Model
type FactrakAgreementModel struct {
	*BaseModel
}

func NewFactrakAgreementModel(db *gorm.DB, log *zap.SugaredLogger) *FactrakAgreementModel {
	return &FactrakAgreementModel{
		BaseModel: NewBaseModel(db, log),
	}
}

// Gets factrak agreement by its user id and survey id.
func (m *FactrakAgreementModel) GetAgreementByUserAndSurvey(userID uint, surveyID uint, p *FactrakAgreement) (err error) {
	err = m.DB.Where(&FactrakAgreement{
		UserID: userID, FactrakSurveyID: surveyID,
	}).First(p).Error
	return
}

// Checks if the factrak agreement by its user id and survey id exists.
func (m *FactrakAgreementModel) DoesAgreementByUserAndSurveyExist(userID uint, surveyID uint) (exists bool, err error) {
	var count int
	err = m.DB.Model(&FactrakAgreement{}).Where(&FactrakAgreement{
		UserID: userID, FactrakSurveyID: surveyID,
	}).Count(&count).Error
	exists = count > 0
	return
}

// Creates a new factrak agreement
func (m *FactrakAgreementModel) CreateAgreement(agreement *FactrakAgreement) (err error) {
	err = m.DB.Create(agreement).Error
	return
}

// Updates a factrak agreement
func (m *FactrakAgreementModel) UpdateAgreement(agreement *FactrakAgreement) (err error) {
	err = m.DB.Save(agreement).Error
	return
}

// Deletes a factrak agreement permanently
func (m *FactrakAgreementModel) DeleteAgreement(agreement *FactrakAgreement) (err error) {
	err = m.DB.Unscoped().Delete(agreement).Error
	return
}
