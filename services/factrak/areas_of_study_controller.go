package factrak

import (
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

// List all areas of study
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
