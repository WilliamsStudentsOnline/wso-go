package factrak

import (
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

// List all professors
func (t *Controller) ListProfessors(c *gin.Context) {
	var profs []models.User
	err := t.professorModel.GetAllProfessors(&profs)

	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, profs)
}

// Get one professor
func (t *Controller) GetProfessor(c *gin.Context) {
	// Decode professorID.
	profID, err := services.GetUIntParam(c, "professorID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
	}

	// Do database query
	var prof models.User
	err = t.professorModel.GetProfessorByID(profID, &prof)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	if !prof.AtWilliams {
		t.RespondAPIError(c, lib.ErrorUserNotAtWilliams)
		return
	}

	RemoveUserIDFromSurveys(c, prof.ProfessorFactrakSurveys)

	t.RespondOK(c, prof)
}

// List professor's surveys
func (t *Controller) ListProfessorSurveys(c *gin.Context) {
	// Decode professorID.
	profID, err := services.GetUIntParam(c, "professorID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
	}

	// Check if professor exists
	exists, err := t.professorModel.DoesProfessorExist(profID)
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

	err = t.surveyModel.GetSurveysByProfessor(profID, &surveys)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	RemoveUserIDFromSurveys(c, surveys)

	t.RespondOK(c, surveys)
}

// List professor's courses
func (t *Controller) ListProfessorCourses(c *gin.Context) {
	// Decode professorID.
	profID, err := services.GetUIntParam(c, "professorID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
	}

	// Check if professor exists
	exists, err := t.professorModel.DoesProfessorExist(profID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if !exists {
		t.RespondError(c, lib.ErrorRecordNotFound)
		return
	}

	// Do database query
	var surveys []models.Course

	err = t.courseModel.GetCoursesByProfessor(profID, &surveys)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, surveys)
}