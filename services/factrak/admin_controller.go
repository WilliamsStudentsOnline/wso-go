package factrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

// ListFlaggedSurveys godoc
// @Summary List flagged surveys
// @Description lists all surveys that are flagged
// @ID factrak-list-flagged-surveys
// @Tags factrak,factrak-admin,admin
// @Accept  json
// @Produce  json
// @Param professorID query int false "Professor ID"
// @Param courseID query int false "Course ID"
// @Param userID query int false "User ID"
// @Param offset query string false "Offset Pagination (timestamp)"
// @Param limit query int false "Limit Pagination"
// @Param preload query []string false "Preload (course, professor)"
// @Param populateAgreements query bool false "Populate Agreement Counts"
// @Param populateClientAgreement query bool false "Populate Client's Agreement"
// @Success 200 {array} models.FactrakSurvey
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /factrak/admin/surveys [get]
func (t *Controller) ListFlaggedSurveys(c *gin.Context) {
	userID := services.GetUserID(c)
	var surveys []*models.FactrakSurvey

	params := models.GetAllFactrakSurveysOptions{}
	err := c.ShouldBindQuery(&params)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Add the ClientAgreementUserID to the options
	params.ClientAgreementUserID = userID

	params.Flagged = true
	params.ProfAtWilliams = true

	err = t.surveyModel.GetAllSurveys(&surveys, &params)

	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, surveys)
}

// Unflag survey by mods
// UnflagSurvey godoc
// @Summary Unflag survey
// @Description unflags a flagged survey
// @ID factrak-unflag-survey
// @Tags factrak,factrak-admin,admin
// @Accept  json
// @Produce  json
// @Param surveyID path uint true "Survey ID"
// @Success 200
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /factrak/admin/surveys/{surveyID}/flag [delete]
func (t *Controller) UnflagSurvey(c *gin.Context) {
	surveyID, err := services.GetUIntParam(c, "surveyID")
	if err != nil {
		t.RespondError(c, err)
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

	// Do db flag
	err = t.surveyModel.SetSurveyFlag(surveyID, false)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// We know user is admin, so don't need to delete user fields
	t.RespondOK(c, nil)
}
