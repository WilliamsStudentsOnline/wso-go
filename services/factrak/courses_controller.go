package factrak

import (
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

// List all courses
// ListCourses godoc
// @Summary List courses
// @Description lists all courses
// @ID factrak-list-courses
// @Tags factrak
// @Accept  json
// @Produce  json
// @Success 200 {array} models.Course
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /factrak/courses [get]
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
// GetCourse godoc
// @Summary Get course
// @Description get one course with factrak surveys, area of study preloaded.
// @Description May pass an optional "?professorID=XXX" parameter to limit preload scope to a professor and a course.
// @ID factrak-get-course
// @Tags factrak
// @Accept  json
// @Produce  json
// @Param professorID query uint false "Professor ID"
// @Param courseID path uint true "Course ID"
// @Success 200 {object} models.Course
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /factrak/courses/{courseID} [get]
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

	// Remove surveys preload if limited scope
	if IsScopeLimited(c) {
		course.FactrakSurveys = nil
	}

	t.RespondOK(c, course)
}

// ListCourseSurveys godoc
// @Summary List course surveys
// @Description list one course's surveys
// @ID factrak-list-course-surveys
// @Tags factrak
// @Accept  json
// @Produce  json
// @Param professorID query uint false "Professor ID"
// @Param courseID path uint true "Course ID"
// @Param offset query time.Time false "Offset Pagination"
// @Param limit query int false "Limit Pagination"
// @Success 200 {array} models.FactrakSurvey
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /factrak/courses/{courseID}/surveys [get]
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

	pOff, pLim, err := GetSurveyPaginationParams(c)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	err = t.surveyModel.GetSurveysByProfessorOrCourse(profID, &courseID, true, &surveys, t.surveyModel.NewSurveyPaginate(pOff, pLim))
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

// List course's professors
// @Summary List course professors
// @Description list one course's professors
// @ID factrak-list-course-professors
// @Tags factrak
// @Accept  json
// @Produce  json
// @Param courseID path uint true "Course ID"
// @Success 200 {array} models.User
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /factrak/courses/{courseID}/professors [get]
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
// @Summary Get course ratings
// @Description get one course's ratings
// @ID factrak-get-course-ratings
// @Tags factrak
// @Accept  json
// @Produce  json
// @Param professorID query uint false "Professor ID"
// @Param courseID path uint true "Course ID"
// @Success 200 {object} models.FactrakSurveyAvgRatings
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /factrak/courses/{courseID}/ratings [get]
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
