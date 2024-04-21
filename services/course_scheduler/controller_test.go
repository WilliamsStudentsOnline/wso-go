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

func TestController_ListSelections(t *testing.T) {
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
					Abbreviation: "PSYC",
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

	// Check using userID filter
	w, err = utils.DoHTTPReq(router, http.MethodGet, "/get/?user_id=128", nil)
	assert.NoError(err)

	// Check that response matches database selections
	resp = GetSelectionsFromResp(assert, w)
	assert.Equal(courseSchedulerSelections[0].CourseID, resp[0].CourseID)

	// Check using userID and semester and year filter
	w, err = utils.DoHTTPReq(router, http.MethodGet, "/get/?user_id=400&semester=SPRING&year=2024", nil)
	assert.NoError(err)

	// Check that response matches database selections
	resp = GetSelectionsFromResp(assert, w)
	assert.Equal(courseSchedulerSelections[1].CourseID, resp[0].CourseID)

}

func TestController_AddSelection(t *testing.T) {
	assert, db, router := SetupCourseSchedulerSelectionTest(t, auth.ScopeCourseSchedulerAdmin)

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
			Abbreviation: "PSYC",
			Department: &Department{
				Name: "Psychology",
			},
		},
	}

	assert.NoError(db.Create(&user).Create(&course).Error)
	assert.NoError(db.Create(&courseSchedulerSelections[0]).Error)

	// Create test selection
	_, err := utils.DoHTTPReq(router, http.MethodPost, "/add/?userID=1&courseID=1&hidden=FALSE&semester=SPRING&year=2024", nil)
	assert.NoError(err)

	// Get test selections (WARN: USES PREVIOUS TEST)
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/get/", nil)
	assert.NoError(err)

	resp := GetSelectionsFromResp(assert, w)
	assert.Equal(2, len(resp))
	assert.Equal(uint(1), *resp[1].CourseID)

	// Check correct API responses for invalid queries
	// Missing userID
	w, err = utils.DoHTTPReq(router, http.MethodPost, "/add/", nil)
	assert.NoError(err)
	assert.Equal(lib.ErrorCourseSchedulerMissingUserID.HTTPCode, w.Code)

	// Missing courseID
	w, err = utils.DoHTTPReq(router, http.MethodPost, "/add/?userID=1", nil)
	assert.NoError(err)
	assert.Equal(lib.ErrorCourseSchedulerMissingCourseID.HTTPCode, w.Code)

	// Missing semester
	w, err = utils.DoHTTPReq(router, http.MethodPost, "/add/?userID=1&courseID=1&hidden=FALSE", nil)
	assert.NoError(err)
	assert.Equal(lib.ErrorCourseSchedulerMissingSemesterYear.HTTPCode, w.Code)

	// Missing year
	w, err = utils.DoHTTPReq(router, http.MethodPost, "/add/?userID=1&courseID=1&hidden=FALSE&semester=SPRING", nil)
	assert.NoError(err)
	assert.Equal(lib.ErrorCourseSchedulerMissingSemesterYear.HTTPCode, w.Code)

	// Missing courseID in the middle
	w, err = utils.DoHTTPReq(router, http.MethodPost, "/add/?userID=1&hidden=FALSE&semester=SPRING&year=2024", nil)
	assert.NoError(err)
	assert.Equal(lib.ErrorCourseSchedulerMissingCourseID.HTTPCode, w.Code)

	// Invalid userID (no user found with that UUID)
	w, err = utils.DoHTTPReq(router, http.MethodPost, "/add/?userID=999999&courseID=1&hidden=FALSE&semester=SPRING&year=2024", nil)
	assert.NoError(err)
	assert.Equal(lib.ErrorCourseSchedulerInvalidUserID.HTTPCode, w.Code)

	// Invalid courseID (no course found with that UUID)
	w, err = utils.DoHTTPReq(router, http.MethodPost, "/add/?userID=1&courseID=999999&hidden=FALSE&semester=SPRING&year=2024", nil)
	assert.NoError(err)
	assert.Equal(lib.ErrorCourseSchedulerInvalidCourseID.HTTPCode, w.Code)
}

func TestController_RemoveSelection(t *testing.T) {
	assert, db, router := SetupCourseSchedulerSelectionTest(t, auth.ScopeCourseSchedulerAdmin)

	var userID1 uint = 128
	var userID2 uint = 400
	var userID3 uint = 200
	var userID4 uint = 300
	var userID5 uint = 500
	var courseID1 uint = 256
	var courseID2 uint = 101
	var courseID3 uint = 247
	var courseID4 uint = 208
	var courseID5 uint = 151
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
					Abbreviation: "PSYC",
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
		{
			User: &User{
				Type:        UserTypeStudent,
				Name:        "Foo",
				UnixID:      "foo2",
				Title:       lib.StrToPtr("Crash Test Dummy"),
				ClassYear:   lib.IntToPtr(2021),
				Major:       lib.StrToPtr("Sociology"),
				SUBox:       lib.StrToPtr("1000"),
				Entry:       lib.StrToPtr("MD1"),
				DormVisible: lib.BoolToPtr(true),
				DormRoom: &DormRoom{
					Dorm: &Dorm{
						Neighborhood: &Neighborhood{
							Name: "Currier",
						},
						Name: "East",
					},
					Number: "201",
				},
				HomeVisible: lib.BoolToPtr(true),
				HomeTown:    lib.StrToPtr("Chicago"),
				HomeState:   lib.StrToPtr("Illinois"),
				HomeCountry: lib.StrToPtr("United States"),
			},
			UserID: &userID3,
			Course: &Course{
				Number: "247",
				AreaOfStudy: &AreaOfStudy{
					Name:         "American History",
					Abbreviation: "AMST",
					Department: &Department{
						Name: "American History",
					},
				},
			},
			CourseID: &courseID3,
			Hidden:   true,
			Semester: SemesterFall,
			Year:     &year,
		},
		{
			User: &User{
				Type:        UserTypeStudent,
				Name:        "Bar",
				UnixID:      "bar2",
				Title:       lib.StrToPtr("Software Engineer"),
				ClassYear:   lib.IntToPtr(2024),
				Major:       lib.StrToPtr("Mathematics"),
				SUBox:       lib.StrToPtr("3000"),
				Entry:       lib.StrToPtr("AP4"),
				DormVisible: lib.BoolToPtr(true),
				DormRoom: &DormRoom{
					Dorm: &Dorm{
						Neighborhood: &Neighborhood{
							Name: "Greylock",
						},
						Name: "Greylock",
					},
					Number: "204",
				},
				HomeVisible: lib.BoolToPtr(true),
				HomeTown:    lib.StrToPtr("Seattle"),
				HomeState:   lib.StrToPtr("Washington"),
				HomeCountry: lib.StrToPtr("United States"),
			},
			UserID: &userID4,
			Course: &Course{
				Number: "208",
				AreaOfStudy: &AreaOfStudy{
					Name:         "German",
					Abbreviation: "GERM",
					Department: &Department{
						Name: "German",
					},
				},
			},
			CourseID: &courseID4,
			Hidden:   false,
			Semester: SemesterSpring,
			Year:     &year,
		},
		{
			User: &User{
				Type:        UserTypeStudent,
				Name:        "Foo",
				UnixID:      "foo3",
				Title:       lib.StrToPtr("Goodrich Barista"),
				ClassYear:   lib.IntToPtr(2027),
				Major:       lib.StrToPtr("Physics"),
				SUBox:       lib.StrToPtr("4000"),
				Entry:       lib.StrToPtr("AP1"),
				DormVisible: lib.BoolToPtr(true),
				DormRoom: &DormRoom{
					Dorm: &Dorm{
						Neighborhood: &Neighborhood{
							Name: "Greylock",
						},
						Name: "Greylock",
					},
					Number: "304",
				},
				HomeVisible: lib.BoolToPtr(true),
				HomeTown:    lib.StrToPtr("Miami"),
				HomeState:   lib.StrToPtr("Florida"),
				HomeCountry: lib.StrToPtr("United States"),
			},
			UserID: &userID5,
			Course: &Course{
				Number: "151",
				AreaOfStudy: &AreaOfStudy{
					Name:         "Mathematics",
					Abbreviation: "MATH",
					Department: &Department{
						Name: "Mathematics",
					},
				},
			},
			CourseID: &courseID5,
			Hidden:   false,
			Semester: SemesterFall,
			Year:     &year,
		},
	}

	assert.NoError(db.Create(&courseSchedulerSelections[0]).Error)
	assert.NoError(db.Create(&courseSchedulerSelections[1]).Error)
	assert.NoError(db.Create(&courseSchedulerSelections[2]).Error)
	assert.NoError(db.Create(&courseSchedulerSelections[3]).Error)
	assert.NoError(db.Create(&courseSchedulerSelections[4]).Error)

	// Check correct non-deletion with valid query
	assert.Equal(5, DeletionCheck(assert, router, "/del/?courseID=99999"))                         // by courseID
	assert.Equal(5, DeletionCheck(assert, router, "/del/?semester=SPRING&year=2020"))              // by semester and year
	assert.Equal(5, DeletionCheck(assert, router, "/del/?userID=99999&courseID=99999"))            // by userID and courseID
	assert.Equal(5, DeletionCheck(assert, router, "/del/?userID=99999&semester=SPRING&year=2020")) // by userID and semester and year

	/*
		Entries in table
		user 	course 	semester	year
		128 	256 	Fall 		2024
		400 	101 	Spring		2024
		200 	247 	Fall 		2024
		300 	208 	Spring 		2024
		500		151		Fall		2024
	*/

	// Check correct deletion with valid queries
	// Using courseID
	assert.Equal(4, DeletionCheck(assert, router, "/del/?courseID=1"))

	// Using semester and year
	assert.Equal(2, DeletionCheck(assert, router, "/del/?semester=SPRING&year=2024"))

	// Using userID and courseID
	assert.Equal(1, DeletionCheck(assert, router, "/del/?userID=3&courseID=3"))

	// Using userID and semester and year
	assert.Equal(0, DeletionCheck(assert, router, "/del/?userID=5&semester=FALL&year=2024"))

	// Check correct API response for missing params
	w, err := utils.DoHTTPReq(router, http.MethodDelete, "/del/?userID=1", nil)
	assert.NoError(err)
	assert.Equal(lib.ErrorCourseSchedulerMissingDeletionQueryParam.HTTPCode, w.Code)
}

func DeletionCheck(assert *testify.Assertions, router *gin.Engine, delQuery string) int {
	// Perform deletion query
	_, err := utils.DoHTTPReq(router, http.MethodDelete, delQuery, nil)
	assert.NoError(err)

	// Check that length matches expected (WARN: USES PREVIOUS TEST)
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/get/", nil)
	assert.NoError(err)
	resp := GetSelectionsFromResp(assert, w)
	return len(resp)
}

func TestController_HideSelection(t *testing.T) {
	assert, db, router := SetupCourseSchedulerSelectionTest(t, auth.ScopeCourseSchedulerAdmin)

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
					Abbreviation: "PSYC",
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

	// Check using userID and courseID
	_, err := utils.DoHTTPReq(router, http.MethodPatch, "/hide/?userID=1&courseID=1&hidden=false", nil)
	assert.NoError(err)

	// WARN: uses previous test
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/get/?userID=1", nil)
	assert.NoError(err)
	resp := GetSelectionsFromResp(assert, w)
	assert.Equal(false, resp[0].Hidden)

	// Check using userID and semester and year
	_, err = utils.DoHTTPReq(router, http.MethodPatch, "/hide/?userID=2&semester=SPRING&year=2024&hidden=true", nil)
	assert.NoError(err)

	// WARN: uses previous test
	w, err = utils.DoHTTPReq(router, http.MethodGet, "/get/?userID=2", nil)
	assert.NoError(err)
	resp = GetSelectionsFromResp(assert, w)
	assert.Equal(true, resp[0].Hidden)

	// Check correct API response for missing userID
	w, err = utils.DoHTTPReq(router, http.MethodPatch, "/hide/", nil)
	assert.NoError(err)
	assert.Equal(lib.ErrorCourseSchedulerMissingUserID.HTTPCode, w.Code)

	// Check correct API response for missing params
	w, err = utils.DoHTTPReq(router, http.MethodPatch, "/hide/?userID=1", nil)
	assert.NoError(err)
	assert.Equal(lib.ErrorCourseSchedulerMissingHideQueryParam.HTTPCode, w.Code)

}
