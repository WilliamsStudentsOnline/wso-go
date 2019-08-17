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
// ListUserSurveys godoc
// @Summary List user surveys
// @Description list one user's surveys
// @ID factrak-list-user-surveys
// @Tags factrak
// @Accept  json
// @Produce  json
// @Param userID path uint true "User ID"
// @Param offset query string false "Offset Pagination (timestamp)"
// @Param limit query int false "Limit Pagination"
// @Param professorID query int false "Professor ID"
// @Param courseID query int false "Course ID"
// @Param preload query []string false "Preload (course, professor)"
// @Success 200 {array} models.FactrakSurvey
// @Failure 1331 {object} lib.APIError "must be self"
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /factrak/users/{userID}/surveys [get]
// @Deprecated
func (t *Controller) ListUserSurveys(c *gin.Context) {
	// Decode userID.
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
	if userID != services.GetUserID(c) && !auth.HasScope(c, auth.ScopeAdminAll, auth.ScopeFactrakAdmin) {
		t.RespondError(c, lib.ErrorMustBeSelf)
		return
	}

	params := models.GetAllFactrakSurveysOptions{}
	err = c.ShouldBindQuery(&params)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Do database query
	var surveys []*models.FactrakSurvey
	err = t.surveyModel.GetSurveysByAuthor(userID, &surveys, &params)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, surveys)
}
