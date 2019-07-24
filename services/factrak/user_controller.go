package factrak

import (
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

// List professor's surveys
func (t *Controller) ListUserSurveys(c *gin.Context) {
	// Decode professorID.
	userID, err := services.GetUIntParam(c, "userID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
	}

	// Check if professor exists
	exists, err := t.userModel.DoesUserExist(userID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if !exists {
		t.RespondError(c, lib.ErrorRecordNotFound)
		return
	}

	// Do database query
	var surveys []*models.FactrakSurvey

	err = t.surveyModel.GetSurveysByAuthor(userID, &surveys)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	RemoveUserIDFromSurveys(c, surveys)

	t.RespondOK(c, surveys)
}
