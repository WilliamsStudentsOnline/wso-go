package factrak

import (
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

// List all professors. This, like all methods here, scopes professors to AtWilliams = true.
func (t *Controller) ListProfessors(c *gin.Context) {
	var profs []models.User
	err := t.professorModel.GetAllProfessors(&profs)

	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, profs)
}

// Get a professor. May pass an optional "?courseID=XXX" parameter to limit preload (ProfessorFactrakSurveys)
// scope to a professor and a course.
func (t *Controller) GetProfessor(c *gin.Context) {
	// Decode professorID.
	profID, err := services.GetUIntParam(c, "professorID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	courseID := t.getQueryID(c, "courseID")
	if c.IsAborted() {
		return
	}

	// Do database query
	var prof models.User
	err = t.professorModel.GetProfessorByIDWithCourse(profID, &prof, courseID)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	RemoveUserIDFromSurveys(c, prof.FactrakSurveys)

	// Remove surveys preload if limited scope
	if IsScopeLimited(c) {
		prof.FactrakSurveys = nil
	}

	t.RespondOK(c, prof)
}

// List professor's surveys. May pass an optional "?courseID=XXX" parameter to limit scope to a professor and a course.
func (t *Controller) ListProfessorSurveys(c *gin.Context) {
	// Decode professorID.
	profID, err := services.GetUIntParam(c, "professorID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
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

	courseID := t.getQueryID(c, "courseID")
	if c.IsAborted() {
		return
	}

	// Do database query
	var surveys []*models.FactrakSurvey

	// We already know prof is at williams, so we don't need ot do the join
	err = t.surveyModel.GetSurveysByProfessorOrCourse(&profID, courseID, false, &surveys)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	err = t.surveyModel.PopulateAgreementCountsSlice(surveys)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	RemoveUserIDFromSurveys(c, surveys)

	t.RespondOK(c, surveys)
}

// List professor's courses.
func (t *Controller) ListProfessorCourses(c *gin.Context) {
	// Decode professorID.
	profID, err := services.GetUIntParam(c, "professorID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
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
	var courses []models.Course

	err = t.courseModel.GetCoursesByProfessor(profID, &courses)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, courses)
}

// Gets average ratings for a professor. May pass an optional "?courseID=XXX" parameter to limit scope to a
// professor and a course.
func (t *Controller) GetProfessorRatings(c *gin.Context) {
	// Decode professorID.
	profID, err := services.GetUIntParam(c, "professorID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
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

	courseID := t.getQueryID(c, "courseID")
	if c.IsAborted() {
		return
	}

	// Do database query
	var ratings models.FactrakSurveyAvgRatings

	err = t.surveyModel.GetSurveyRatingsByProfessorOrCourse(&profID, courseID, &ratings)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, ratings)
}
