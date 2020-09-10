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
// @Param offset query int false "Offset Pagination"
// @Param limit query int false "Limit Pagination"
// @Param preload query []string false "Preload List"
// @Param q query string false "Search Query"
// @Param areaOfStudyID query string false "Area of Study ID"
// @Param departmentID query string false "Department ID"
// @Param professorID query string false "Professor ID"
// @Param metric query string false "Ranking Metric"
// @Param direction query bool false "Sorting Direction"
// @Success 200 {array} models.Course
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /factrak/courses [get]
func (t *Controller) ListCourses(c *gin.Context) {
	var courses []*models.Course
	var err error

	opts := models.GetAllCoursesOptions{}
	if err = c.ShouldBindQuery(&opts); err != nil {
		t.RespondBadBind(c, err)
		return
	}

	if query, ok := c.GetQuery("q"); ok {
		err = t.factrakSearch.SearchCourses(query, &courses, &opts)
	} else if sort, ok := c.GetQuery("metric"); ok {
		err = t.courseModel.GetCoursesRanked(sort, &courses, &opts)
	} else {
		err = t.courseModel.GetAllCourses(&courses, &opts)
	}

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
// @Description May pass an optional "?professorID=XX" parameter to limit preload scope to a professor and a course.
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
// @Param courseID path uint true "Course ID"
// @Param offset query string false "Offset Pagination (timestamp)"
// @Param limit query int false "Limit Pagination"
// @Param professorID query int false "Professor ID"
// @Param preload query []string false "Preload (course, professor)"
// @Param populateAgreements query bool false "Populate Agreement Counts"
// @Param populateClientAgreement query bool false "Populate Client's Agreement"
// @Success 200 {array} models.FactrakSurvey
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /factrak/courses/{courseID}/surveys [get]
// @Deprecated
func (t *Controller) ListCourseSurveys(c *gin.Context) {
	// Decode courseID.
	userID := services.GetUserID(c)
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

	params := models.GetAllFactrakSurveysOptions{}
	err = c.ShouldBindQuery(&params)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	params.ProfAtWilliams = true
	params.UserID = nil
	params.CourseID = &courseID
	params.ClientAgreementUserID = userID

	// Do database query
	var surveys []*models.FactrakSurvey
	err = t.surveyModel.GetAllSurveys(&surveys, &params)
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
// @Deprecated
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
	var profs []*models.User

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
