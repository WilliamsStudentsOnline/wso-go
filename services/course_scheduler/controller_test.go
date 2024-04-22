package course_scheduler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services/course_scheduler"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestController_ListSelections(t *testing.T) {
	var year uint = 2024
	users := []User{
		{
			Type:   UserTypeStudent,
			Name:   "Foo",
			UnixID: "foo1",
		},
		{
			Type:   UserTypeStudent,
			Name:   "Bar",
			UnixID: "bar1",
		},
	}

	courses := []Course{
		{
			Number: "256",
			AreaOfStudy: &AreaOfStudy{
				Name:         "Computer Science",
				Abbreviation: "CSCI",
				Department: &Department{
					Name: "Computer Science",
				},
			},
		},
		{
			Number: "101",
			AreaOfStudy: &AreaOfStudy{
				Name:         "Psychology",
				Abbreviation: "PSYC",
				Department: &Department{
					Name: "Psychology",
				},
			},
		},
	}

	courseSchedulerSelections := []CourseSchedulerSelection{
		{
			User:     &users[0],
			UserID:   &users[0].ID,
			Course:   &courses[0],
			CourseID: &courses[0].ID,
			Hidden:   true,
			Semester: SemesterFall,
			Year:     &year,
		},
		{
			User:     &users[1],
			UserID:   &users[1].ID,
			Course:   &courses[1],
			CourseID: &courses[1].ID,
			Hidden:   false,
			Semester: SemesterSpring,
			Year:     &year,
		},
	}

	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	assert.NoError(db.Create(&users[0]).Create(&users[1]).Error)
	assert.NoError(db.Create(&courses[0]).Create(&courses[1]).Error)
	assert.NoError(db.Create(&courseSchedulerSelections[0]).Create(&courseSchedulerSelections[1]).Error)

	router := utils.SetupRouter(auth.ScopeUsers)
	utils.AddUserContexts(router, users[0].ID)
	cfg := utils.SetupConfig()
	course_scheduler.SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	// Get test selections
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/selections", nil)
	assert.NoError(err)

	// Check that response matches database selections
	respData := utils.GetGoodResp(assert, w)
	var resp []CourseSchedulerSelection
	assert.NoError(json.Unmarshal(respData.Data, &resp))
	assert.Equal(courseSchedulerSelections[0].CourseID, resp[0].CourseID)

	// Check using semester and year filter
	w, err = utils.DoHTTPReq(router, http.MethodGet, "/selections?semester=FALL&year=2024", nil)
	assert.NoError(err)

	// Check that response matches database selections
	respData = utils.GetGoodResp(assert, w)
	assert.NoError(json.Unmarshal(respData.Data, &resp))
	assert.Equal(courseSchedulerSelections[0].CourseID, resp[0].CourseID)

	// Check using semester and year filter for empty responses
	w, err = utils.DoHTTPReq(router, http.MethodGet, "/selections?semester=FALL&year=2025", nil)
	assert.NoError(err)

	// Check that response is empty
	respData = utils.GetGoodResp(assert, w)
	assert.NoError(json.Unmarshal(respData.Data, &resp))
	assert.Equal(0, len(resp))

}

func TestController_AddSelection(t *testing.T) {
	var year uint = 2024
	var b bool = true
	users := []User{
		{
			Type:       UserTypeStudent,
			Name:       "Foo",
			UnixID:     "foo1",
			Visible:    &b,
			AtWilliams: &b,
		},
		{
			Type:       UserTypeStudent,
			Name:       "Bar",
			UnixID:     "bar1",
			Visible:    &b,
			AtWilliams: &b,
		},
	}

	courses := []Course{
		{
			Number: "256",
			AreaOfStudy: &AreaOfStudy{
				Name:         "Computer Science",
				Abbreviation: "CSCI",
				Department: &Department{
					Name: "Computer Science",
				},
			},
		},
		{
			Number: "101",
			AreaOfStudy: &AreaOfStudy{
				Name:         "Psychology",
				Abbreviation: "PSYC",
				Department: &Department{
					Name: "Psychology",
				},
			},
		},
	}

	courseSchedulerSelections := []CourseSchedulerSelection{
		{
			User:     &users[0],
			UserID:   &users[0].ID,
			Course:   &courses[0],
			CourseID: &courses[0].ID,
			Hidden:   true,
			Semester: SemesterFall,
			Year:     &year,
		},
	}

	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	assert.NoError(db.Create(&users[0]).Create(&users[1]).Error)
	assert.NoError(db.Create(&courses[0]).Create(&courses[1]).Error)
	assert.NoError(db.Create(&courseSchedulerSelections[0]).Error)

	router := utils.SetupRouter(auth.ScopeUsers)
	utils.AddUserContexts(router, users[1].ID)
	cfg := utils.SetupConfig()
	course_scheduler.SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	// Create test selection
	params := course_scheduler.CourseSchedulerSelectionCreateParams{
		Semester: models.SemesterSpring,
		Year:     year,
		CourseID: courses[1].ID,
		Hidden:   false,
	}
	payload, err := json.Marshal(&params)
	assert.NoError(err)
	w, err := utils.DoHTTPReq(router, http.MethodPost, "/selections", bytes.NewBuffer(payload))
	assert.NoError(err)
	assert.Equal(http.StatusCreated, w.Code)

	// Get test selections (WARN: USES PREVIOUS TEST)
	w, err = utils.DoHTTPReq(router, http.MethodGet, "/selections", nil)
	assert.NoError(err)

	respData := utils.GetGoodResp(assert, w)
	var resp []CourseSchedulerSelection
	assert.NoError(json.Unmarshal(respData.Data, &resp))
	assert.Equal(1, len(resp))
	assert.Equal(courses[1].ID, *resp[0].CourseID)

	// Check correct API responses for invalid queries
	// Missing courseID
	params = course_scheduler.CourseSchedulerSelectionCreateParams{
		Semester: models.SemesterSpring,
		Year:     year,
		Hidden:   false,
	}
	payload, err = json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(router, http.MethodPost, "/selections", bytes.NewBuffer(payload))
	assert.NoError(err)
	assert.Equal(http.StatusBadRequest, w.Code)

	// Missing semester
	params = course_scheduler.CourseSchedulerSelectionCreateParams{
		CourseID: courses[1].ID,
		Year:     year,
		Hidden:   false,
	}
	payload, err = json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(router, http.MethodPost, "/selections", bytes.NewBuffer(payload))
	assert.NoError(err)
	assert.Equal(http.StatusBadRequest, w.Code)

	// Missing year
	params = course_scheduler.CourseSchedulerSelectionCreateParams{
		CourseID: courses[1].ID,
		Semester: models.SemesterSpring,
		Hidden:   false,
	}
	payload, err = json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(router, http.MethodPost, "/selections", bytes.NewBuffer(payload))
	assert.NoError(err)
	assert.Equal(http.StatusBadRequest, w.Code)

	// Invalid courseID (no course found with that UUID)
	params = course_scheduler.CourseSchedulerSelectionCreateParams{
		CourseID: 999,
		Semester: models.SemesterSpring,
		Year:     year,
		Hidden:   false,
	}
	payload, err = json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(router, http.MethodPost, "/selections", bytes.NewBuffer(payload))
	assert.NoError(err)
	assert.Equal(lib.ErrorCourseSchedulerInvalidCourseID.HTTPCode, w.Code)
}
