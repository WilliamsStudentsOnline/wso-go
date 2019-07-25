package factrak

import (
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

// List all courses
func (t *Controller) ListCourses(c *gin.Context) {
	var courses []models.Course
	err := t.courseModel.GetAllCourses(&courses)

	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, courses)
}

// Get one course
func (t *Controller) GetCourse(c *gin.Context) {
	// Decode courseID.
	courseID, err := services.GetUIntParam(c, "courseID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	profID := t.getQueryID(c, "professorID")
	if c.IsAborted() {
		return
	}

	// Do database query
	var course models.Course
	err = t.courseModel.GetCourseByIDWithProfessor(courseID, &course, profID)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	RemoveUserIDFromSurveys(c, course.FactrakSurveys)

	t.RespondOK(c, course)
}

func (t *Controller) ListCourseSurveys(c *gin.Context) {
	// Decode courseID.
	courseID, err := services.GetUIntParam(c, "courseID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	exists, err := t.courseModel.DoesCourseExist(courseID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if !exists {
		t.RespondError(c, lib.ErrorRecordNotFound)
		return
	}

	profID := t.getQueryID(c, "professorID")
	if c.IsAborted() {
		return
	}

	// Do database query
	var surveys []*models.FactrakSurvey

	err = t.surveyModel.GetSurveysByProfessorOrCourse(profID, &courseID, true, &surveys)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	RemoveUserIDFromSurveys(c, surveys)

	t.RespondOK(c, surveys)
}

// List course's professors
func (t *Controller) ListCourseProfessors(c *gin.Context) {
	// Decode courseID.
	courseID, err := services.GetUIntParam(c, "courseID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Check if course exists
	exists, err := t.courseModel.DoesCourseExist(courseID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if !exists {
		t.RespondError(c, lib.ErrorRecordNotFound)
		return
	}

	// Do database query
	var profs []models.User

	err = t.professorModel.GetProfessorsByCourse(courseID, &profs)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, profs)
}

// Gets average ratings for a course. May pass an optional "?professorID=XXX" parameter to limit scope to a
// course and professor
func (t *Controller) GetCourseRatings(c *gin.Context) {
	// Decode courseID.
	courseID, err := services.GetUIntParam(c, "courseID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Check if course exists
	exists, err := t.courseModel.DoesCourseExist(courseID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if !exists {
		t.RespondError(c, lib.ErrorRecordNotFound)
		return
	}

	profID := t.getQueryID(c, "professorID")
	if c.IsAborted() {
		return
	}

	// Do database query
	var ratings models.FactrakSurveyAvgRatings

	err = t.surveyModel.GetSurveyRatingsByProfessorOrCourse(profID, &courseID, &ratings)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, ratings)
}
