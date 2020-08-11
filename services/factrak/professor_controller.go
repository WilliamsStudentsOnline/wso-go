package factrak

import (
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

// List all professors. This, like all methods here, scopes professors to AtWilliams = true.
// ListProfessors godoc
// @Summary List professors
// @Description lists all professors at Williams
// @ID factrak-list-professors
// @Tags factrak
// @Accept  json
// @Produce  json
// @Param offset query int false "Offset Pagination"
// @Param limit query int false "Limit Pagination"
// @Param preload query []string false "Preload List"
// @Param courseID query int false "Course ID"
// @Param departmentID query int false "Department ID"
// @Param areaOfStudyID query int false "Area Of Study ID"
// @Param q query string false "Search Query"
// @Success 200 {array} models.User
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /factrak/professors [get]
func (t *Controller) ListProfessors(c *gin.Context) {
	var profs []*models.User
	var err error

	opts := models.GetAllProfessorsOptions{}
	if err = c.ShouldBindQuery(&opts); err != nil {
		t.RespondBadBind(c, err)
		return
	}

	if query, ok := c.GetQuery("q"); ok {
		err = t.factrakSearch.SearchProfessors(query, &profs, &opts)
	} else if sort, ok := c.GetQuery("sort"); ok {
		err = t.professorModel.GetProfessorsRanked(sort, &profs, &opts)
	} else {
		err = t.professorModel.GetAllProfessors(&profs, &opts)
	}

	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, profs)
}

// Lists all professors in order of a metric. This scopes professors to AtWilliams = true.
// ListProfessorsRanked godoc
// @Summary Lists ranked professors
// @Description Lists all professors at Williams ranked by a survey metric
// @ID factrak-list-professors-ranked
// @Tags factrak
// @Accept json
// @Produce json
// @Param offset query int false "Offset Pagination"
// @Param limit query int false "Limit Pagination"
// @Param preload query []string false "Preload List"
// @Param courseID query int false "Course ID"
// @Param departmentID query int false "Department ID"
// @Param areaOfStudyID query int false "Area Of Study ID"
// @Param metric path string true "Ranking Metric"
// @Param direction query bool false "Direction of Ordering"
// @Success 200 {array} models.User
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /factrak/professors/ranked/{metric} [get]
func (t *Controller) ListProfessorsRanked(c *gin.Context) {
	var profs []*models.User
	var err error

	opts := models.GetAllProfessorsOptions{}
	if err = c.ShouldBindQuery(&opts); err != nil {
		t.RespondBadBind(c, err)
		return
	}
	if sort, ok := c.GetQuery("sort"); ok {
		err = t.professorModel.GetProfessorsRanked(sort, &profs, &opts)
	}

	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, profs)
}

// Get a professor. May pass an optional "?courseID=XX" parameter to limit preload (ProfessorFactrakSurveys)
// scope to a professor and a course.
// GetProfessor godoc
// @Summary Get professor
// @Description get one course with factrak surveys, area of study preloaded,
// @Description May pass an optional "?courseID=XX" parameter to limit preload scope to a professor and a course.
// @ID factrak-get-professor
// @Tags factrak
// @Accept  json
// @Produce  json
// @Param courseID query uint false "Course ID"
// @Param professorID path uint true "Professor ID"
// @Success 200 {object} models.User
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /factrak/professors/{professorID} [get]
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
// ListProfessorSurveys godoc
// @Summary List professor surveys
// @Description list one professor's surveys
// @ID factrak-list-professor-surveys
// @Tags factrak
// @Accept  json
// @Produce  json
// @Param professorID path uint true "Professor ID"
// @Param offset query string false "Offset Pagination (timestamp)"
// @Param limit query int false "Limit Pagination"
// @Param courseID query int false "Course ID"
// @Param preload query []string false "Preload (course, professor)"
// @Param populateAgreements query bool false "Populate Agreement Counts"
// @Param populateClientAgreement query bool false "Populate Client's Agreement"
// @Success 200 {array} models.FactrakSurvey
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /factrak/professors/{professorID}/surveys [get]
// @Deprecated
func (t *Controller) ListProfessorSurveys(c *gin.Context) {
	userID := services.GetUserID(c)
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

	params := models.GetAllFactrakSurveysOptions{}
	err = c.ShouldBindQuery(&params)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	params.ProfAtWilliams = false
	params.UserID = nil
	params.ProfessorID = &profID
	params.ClientAgreementUserID = userID

	// Do database query
	var surveys []*models.FactrakSurvey
	// We already know prof is at williams, so we don't need to do the join
	err = t.surveyModel.GetAllSurveys(&surveys, &params)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	RemoveUserIDFromSurveys(c, surveys)

	t.RespondOK(c, surveys)
}

// List professor's courses.
// ListProfessorCourses godoc
// @Summary List professor courses
// @Description list one professor's courses
// @ID factrak-list-professor-courses
// @Tags factrak
// @Accept  json
// @Produce  json
// @Param professorID path uint true "Professor ID"
// @Success 200 {array} models.Course
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /factrak/professors/{professorID}/courses [get]
// @Deprecated
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
	var courses []*models.Course

	err = t.courseModel.GetCoursesByProfessor(profID, &courses)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, courses)
}

// Gets average ratings for a professor. May pass an optional "?courseID=XX" parameter to limit scope to a
// professor and a course.
// @Summary Get professor ratings
// @Description get one professor's ratings
// @ID factrak-get-professor-ratings
// @Tags factrak
// @Accept  json
// @Produce  json
// @Param courseID query uint false "Course ID"
// @Param professorID path uint true "Professor ID"
// @Success 200 {object} models.FactrakSurveyAvgRatings
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /factrak/professors/{professorID}/ratings [get]
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
