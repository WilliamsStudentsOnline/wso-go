package factrak_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
)

// Check that a slice of departments matches the expected slice, by comparing course IDs
func EqualDepartmentIDs(expected, resp []models.Department) error {
	if len(resp) != len(expected) {
		return errors.New(fmt.Sprintf("Expected length %d, got %d", len(expected), len(resp)))
	}

	for i, v := range expected {
		if v.ID != resp[i].ID {
			return errors.New(fmt.Sprintf("At index %d, expected ID %d, got %d", i, v.ID, resp[i].ID))
		}
	}
	return nil
}

// Unmarshal a slice of departments from an http response
func GetDepartmentsFromResp(assert *testify.Assertions, w *httptest.ResponseRecorder) []models.Department {
	respData := utils.GetGoodResp(assert, w)
	var resp []models.Department
	assert.NoError(json.Unmarshal(respData.Data, &resp))
	return resp
}

// Unmarshal a single department from an http response
func GetDepartmentFromResp(assert *testify.Assertions, w *httptest.ResponseRecorder) models.Department {
	respData := utils.GetGoodResp(assert, w)
	var resp models.Department
	assert.NoError(json.Unmarshal(respData.Data, &resp))
	return resp
}

func TestController_ListDepartments(t *testing.T) {
	// Set up the test environment
	assert, db, router := SetupFactrakTest(t)

	d1 := models.Department{
		Name: "Computer Science",
	}
	d2 := models.Department{
		Name: "Economics",
	}
	assert.NoError(db.Create(&d1).Create(&d2).Error)

	// Get all departments
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/departments", nil)
	assert.NoError(err)

	// Check if response has correct departments
	resp := GetDepartmentsFromResp(assert, w)
	assert.NoError(EqualDepartmentIDs([]models.Department{d1, d2}, resp))
	assert.Len(resp, 2)
	assert.Equal(d1.Name, resp[0].Name)
	assert.Equal(d2.Name, resp[1].Name)
}

func TestController_GetDepartment(t *testing.T) {
	// Set up the test environment
	assert, db, router := SetupFactrakTest(t)

	d1 := models.Department{
		Name: "Computer Science",
		AreasOfStudy: []*models.AreaOfStudy{
			{
				Name:         "Computer Science",
				Abbreviation: "CSCI",
			},
			{
				Name:         "Williams Students Online",
				Abbreviation: "WSO",
			},
		},
	}
	d2 := models.Department{
		Name: "Economics",
	}
	assert.NoError(db.Create(&d1).Create(&d2).Error)

	// Get department 1
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/departments/%d", d1.ID), nil)
	assert.NoError(err)

	// Check if response has department 1's details
	resp := GetDepartmentFromResp(assert, w)
	assert.Equal(d1.Name, resp.Name)
	assert.Len(resp.AreasOfStudy, 2)

	// Check if we got areas of study (in reverse order)
	assert.Equal(d1.AreasOfStudy[0].Name, resp.AreasOfStudy[0].Name)
	assert.Equal(d1.AreasOfStudy[0].Abbreviation, resp.AreasOfStudy[0].Abbreviation)
	assert.Equal(d1.AreasOfStudy[1].Name, resp.AreasOfStudy[1].Name)
	assert.Equal(d1.AreasOfStudy[1].Abbreviation, resp.AreasOfStudy[1].Abbreviation)

	/* Get nonexistent course id (expect failure) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/departments/%d", 42), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusNotFound, w.Code)
}

func TestController_ListDepartmentProfessors(t *testing.T) {
	// Set up the test environment
	assert, db, router := SetupFactrakTest(t)

	// Insert test user into db
	p1 := models.User{
		Type:   models.UserTypeProfessor,
		Name:   "Professor 1",
		UnixID: "p1",
	}
	p2 := models.User{
		Type:   models.UserTypeProfessor,
		Name:   "Professor 2",
		UnixID: "p2",
	}
	// Not at williams
	p3 := models.User{
		Type:       models.UserTypeProfessor,
		Name:       "Professor 3",
		UnixID:     "p3",
		AtWilliams: lib.BoolToPtr(false),
	}
	// Other dept
	p4 := models.User{
		Type:   models.UserTypeProfessor,
		Name:   "Professor 4",
		UnixID: "p4",
	}
	// Staff
	s1 := models.User{
		Type:   models.UserTypeStaff,
		Name:   "Staff 1",
		UnixID: "s1",
	}
	assert.NoError(db.Create(&p1).Create(&p2).Create(&p3).Create(&p4).Create(&s1).Error)

	d1 := models.Department{
		Name: "Computer Science",
		AreasOfStudy: []*models.AreaOfStudy{
			{
				Name:         "Computer Science",
				Abbreviation: "CSCI",
			},
		},
		Users: []*models.User{
			&p1,
			&p2,
			&p3,
			&s1,
		},
	}
	d2 := models.Department{
		Name: "Economics",
		Users: []*models.User{
			&p4,
		},
	}
	assert.NoError(db.Create(&d1).Create(&d2).Error)

	/* Get test dept 1 (expect success) */
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/departments/%d/professors", d1.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Check if response has all dept professors, in order they were created
	resp := GetUsersFromResp(assert, w)
	assert.NoError(EqualUserIDs([]models.User{p1, p2}, resp))

	/* Get bad course (expect failure) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/departments/%d/professors", 42), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusNotFound, w.Code)

	/* Get professors for test dept 2 (expect success) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/departments/%d/professors", d2.ID), nil)
	assert.NoError(err)

	// Check if response is valid and has the professor
	resp = GetUsersFromResp(assert, w)
	assert.NoError(EqualUserIDs([]models.User{p4}, resp))
}

func TestController_ListDepartmentCourses(t *testing.T) {
	// Set up the test environment
	assert, db, router := SetupFactrakTest(t)

	// Need this to satisfy not null
	d1 := models.Department{
		Name: "Computer Science",
	}
	d2 := models.Department{
		Name: "Economics",
	}
	assert.NoError(db.Create(&d1).Create(&d2).Error)

	a1 := models.AreaOfStudy{
		Name:         "Computer Science",
		Abbreviation: "CSCI",
		Department:   &d1,
	}
	a2 := models.AreaOfStudy{
		Name:         "Williams Students Online",
		Abbreviation: "WSO",
		Department:   &d1,
	}
	a3 := models.AreaOfStudy{
		Name:         "Economics",
		Abbreviation: "ECON",
		Department:   &d2,
	}
	assert.NoError(db.Create(&a1).Create(&a2).Create(&a3).Error)

	c1 := models.Course{
		Number:      "Course 1",
		AreaOfStudy: &a1,
	}
	c2 := models.Course{
		Number:      "Course 2",
		AreaOfStudy: &a1,
	}
	c3 := models.Course{
		Number:      "Course 3",
		AreaOfStudy: &a2,
	}
	c4 := models.Course{
		Number:      "Course 4",
		AreaOfStudy: &a3,
	}
	assert.NoError(db.Create(&c1).Create(&c2).Create(&c3).Create(&c4).Error)

	/* Get courses for department 1 (expect success) */
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/departments/%d/courses", d1.ID), nil)
	assert.NoError(err)

	// Check if response is valid and has all dept courses
	resp := GetCoursesFromResp(assert, w)
	assert.NoError(EqualCourseIDs([]models.Course{c1, c2, c3}, resp))

	/* Get courses for nonexistent department (expect failure) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/departments/%d/courses", 42), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusNotFound, w.Code)

	/* Get courses for department 2 (expect success) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/departments/%d/courses", d2.ID), nil)
	assert.NoError(err)

	// Check if response is valid and has all dept courses
	resp = GetCoursesFromResp(assert, w)
	assert.NoError(EqualCourseIDs([]models.Course{c4}, resp))
}
