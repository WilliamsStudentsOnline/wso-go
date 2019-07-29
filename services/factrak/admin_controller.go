package factrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

// List all surveys
func (t *Controller) ListFlaggedSurveys(c *gin.Context) {
	var surveys []*models.FactrakSurvey
	err := t.surveyModel.GetAllFlaggedSurveys(&surveys)

	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, surveys)
}

// Unflag survey by mods
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

	// Do DB flag
	err = t.surveyModel.SetSurveyFlag(surveyID, false)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// We know user is admin, so don't need to delete user fields
	t.RespondOK(c, nil)
}
