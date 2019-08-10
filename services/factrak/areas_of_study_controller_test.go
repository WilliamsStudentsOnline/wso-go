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

func TestController_ListAreasOfStudy(t *testing.T) {
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

	// Get test area
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/areas-of-study", nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp []models.AreaOfStudy
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct area of study
	assert.Len(resp, 3)
	assert.Equal(a1.Name, resp[0].Name)
	assert.Equal(a1.Abbreviation, resp[0].Abbreviation)
	assert.Equal(a2.Name, resp[1].Name)
	assert.Equal(a2.Abbreviation, resp[1].Abbreviation)
	assert.Equal(a3.Name, resp[2].Name)
	assert.Equal(a3.Abbreviation, resp[2].Abbreviation)
}

func TestController_GetAreaOfStudy(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg)

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

	// Get test area of study
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/areas-of-study/%d", a1.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.AreaOfStudy
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct department
	assert.Equal(a1.Name, resp.Name)
	assert.Equal(a1.Department.Name, resp.Department.Name)

	/* Get test bad course id (expect failure) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/areas-of-study/%d", 42), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusNotFound, w.Code)
}

func TestController_ListAreaOfStudyProfessors(t *testing.T) {
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

	// Have to do this because at_williams is not a pointer.
	assert.NoError(db.Model(&p3).Update("at_williams", false).Error)

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

	/* Get test area 1 (expect success) */
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/areas-of-study/%d/professors", d1.AreasOfStudy[0].ID), nil)
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
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/areas-of-study/%d/professors", 42), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusNotFound, w.Code)

	/* Get test area 2 (expect success) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/areas-of-study/%d/professors", d2.AreasOfStudy[0].ID), nil)
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

func TestController_ListAreaOfStudyCourses(t *testing.T) {
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
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/areas-of-study/%d/courses", a1.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp []models.Course
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if is courses x2
	assert.Len(resp, 2)
	assert.Equal(c1.Number, resp[0].Number)
	assert.Equal(c2.Number, resp[1].Number)

	/* Get test bad department (expect failure) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/areas-of-study/%d/courses", 42), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusNotFound, w.Code)

	/* Get test area 2 (expect success) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/areas-of-study/%d/courses", a3.ID), nil)
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
