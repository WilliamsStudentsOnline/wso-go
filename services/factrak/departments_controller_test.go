package factrak_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/factrak"
	testify "github.com/stretchr/testify/assert"
)

func TestController_ListDepartments(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg)

	d1 := models.Department{
		Name: "Computer Science",
	}
	d2 := models.Department{
		Name: "Economics",
	}
	assert.NoError(db.Create(&d1).Create(&d2).Error)

	// Get test dept
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/departments", nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp []models.Department
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct departments
	assert.Len(resp, 2)
	assert.Equal(d1.Name, resp[0].Name)
	assert.Equal(d2.Name, resp[1].Name)
}

func TestController_GetDepartment(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg)

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

	// Get test department
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/departments/%d", d1.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.Department
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct department
	assert.Equal(d1.Name, resp.Name)
	assert.Len(resp.AreasOfStudy, 2)

	// Check if we got areas of study (in reverse order)
	assert.Equal(d1.AreasOfStudy[0].Name, resp.AreasOfStudy[0].Name)
	assert.Equal(d1.AreasOfStudy[0].Abbreviation, resp.AreasOfStudy[0].Abbreviation)
	assert.Equal(d1.AreasOfStudy[1].Name, resp.AreasOfStudy[1].Name)
	assert.Equal(d1.AreasOfStudy[1].Abbreviation, resp.AreasOfStudy[1].Abbreviation)

	/* Get test bad course id (expect failure) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/departments/%d", 42), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusNotFound, w.Code)
}

func TestController_ListDepartmentProfessors(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg)

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

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp []models.User
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if is profs 1 and 2
	assert.Len(resp, 2)

	// It should be in order of created first to created last
	assert.Equal(p1.UnixID, resp[0].UnixID)
	assert.Equal(p2.UnixID, resp[1].UnixID)

	/* Get bad course (expect failure) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/departments/%d/professors", 42), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusNotFound, w.Code)

	/* Get test dept 2 (expect success) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/departments/%d/professors", d2.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	resp = []models.User{}
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if is survey 3
	assert.Len(resp, 1)
	assert.Equal(p4.UnixID, resp[0].UnixID)
}

func TestController_ListDepartmentCourses(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg)

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

	/* Get test prof 1 (expect success) */
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/departments/%d/courses", d1.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp []models.Course
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if is courses x3
	assert.Len(resp, 3)
	assert.Equal(c1.Number, resp[0].Number)
	assert.Equal(c2.Number, resp[1].Number)
	assert.Equal(c3.Number, resp[2].Number)

	/* Get test bad department (expect failure) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/departments/%d/courses", 42), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusNotFound, w.Code)

	/* Get test dept 2 (expect success) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/departments/%d/courses", d2.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	resp = []models.Course{}
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if is courses x1
	assert.Len(resp, 1)
	assert.Equal(c4.Number, resp[0].Number)
}
