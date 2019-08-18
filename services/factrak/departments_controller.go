package factrak

import (
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

// List all departments
// ListDepartments godoc
// @Summary List departments
// @Description lists all departments
// @ID factrak-list-departments
// @Tags factrak
// @Accept  json
// @Produce  json
// @Success 200 {array} models.Department
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /factrak/departments [get]
func (t *Controller) ListDepartments(c *gin.Context) {
	var depts []models.Department
	err := t.departmentModel.GetAllDepartments(&depts)

	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, depts)
}

// Get one department
// GetDepartment godoc
// @Summary Get department
// @Description get one department with areas of study preloaded
// @ID factrak-get-department
// @Tags factrak
// @Accept  json
// @Produce  json
// @Param departmentID path uint true "Department ID"
// @Success 200 {object} models.Department
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /factrak/departments/{departmentID} [get]
func (t *Controller) GetDepartment(c *gin.Context) {
	// Decode departmentID.
	deptID, err := services.GetUIntParam(c, "departmentID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Do database query
	var dept models.Department
	err = t.departmentModel.GetDepartmentByID(deptID, &dept)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, dept)
}

// List department's professors
// ListDepartmentProfessors godoc
// @Summary List department professors
// @Description list one department's professors
// @ID factrak-list-department-professors
// @Tags factrak
// @Accept  json
// @Produce  json
// @Param departmentID path uint true "Department ID"
// @Success 200 {array} models.User
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /factrak/departments/{departmentID}/professors [get]
// @Deprecated
func (t *Controller) ListDepartmentProfessors(c *gin.Context) {
	// Decode departmentID.
	deptID, err := services.GetUIntParam(c, "departmentID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Check if dept exists
	exists, err := t.departmentModel.DoesDepartmentExist(deptID)
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

	err = t.professorModel.GetProfessorsByDepartment(deptID, &profs)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, profs)
}

// List department's courses.
// ListDepartmentCourses godoc
// @Summary List department courses
// @Description list one department's courses
// @ID factrak-list-department-courses
// @Tags factrak
// @Accept  json
// @Produce  json
// @Param departmentID path uint true "Department ID"
// @Success 200 {array} models.Course
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /factrak/departments/{departmentID}/courses [get]
// @Deprecated
func (t *Controller) ListDepartmentCourses(c *gin.Context) {
	// Decode departmentID.
	deptID, err := services.GetUIntParam(c, "departmentID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Check if dept exists
	exists, err := t.departmentModel.DoesDepartmentExist(deptID)
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

	err = t.courseModel.GetCoursesByDepartment(deptID, &courses)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, courses)
}
