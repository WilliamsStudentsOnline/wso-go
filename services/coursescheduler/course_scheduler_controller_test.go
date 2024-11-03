package coursescheduler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/services/coursescheduler"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

// WARNING: THE TESTS IN THIS MODULE WILL FAIL LOCALLY IF REDIS IS NOT RUNNING UNAUTHED AT PORT 6379

func TestCourseScheduler_RedisSet(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeAdminAll)
	testUserID := uint(1)
	testUserIDString := strconv.FormatUint(uint64(testUserID), 10)
	utils.AddUserContexts(router, testUserID)
	cfg := utils.SetupConfig()
	coursescheduler.SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())
	utils.SetupClientForTest()

	params := coursescheduler.SelectionSetRequest{
		Courses: "CSCI;1046,CSCI;1056,STAT;1520,MATH;1392,SILP;1812,CSCI;3048,PHIL;3517,MATH;3547,PHIL;3522,SILP;3808",
	}
	payload, err := json.Marshal(params)
	assert.NoError(err)

	w, err := utils.DoHTTPReq(router, http.MethodPost, "/set?userID="+testUserIDString, bytes.NewBuffer(payload))
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)
}

func TestCourseScheduler_RedisGet(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeAdminAll)
	testUserID := uint(1)
	testUserIDString := strconv.FormatUint(uint64(testUserID), 10)
	utils.AddUserContexts(router, testUserID)
	cfg := utils.SetupConfig()
	coursescheduler.SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())
	utils.SetupClientForTest()

	// runs the set test to ensure that a proper value is stored
	params := coursescheduler.SelectionSetRequest{
		Courses: "CSCI;1046,CSCI;1056,STAT;1520,MATH;1392,SILP;1812,CSCI;3048,PHIL;3517,MATH;3547,PHIL;3522,SILP;3808",
	}
	payload, err := json.Marshal(params)
	assert.NoError(err)

	w, err := utils.DoHTTPReq(router, http.MethodPost, "/set?userID="+testUserIDString, bytes.NewBuffer(payload))
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// get test
	w, err = utils.DoHTTPReq(router, http.MethodGet, "/get?userID="+testUserIDString, nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	respData := utils.GetGoodResp(assert, w)
	var resp coursescheduler.CourseSelectionsString
	assert.NoError(json.Unmarshal(respData.Data, &resp))
	assert.Equal("CSCI;1046,CSCI;1056,STAT;1520,MATH;1392,SILP;1812,CSCI;3048,PHIL;3517,MATH;3547,PHIL;3522,SILP;3808", resp.Courses)
}
