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
	var peoplesoftIdCS256 uint = 3044
	var peoplesoftIdPSYC101 uint = 3720

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

	profs := []User{
		{
			Type: UserTypeProfessor,
			Name: "CSCIProf",
			UnixID: "csprof",
		},
		{
			Type: UserTypeProfessor,
			Name: "PSYCProf",
			UnixID: "psyprof",
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
			AreaOfStudyAbbr: &courses[0].AreaOfStudy.Abbreviation,
			ProfessorID: &profs[0].ID,
			PeoplesoftID: &peoplesoftIdCS256,
			Semester: SemesterFall,
			Year:     &year,
		},
		{
			User:     &users[1],
			UserID:   &users[1].ID,
			Course:   &courses[1],
			CourseID: &courses[1].ID,
			Hidden:   false,
			AreaOfStudyAbbr: &courses[1].AreaOfStudy.Abbreviation,
			ProfessorID: &profs[1].ID,
			PeoplesoftID: &peoplesoftIdPSYC101,
			Semester: SemesterSpring,
			Year:     &year,
		},
	}

	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	assert.NoError(db.Create(&users[0]).Create(&users[1]).Error)
	assert.NoError(db.Create(&courses[0]).Create(&courses[1]).Error)
	assert.NoError(db.Create(&courseSchedulerSelections[0]).Create(&courseSchedulerSelections[1]).Error)

	router := utils.SetupRouter(auth.ScopeCourseSchedulerFull)
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

	// Check that get works as expected for admin requests
	assert = testify.New(t)
	db = utils.SetupServiceTest(assert)
	assert.NoError(db.Create(&users[0]).Create(&users[1]).Error)
	assert.NoError(db.Create(&courses[0]).Create(&courses[1]).Error)
	assert.NoError(db.Create(&courseSchedulerSelections[0]).Create(&courseSchedulerSelections[1]).Error)

	router = utils.SetupRouter(auth.ScopeCourseSchedulerAdmin)
	cfg = utils.SetupConfig()
	course_scheduler.SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	w, err = utils.DoHTTPReq(router, http.MethodGet, "/selections?userID=1", nil)
	assert.NoError(err)
	respData = utils.GetGoodResp(assert, w)
	assert.NoError(json.Unmarshal(respData.Data, &resp))
	assert.Equal(1, len(resp))

}

func TestController_AddSelection(t *testing.T) {
	var year uint = 2024
	var peoplesoftIdCS256 uint = 3044
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

	profs := []User{
		{
			Type: UserTypeProfessor,
			Name: "CSCIProf",
			UnixID: "csprof",
		},
		{
			Type: UserTypeProfessor,
			Name: "PSYCProf",
			UnixID: "psyprof",
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
			AreaOfStudyAbbr: &courses[0].AreaOfStudy.Abbreviation,
			ProfessorID: &profs[0].ID,
			PeoplesoftID: &peoplesoftIdCS256,
			Semester: SemesterFall,
			Year:     &year,
		},
	}

	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	assert.NoError(db.Create(&users[0]).Create(&users[1]).Error)
	assert.NoError(db.Create(&courses[0]).Create(&courses[1]).Error)
	assert.NoError(db.Create(&courseSchedulerSelections[0]).Error)

	router := utils.SetupRouter(auth.ScopeCourseSchedulerFull)
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

	// Check that post works as expected for admin requests
	assert = testify.New(t)
	db = utils.SetupServiceTest(assert)
	assert.NoError(db.Create(&users[0]).Create(&users[1]).Error)
	assert.NoError(db.Create(&courses[0]).Create(&courses[1]).Error)
	assert.NoError(db.Create(&courseSchedulerSelections[0]).Error)

	router = utils.SetupRouter(auth.ScopeCourseSchedulerAdmin)
	cfg = utils.SetupConfig()
	course_scheduler.SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	params = course_scheduler.CourseSchedulerSelectionCreateParams{
		CourseID: 2,
		Semester: models.SemesterSpring,
		Year:     year,
		Hidden:   false,
	}

	payload, err = json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(router, http.MethodPost, "/selections?userID=2", bytes.NewBuffer(payload))
	assert.NoError(err)
	assert.Equal(http.StatusCreated, w.Code)

}

func TestController_DeleteSelection(t *testing.T) {
	var year uint = 2024
	var peoplesoftIdCS256 uint = 3044
	var peoplesoftIdPSYC101 uint = 3720

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

	profs := []User{
		{
			Type: UserTypeProfessor,
			Name: "CSCIProf",
			UnixID: "csprof",
		},
		{
			Type: UserTypeProfessor,
			Name: "PSYCProf",
			UnixID: "psyprof",
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
			AreaOfStudyAbbr: &courses[0].AreaOfStudy.Abbreviation,
			ProfessorID: &profs[0].ID,
			PeoplesoftID: &peoplesoftIdCS256,
			Semester: SemesterFall,
			Year:     &year,
		},
		{
			User:     &users[1],
			UserID:   &users[1].ID,
			Course:   &courses[1],
			CourseID: &courses[1].ID,
			Hidden:   false,
			AreaOfStudyAbbr: &courses[1].AreaOfStudy.Abbreviation,
			ProfessorID: &profs[1].ID,
			PeoplesoftID: &peoplesoftIdPSYC101,
			Semester: SemesterSpring,
			Year:     &year,
		},
	}

	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	assert.NoError(db.Create(&users[0]).Create(&users[1]).Error)
	assert.NoError(db.Create(&courses[0]).Create(&courses[1]).Error)
	assert.NoError(db.Create(&courseSchedulerSelections[0]).Create(&courseSchedulerSelections[1]).Error)

	router := utils.SetupRouter(auth.ScopeCourseSchedulerFull)
	utils.AddUserContexts(router, users[0].ID)
	cfg := utils.SetupConfig()
	course_scheduler.SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	// Check correct deletion by ID (for a user's selection)
	w, err := utils.DoHTTPReq(router, http.MethodDelete, "/selections/1", nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Check failure to delete for unauthorized selection
	w, err = utils.DoHTTPReq(router, http.MethodDelete, "/selections/2", nil)
	assert.NoError(err)
	assert.Equal(http.StatusBadRequest, w.Code)

	// Check if admins can delete any selection
	assert = testify.New(t)
	db = utils.SetupServiceTest(assert)
	assert.NoError(db.Create(&users[0]).Create(&users[1]).Error)
	assert.NoError(db.Create(&courses[0]).Create(&courses[1]).Error)
	assert.NoError(db.Create(&courseSchedulerSelections[0]).Create(&courseSchedulerSelections[1]).Error)

	router = utils.SetupRouter(auth.ScopeCourseSchedulerAdmin)
	cfg = utils.SetupConfig()
	course_scheduler.SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	w, err = utils.DoHTTPReq(router, http.MethodDelete, "/selections/2", nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

}

func TestController_HideSelection(t *testing.T) {
	var year uint = 2024
	var peoplesoftIdCS256 uint = 3044
	var peoplesoftIdPSYC101 uint = 3720

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

	profs := []User{
		{
			Type: UserTypeProfessor,
			Name: "CSCIProf",
			UnixID: "csprof",
		},
		{
			Type: UserTypeProfessor,
			Name: "PSYCProf",
			UnixID: "psyprof",
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
			AreaOfStudyAbbr: &courses[0].AreaOfStudy.Abbreviation,
			ProfessorID: &profs[0].ID,
			PeoplesoftID: &peoplesoftIdCS256,
			Semester: SemesterFall,
			Year:     &year,
		},
		{
			User:     &users[1],
			UserID:   &users[1].ID,
			Course:   &courses[1],
			CourseID: &courses[1].ID,
			Hidden:   false,
			AreaOfStudyAbbr: &courses[1].AreaOfStudy.Abbreviation,
			ProfessorID: &profs[1].ID,
			PeoplesoftID: &peoplesoftIdPSYC101,
			Semester: SemesterSpring,
			Year:     &year,
		},
	}

	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	assert.NoError(db.Create(&users[0]).Create(&users[1]).Error)
	assert.NoError(db.Create(&courses[0]).Create(&courses[1]).Error)
	assert.NoError(db.Create(&courseSchedulerSelections[0]).Create(&courseSchedulerSelections[1]).Error)

	router := utils.SetupRouter(auth.ScopeCourseSchedulerFull)
	utils.AddUserContexts(router, users[0].ID)
	cfg := utils.SetupConfig()
	course_scheduler.SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	hidden := false
	params := course_scheduler.CourseSchedulerSelectionUpdateParams{
		Hidden: &hidden,
	}
	payload, err := json.Marshal(&params)
	assert.NoError(err)

	// Check correct update by ID
	w, err := utils.DoHTTPReq(router, http.MethodPatch, "/selections/1", bytes.NewBuffer(payload))
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Check failure to update for unauthorized selection
	w, err = utils.DoHTTPReq(router, http.MethodPatch, "/selections/2", bytes.NewBuffer(payload))
	assert.NoError(err)
	assert.Equal(http.StatusBadRequest, w.Code)

	// Check if admins can update any selection
	assert = testify.New(t)
	db = utils.SetupServiceTest(assert)
	assert.NoError(db.Create(&users[0]).Create(&users[1]).Error)
	assert.NoError(db.Create(&courses[0]).Create(&courses[1]).Error)
	assert.NoError(db.Create(&courseSchedulerSelections[0]).Create(&courseSchedulerSelections[1]).Error)

	router = utils.SetupRouter(auth.ScopeCourseSchedulerAdmin)
	cfg = utils.SetupConfig()
	course_scheduler.SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	w, err = utils.DoHTTPReq(router, http.MethodPatch, "/selections/2", bytes.NewBuffer(payload))
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

}
