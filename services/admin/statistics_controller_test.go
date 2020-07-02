package admin

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestController_GetStats(t *testing.T) {
	require := assert.New(t)              // This creates a testing assertion for us to use
	db := utils.SetupServiceTest(require) // This initializes our test DB

	// Insert test users into db
	u1 := models.User{
		BaseSchema: models.BaseSchema{
			ID: 101,
		},
		Type:      "student",
		Name:      "Test 1",
		UnixID:    "unix1",
		ClassYear: lib.IntToPtr(2022),
		Admin:     lib.BoolToPtr(true),
	} // Create a test user to put in the DB
	require.NoError(db.Create(&u1).Error) // This actually puts the user into the DB and ensures it worked

	// 2nd user is not a student
	u2 := models.User{
		Type:      "professor",
		Name:      "Test 2",
		UnixID:    "unix2",
		ClassYear: lib.IntToPtr(2022),
	}
	require.NoError(db.Create(&u2).Error)

	// 3rd user is student
	u3 := models.User{
		BaseSchema: models.BaseSchema{
			ID: 201,
		},
		Type:      "student",
		Name:      "Test 3",
		UnixID:    "unix3",
		ClassYear: lib.IntToPtr(2020),
		Admin:     lib.BoolToPtr(false),
	}
	require.NoError(db.Create(&u3).Error)

	// create survey made by user 1
	f1 := models.FactrakSurvey{
		UserID:               101,
		CourseID:             100,
		WouldRecommendCourse: lib.BoolToPtr(true),
		CourseWorkload:       lib.IntToPtr(3),
	}
	require.NoError(db.Create(&f1).Error)

	// create match between user 1 and 3
	m1 := models.EphmatchMatch{
		UserAID: 101,
		UserBID: 201,
	}
	require.NoError(db.Create(&m1).Error)

	// user 1 ephmatch profile
	p1 := models.EphmatchProfile{
		UserID: 101,
	}
	require.NoError(db.Create(&p1).Error)

	// user 2 ephmatch profile
	p2 := models.EphmatchProfile{
		UserID: 201,
	}
	require.NoError(db.Create(&p2).Error)

	// Initialize Routing
	router := utils.SetupRouter(auth.ScopeAdminAll) // Creates a test router with the scope admin-all. Thus, it will be authorized
	utils.AddUserContexts(router, u1.ID)            // Set the router to look like requests are coming from user 1.
	cfg := utils.SetupConfig()                      // Creates the testing server config
	logger := zap.S()                               // Creates the testing server logger
	SetupRouter(router, db, cfg, logger)            // Pass all these parameters into our SetupRouter function that will set up the controller

	w, err := utils.DoHTTPReq(router, http.MethodGet, "/get-stats", nil) // does a GET request to get-stats on our router
	require.NoError(err)                                                 // Ensure we don't fail

	// Decode response
	resp := utils.GetHTTPDataResp(require, w.Body.Bytes()) // Decodes the response into our API response structure (everything from the API is in this format)
	require.Nil(resp.Error)                                // Ensure no errors from response

	// Get the decoded response in respStat
	respStat := Stats{}                        // Initialize an empty stat struct to put the response data into
	err = json.Unmarshal(resp.Data, &respStat) // Put the stat data we got from the API into the Stats struct we just created
	require.NoError(err)

	// The tests:
	require.Equal(http.StatusOK, w.Code) // status code = 200

	// stats retrieved are the same as what we would expect
	require.Equal(3, respStat.NumberOfUsers)
	require.Equal(2, respStat.NumberOfUserStudents)
	require.Equal(1, respStat.NumberOfEphmatchMatches)
	require.Equal(1, respStat.NumberOfFactrakReviews)
	require.Equal(2, respStat.NumberOfEphmatchUsers)
	require.Equal(float64(100), respStat.PercentageUseEphmatch) // all students use ephmatch

}
