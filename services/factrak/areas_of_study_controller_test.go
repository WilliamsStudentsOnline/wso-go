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

// Check that a slice of areas of study matches the expected slice, by comparing course IDs
func EqualAreaIDs(expected, resp []models.AreaOfStudy) error {
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
func GetAreasFromResp(assert *testify.Assertions, w *httptest.ResponseRecorder) []models.AreaOfStudy {
	respData := utils.GetGoodResp(assert, w)
	var resp []models.AreaOfStudy
	assert.NoError(json.Unmarshal(respData.Data, &resp))
	return resp
}

// Unmarshal a single department from an http response
func GetAreaFromResp(assert *testify.Assertions, w *httptest.ResponseRecorder) models.AreaOfStudy {
	respData := utils.GetGoodResp(assert, w)
	var resp models.AreaOfStudy
	assert.NoError(json.Unmarshal(respData.Data, &resp))
	return resp
}

func TestController_ListAreasOfStudy(t *testing.T) {
	// Set up the test environment
	assert, db, router := SetupFactrakTest(t)

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

	// Get all areas of study
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/areas-of-study", nil)
	assert.NoError(err)

	// Check if response has correct areas and details (sorted by name)
	resp := GetAreasFromResp(assert, w)
	assert.Len(resp, 3)
	assert.Equal(a1.Name, resp[0].Name)
	assert.Equal(a1.Abbreviation, resp[0].Abbreviation)
	assert.Equal(a3.Name, resp[1].Name)
	assert.Equal(a3.Abbreviation, resp[1].Abbreviation)
	assert.Equal(a2.Name, resp[2].Name)
	assert.Equal(a2.Abbreviation, resp[2].Abbreviation)
}

func TestController_GetAreaOfStudy(t *testing.T) {
	// Set up the test environment
	assert, db, router := SetupFactrakTest(t)

	a1 := models.AreaOfStudy{
		Name:         "Computer Science",
		Abbreviation: "CSCI",
		Department: &models.Department{
			Name: "Computer Science",
		},
	}
	a2 := models.AreaOfStudy{
		Name:         "Economics",
		Abbreviation: "ECON",
		Department: &models.Department{
			Name: "Economics",
		},
	}

	assert.NoError(db.Create(&a1).Create(&a2).Error)

	// Get area of study 1
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/areas-of-study/%d", a1.ID), nil)
	assert.NoError(err)

	// Check if response is valid and contains the right area of study
	resp := GetAreaFromResp(assert, w)
	assert.Equal(a1.Name, resp.Name)
	assert.Equal(a1.Department.Name, resp.Department.Name)

	/* Get nonexistent area of study (expect failure) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/areas-of-study/%d", 42), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusNotFound, w.Code)
}

func TestController_ListAreaOfStudyProfessors(t *testing.T) {
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
	// Other area/dept
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
		AreasOfStudy: []*models.AreaOfStudy{
			{
				Name:         "Economics",
				Abbreviation: "ECON",
			},
		},
		Users: []*models.User{
			&p4,
		},
	}
	assert.NoError(db.Create(&d1).Create(&d2).Error)

	/* Get professors for area 1 (expect success) */
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/areas-of-study/%d/professors", d1.AreasOfStudy[0].ID), nil)
	assert.NoError(err)

	// Check if response has correct professors, ordered from first to last created
	resp := GetUsersFromResp(assert, w)
	assert.Len(resp, 2)
	assert.NoError(EqualUserIDs([]models.User{p1, p2}, resp))

	/* Get professors for nonexistent area (expect failure) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/areas-of-study/%d/professors", 42), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusNotFound, w.Code)

	/* Get professors for area 2 (expect success) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/areas-of-study/%d/professors", d2.AreasOfStudy[0].ID), nil)
	assert.NoError(err)

	// Check if is survey 3
	resp = GetUsersFromResp(assert, w)
	assert.Len(resp, 1)
	assert.NoError(EqualUserIDs([]models.User{p4}, resp))
}

func TestController_ListAreaOfStudyCourses(t *testing.T) {
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

	/* Get courses in area of study 1 (expect success) */
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/areas-of-study/%d/courses", a1.ID), nil)
	assert.NoError(err)

	// Check the response has the correct courses, ordered from first to last created
	resp := GetCoursesFromResp(assert, w)
	assert.NoError(EqualCourseIDs([]models.Course{c1, c2}, resp))

	/* Get courses for nonexistent department (expect failure) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/areas-of-study/%d/courses", 42), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusNotFound, w.Code)

	/* Get courses for area 3 (expect success) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/areas-of-study/%d/courses", a3.ID), nil)
	assert.NoError(err)

	// Check if is courses x1
	resp = GetCoursesFromResp(assert, w)
	assert.NoError(EqualCourseIDs([]models.Course{c4}, resp))
}
