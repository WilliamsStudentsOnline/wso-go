package factrak

import (
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

// List user's surveys
func (t *Controller) ListUserSurveys(c *gin.Context) {
	// Decode professorID.
	userID, err := services.GetUIntParam(c, "userID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Check if user exists
	exists, err := t.userModel.DoesUserExist(userID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if !exists {
		t.RespondError(c, lib.ErrorRecordNotFound)
		return
	}

	// Can only be view self (unless admin or factrak admin)
	if userID != services.GetUserID(c) || auth.HasScope(c, auth.ScopeAdminAll, auth.ScopeAdminFactrak) {
		t.RespondError(c, lib.ErrorMustBeSelf)
		return
	}

	// Do database query
	var surveys []*models.FactrakSurvey

	err = t.surveyModel.GetSurveysByAuthor(userID, &surveys)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, surveys)
}
