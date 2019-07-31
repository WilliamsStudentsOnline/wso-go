package factrak

import (
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

// List all areas of study
// ListAreasOfStudy godoc
// @Summary List areas of study
// @Description lists all areas of study
// @ID factrak-list-areas-of-study
// @Tags factrak
// @Accept  json
// @Produce  json
// @Success 200 {array} models.AreaOfStudy
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /factrak/areas-of-study [get]
func (t *Controller) ListAreasOfStudy(c *gin.Context) {
	var areas []models.AreaOfStudy
	err := t.areaOfStudyModel.GetAllAreasOfStudy(&areas)

	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, areas)
}

// Get one area of study
// GetAreaOfStudy godoc
// @Summary Get area of study
// @Description get one area of study with department preload
// @ID factrak-get-area-of-study
// @Tags factrak
// @Accept  json
// @Produce  json
// @Param areaOfStudyID path uint true "Area of Study ID"
// @Success 200 {object} models.AreaOfStudy
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /factrak/areas-of-study/{areaOfStudyID} [get]
func (t *Controller) GetAreaOfStudy(c *gin.Context) {
	// Decode areaOfStudyID.
	areaID, err := services.GetUIntParam(c, "areaOfStudyID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Do database query
	var area models.AreaOfStudy
	err = t.areaOfStudyModel.GetAreaOfStudyByID(areaID, &area)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, area)
}

// List area of study's professors
// ListAreaOfStudyProfessors godoc
// @Summary List area of study professors
// @Description list one area of study's professors
// @ID factrak-list-area-of-study-professors
// @Tags factrak
// @Accept  json
// @Produce  json
// @Param areaOfStudyID path uint true "Area of Study ID"
// @Success 200 {array} models.User
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /factrak/areas-of-study/{areaOfStudyID}/professors [get]
func (t *Controller) ListAreaOfStudyProfessors(c *gin.Context) {
	// Decode areaOfStudyID.
	areaID, err := services.GetUIntParam(c, "areaOfStudyID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Check if area exists
	exists, err := t.areaOfStudyModel.DoesAreaOfStudyExist(areaID)
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

	err = t.professorModel.GetProfessorsByAreaOfStudy(areaID, &profs)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, profs)
}

// List area of study's courses.
// ListAreaOfStudyCourses godoc
// @Summary List area of study courses
// @Description list one area of study's courses
// @ID factrak-list-area-of-study-courses
// @Tags factrak
// @Accept  json
// @Produce  json
// @Param areaOfStudyID path uint true "Area of Study ID"
// @Success 200 {array} models.Course
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /factrak/areas-of-study/{areaOfStudyID}/courses [get]
func (t *Controller) ListAreaOfStudyCourses(c *gin.Context) {
	// Decode areaOfStudyID.
	areaID, err := services.GetUIntParam(c, "areaOfStudyID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Check if area exists
	exists, err := t.areaOfStudyModel.DoesAreaOfStudyExist(areaID)
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

	err = t.courseModel.GetCoursesByAreaOfStudy(areaID, &courses)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, courses)
}
