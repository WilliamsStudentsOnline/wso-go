package course_scheduler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	. "github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services/course_scheduler"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

// Quick setup for course scheduler tests, configuring the router and database
func SetupCourseSchedulerSelectionTest(t *testing.T, scope string) (*testify.Assertions, *gorm.DB, *gin.Engine) {
	env := utils.SetupTest(t, scope)
	course_scheduler.SetupRouter(env.Router, env.DB, env.Cfg, zaptest.NewLogger(t).Sugar())

	return env.Assert, env.DB, env.Router
}

// Unmarshal a slice of selections from an http response
func GetSelectionsFromResp(assert *testify.Assertions, w *httptest.ResponseRecorder) []CourseSchedulerSelection {
	respData := utils.GetGoodResp(assert, w)
	var resp []CourseSchedulerSelection
	assert.NoError(json.Unmarshal(respData.Data, &resp))
	return resp
}

// Unmarshal a single selection from an http response
func GetSelectionFromResp(assert *testify.Assertions, w *httptest.ResponseRecorder) CourseSchedulerSelection {
	respData := utils.GetGoodResp(assert, w)
	var resp CourseSchedulerSelection
	assert.NoError(json.Unmarshal(respData.Data, &resp))
	return resp
}

func TestController_ListCourseSchedulerSelections(t *testing.T) {
	assert, db, router := SetupCourseSchedulerSelectionTest(t, auth.ScopeUsers)

	var userID1 uint = 128
	var userID2 uint = 400
	var courseID1 uint = 256
	var courseID2 uint = 101
	var year uint = 2024
	courseSchedulerSelections := []CourseSchedulerSelection{
		{
			User: &User{
				Type:        UserTypeStudent,
				Name:        "Foo",
				UnixID:      "foo1",
				Title:       lib.StrToPtr("Research Assistant"),
				ClassYear:   lib.IntToPtr(2022),
				Major:       lib.StrToPtr("Computer Science"),
				SUBox:       lib.StrToPtr("2885"),
				Entry:       lib.StrToPtr("AP3"),
				DormVisible: lib.BoolToPtr(true),
				DormRoom: &DormRoom{
					Dorm: &Dorm{
						Neighborhood: &Neighborhood{
							Name: "Currier",
						},
						Name: "East",
					},
					Number: "103",
				},
				HomeVisible: lib.BoolToPtr(true),
				HomeTown:    lib.StrToPtr("Palo Alto"),
				HomeState:   lib.StrToPtr("California"),
				HomeCountry: lib.StrToPtr("United States"),
			},
			UserID: &userID1,
			Course: &Course{
				Number: "256",
				AreaOfStudy: &AreaOfStudy{
					Name:         "Computer Science",
					Abbreviation: "CSCI",
					Department: &Department{
						Name: "Computer Science",
					},
				},
			},
			CourseID: &courseID1,
			Hidden:   true,
			Semester: SemesterFall,
			Year:     &year,
		},
		{
			User: &User{
				Type:        UserTypeStudent,
				Name:        "Bar",
				UnixID:      "bar1",
				Title:       lib.StrToPtr("Research Assistant"),
				ClassYear:   lib.IntToPtr(2021),
				Major:       lib.StrToPtr("Psychology"),
				SUBox:       lib.StrToPtr("1024"),
				Entry:       lib.StrToPtr("MD3"),
				DormVisible: lib.BoolToPtr(true),
				DormRoom: &DormRoom{
					Dorm: &Dorm{
						Neighborhood: &Neighborhood{
							Name: "Greylock",
						},
						Name: "Greylock",
					},
					Number: "104",
				},
				HomeVisible: lib.BoolToPtr(true),
				HomeTown:    lib.StrToPtr("New York"),
				HomeState:   lib.StrToPtr("New York"),
				HomeCountry: lib.StrToPtr("United States"),
			},
			UserID: &userID2,
			Course: &Course{
				Number: "101",
				AreaOfStudy: &AreaOfStudy{
					Name:         "Psychology",
					Abbreviation: "101",
					Department: &Department{
						Name: "Psychology",
					},
				},
			},
			CourseID: &courseID2,
			Hidden:   false,
			Semester: SemesterSpring,
			Year:     &year,
		},
	}

	assert.NoError(db.Create(&courseSchedulerSelections[0]).Create(&courseSchedulerSelections[1]).Error)

	// Get test selections
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/get/", nil)
	assert.NoError(err)

	// Check that response matches database selections
	resp := GetSelectionsFromResp(assert, w)
	assert.Equal(courseSchedulerSelections[0].CourseID, resp[0].CourseID)
	assert.Equal(courseSchedulerSelections[1].CourseID, resp[1].CourseID)

	// Check using filter
	w, err = utils.DoHTTPReq(router, http.MethodGet, "/get/?user_id=128", nil)
	assert.NoError(err)

	// Check that response matches database selections
	resp = GetSelectionsFromResp(assert, w)
	assert.Equal(courseSchedulerSelections[0].CourseID, resp[0].CourseID)
}

func TestController_AddCourse(t *testing.T) {
	assert, db, router := SetupCourseSchedulerSelectionTest(t, auth.ScopeCourseSchedulerFull)

	var userID1 uint = 128
	var courseID1 uint = 256
	var year uint = 2024
	courseSchedulerSelections := []CourseSchedulerSelection{
		{
			User: &User{
				Type:        UserTypeStudent,
				Name:        "Foo",
				UnixID:      "foo1",
				Title:       lib.StrToPtr("Research Assistant"),
				ClassYear:   lib.IntToPtr(2022),
				Major:       lib.StrToPtr("Computer Science"),
				SUBox:       lib.StrToPtr("2885"),
				Entry:       lib.StrToPtr("AP3"),
				DormVisible: lib.BoolToPtr(true),
				DormRoom: &DormRoom{
					Dorm: &Dorm{
						Neighborhood: &Neighborhood{
							Name: "Currier",
						},
						Name: "East",
					},
					Number: "103",
				},
				HomeVisible: lib.BoolToPtr(true),
				HomeTown:    lib.StrToPtr("Palo Alto"),
				HomeState:   lib.StrToPtr("California"),
				HomeCountry: lib.StrToPtr("United States"),
			},
			UserID: &userID1,
			Course: &Course{
				Number: "256",
				AreaOfStudy: &AreaOfStudy{
					Name:         "Computer Science",
					Abbreviation: "CSCI",
					Department: &Department{
						Name: "Computer Science",
					},
				},
			},
			CourseID: &courseID1,
			Hidden:   true,
			Semester: SemesterFall,
			Year:     &year,
		},
	}

	// //
	user := User{
		Type:        UserTypeStudent,
		Name:        "Bar",
		UnixID:      "bar1",
		Title:       lib.StrToPtr("Research Assistant"),
		ClassYear:   lib.IntToPtr(2021),
		Major:       lib.StrToPtr("Psychology"),
		SUBox:       lib.StrToPtr("1024"),
		Entry:       lib.StrToPtr("MD3"),
		DormVisible: lib.BoolToPtr(true),
		DormRoom: &DormRoom{
			Dorm: &Dorm{
				Neighborhood: &Neighborhood{
					Name: "Greylock",
				},
				Name: "Greylock",
			},
			Number: "104",
		},
		HomeVisible: lib.BoolToPtr(true),
		HomeTown:    lib.StrToPtr("New York"),
		HomeState:   lib.StrToPtr("New York"),
		HomeCountry: lib.StrToPtr("United States"),
	}

	course := Course{
		Number: "101",
		AreaOfStudy: &AreaOfStudy{
			Name:         "Psychology",
			Abbreviation: "101",
			Department: &Department{
				Name: "Psychology",
			},
		},
	}

	assert.NoError(db.Create(&user).Create(&course).Error)
	assert.NoError(db.Create(&courseSchedulerSelections[0]).Error)
	// //

	// Create test selection
	// THIS IS FAILING BECAUSE IT DOES NOT GRAB THE USER AND COURSE OBJECTS CORRECTLY USING ID VALUES
	_, err := utils.DoHTTPReq(router, http.MethodPost, "/add/?userID=1&courseID=1&hidden=FALSE&semester=SPRING&year=2024", nil)
	assert.NoError(err)

	// Get test selections (WARN: USES PREVIOUS TEST)
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/get/", nil)
	assert.NoError(err)

	resp := GetSelectionsFromResp(assert, w)
	assert.Equal(2, len(resp))
	assert.Equal(uint(1), resp[1].CourseID)

	// Check that we can't mess up adding a course
	_, err = utils.DoHTTPReq(router, http.MethodPost, "/add/?hidden=TRUE", nil)
	assert.Error(err)
}
