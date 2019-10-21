// This is nominally in its own file, but actually is part of the larger SurveysController.
package factrak

import (
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// Get agreement if user has one
// GetSurveyAgreement godoc
// @Summary Get survey agreement by self
// @Description gets the agreement about a survey made by the self user
// @ID factrak-get-agreement
// @Tags factrak
// @Accept  json
// @Produce  json
// @Param surveyID path uint true "Survey ID"
// @Success 200 {object} models.FactrakAgreement
// @Failure 1551 {object} lib.APIError "survey agreement could not be found"
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /factrak/surveys/{surveyID}/agreement [get]
func (t *Controller) GetAgreement(c *gin.Context) {
	userID := services.GetUserID(c)

	// Decode surveyID.
	surveyID, err := services.GetUIntParam(c, "surveyID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Ensure survey exists
	exists, err := t.surveyModel.DoesSurveyExist(surveyID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if !exists {
		t.RespondError(c, lib.ErrorRecordNotFound)
		return
	}

	// Do database query
	var agreement models.FactrakAgreement
	err = t.agreementModel.GetAgreementByUserAndSurvey(userID, surveyID, &agreement)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			// We have a different response for survey agreement not being found as agreement is an object of
			// survey, rather than its own thing. It is still a 404 not found, however.
			t.RespondError(c, lib.ErrorSurveyAgreementNotFound)
			return
		}
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, agreement)
}

// Parameters for POST agreement
type AgreementCreateParams struct {
	Agree *bool `json:"agree"`
}

// Create an agreement if user does not have one
// CreateSurveyAgreement godoc
// @Summary Create survey agreement
// @Description creates an agreement about a survey made by the self user
// @ID factrak-create-agreement
// @Tags factrak
// @Accept  json
// @Produce  json
// @Param surveyID path uint true "Survey ID"
// @Param createParams body factrak.AgreementCreateParams true "Create Agreement Params"
// @Success 201 {object} models.FactrakAgreement
// @Failure 1553 {object} lib.APIError "cannot create survey agreement with your own survey"
// @Failure 1552 {object} lib.APIError "survey agreement already exists for this user and survey"
// @Failure 1100 {object} lib.APIError "could not parse malformed request data"
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /factrak/surveys/{surveyID}/agreement [post]
func (t *Controller) CreateAgreement(c *gin.Context) {
	userID := services.GetUserID(c)

	// Decode surveyID.
	surveyID, err := services.GetUIntParam(c, "surveyID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Get survey
	var survey models.FactrakSurvey
	err = t.surveyModel.GetSurveyByID(surveyID, &survey)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Cannot make agreement on own survey, so error if we try to
	if survey.UserID == userID {
		t.RespondError(c, lib.ErrorSurveyAgreementNoSelf)
		return
	}

	// Bind create params
	createData := AgreementCreateParams{}
	err = c.ShouldBind(&createData)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	if createData.Agree == nil {
		t.RespondError(c, lib.ErrorMalformedRequestData)
		return
	}

	// Check if agreement already exists
	exists, err := t.agreementModel.DoesAgreementByUserAndSurveyExist(userID, surveyID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if exists {
		t.RespondError(c, lib.ErrorSurveyAgreementAlreadyExists)
		return
	}

	// Do db create
	agreement := models.FactrakAgreement{
		UserID:          userID,
		FactrakSurveyID: surveyID,
		Agrees:          *createData.Agree,
	}
	err = t.agreementModel.CreateAgreement(&agreement)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondCreated(c, agreement)
}

type AgreementUpdateParams struct {
	Agree *bool `json:"agree"`
}

// Update an agreement if user has one
// UpdateSurveyAgreement godoc
// @Summary Update survey agreement
// @Description updates an agreement about a survey made by the self user
// @ID factrak-update-agreement
// @Tags factrak
// @Accept  json
// @Produce  json
// @Param surveyID path uint true "Survey ID"
// @Param updateParams body factrak.AgreementUpdateParams true "Update Agreement Params"
// @Success 200 {object} models.FactrakAgreement
// @Failure 1551 {object} lib.APIError "survey agreement could not be found"
// @Failure 1100 {object} lib.APIError "could not parse malformed request data"
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /factrak/surveys/{surveyID}/agreement [patch]
func (t *Controller) UpdateAgreement(c *gin.Context) {
	userID := services.GetUserID(c)

	// Decode surveyID.
	surveyID, err := services.GetUIntParam(c, "surveyID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Ensure survey exists
	exists, err := t.surveyModel.DoesSurveyExist(surveyID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if !exists {
		t.RespondError(c, lib.ErrorRecordNotFound)
		return
	}

	// Bind update params
	updateData := AgreementUpdateParams{}
	err = c.ShouldBind(&updateData)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	if updateData.Agree == nil {
		t.RespondError(c, lib.ErrorMalformedRequestData)
		return
	}

	// Check if agreement already exists
	var agreement models.FactrakAgreement
	err = t.agreementModel.GetAgreementByUserAndSurvey(userID, surveyID, &agreement)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			// We have a different response for survey agreement not being found as agreement is an object of
			// survey, rather than its own thing. It is still a 404 not found, however.
			t.RespondError(c, lib.ErrorSurveyAgreementNotFound)
			return
		}
		t.RespondError(c, err)
		return
	}

	// Do db update
	agreement.Agrees = *updateData.Agree
	err = t.agreementModel.UpdateAgreement(&agreement)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, agreement)
}

// Delete an agreement if user has one
// DeleteSurveyAgreement godoc
// @Summary Delete survey agreement
// @Description deletes an agreement about a survey made by the self user
// @ID factrak-delete-agreement
// @Tags factrak
// @Accept  json
// @Produce  json
// @Param surveyID path uint true "Survey ID"
// @Success 200 {object} models.FactrakAgreement
// @Failure 1551 {object} lib.APIError "survey agreement could not be found"
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /factrak/surveys/{surveyID}/agreement [delete]
func (t *Controller) DeleteAgreement(c *gin.Context) {
	userID := services.GetUserID(c)

	// Decode surveyID.
	surveyID, err := services.GetUIntParam(c, "surveyID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Ensure survey exists
	exists, err := t.surveyModel.DoesSurveyExist(surveyID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if !exists {
		t.RespondError(c, lib.ErrorRecordNotFound)
		return
	}

	// Check if agreement already exists
	var agreement models.FactrakAgreement
	err = t.agreementModel.GetAgreementByUserAndSurvey(userID, surveyID, &agreement)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			// We have a different response for survey agreement not being found as agreement is an object of
			// survey, rather than its own thing. It is still a 404 not found, however.
			t.RespondError(c, lib.ErrorSurveyAgreementNotFound)
			return
		}
		t.RespondError(c, err)
		return
	}

	// Do db delete
	err = t.agreementModel.DeleteAgreement(&agreement)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, agreement)
}
