package factrak

import (
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

// List all departments
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
	var profs []models.User

	err = t.professorModel.GetProfessorsByDepartment(deptID, &profs)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, profs)
}

// List department's courses.
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
	var courses []models.Course

	err = t.courseModel.GetCoursesByDepartment(deptID, &courses)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, courses)
}
