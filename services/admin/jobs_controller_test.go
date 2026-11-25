package admin_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/jobs/jobctl"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	testify "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type jobInfoResp struct {
	Name     string `json:"name"`
	Disabled bool   `json:"disabled"`
}

const testJob = "library-hours-update"

func listJobsDisabled(t *testing.T, env *utils.TestEnv) map[string]bool {
	t.Helper()
	w, err := utils.DoHTTPReq(env.Router, http.MethodGet, "/jobs", nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, w.Code)
	var resp utils.APITestResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	var jobs []jobInfoResp
	require.NoError(t, json.Unmarshal(resp.Data, &jobs))
	out := map[string]bool{}
	for _, j := range jobs {
		out[j.Name] = j.Disabled
	}
	return out
}

func setupAdminJobs(t *testing.T) (*testify.Assertions, *utils.TestEnv) {
	t.Helper()
	utils.RequireRedis(t, jobctl.ConfigureRedisForTest())
	return setupAdminTasks(t)
}

func TestUpdateJobTogglesDisabled(t *testing.T) {
	assert, env := setupAdminJobs(t)

	assert.False(listJobsDisabled(t, env)[testJob])

	w, err := utils.DoHTTPReq(env.Router, http.MethodPatch, "/jobs/"+testJob, strings.NewReader(`{"disabled":true}`))
	require.NoError(t, err)
	var info jobInfoResp
	require.NoError(t, json.Unmarshal(utils.GetGoodResp(assert, w).Data, &info))
	assert.Equal(testJob, info.Name)
	assert.True(info.Disabled)
	assert.True(listJobsDisabled(t, env)[testJob])

	w, err = utils.DoHTTPReq(env.Router, http.MethodPost, "/jobs/"+testJob+"/runs", strings.NewReader(`{}`))
	require.NoError(t, err)
	utils.CheckRespError(assert, w, lib.ErrorJobDisabled)

	w, err = utils.DoHTTPReq(env.Router, http.MethodPatch, "/jobs/"+testJob, strings.NewReader(`{"disabled":false}`))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(utils.GetGoodResp(assert, w).Data, &info))
	assert.False(info.Disabled)
	assert.False(listJobsDisabled(t, env)[testJob])
}

func TestUpdateJobRejectsBadInput(t *testing.T) {
	assert, env := setupAdminJobs(t)

	w, err := utils.DoHTTPReq(env.Router, http.MethodPatch, "/jobs/no-such-job", strings.NewReader(`{"disabled":true}`))
	require.NoError(t, err)
	utils.CheckRespError(assert, w, lib.ErrorRecordNotFound)

	w, err = utils.DoHTTPReq(env.Router, http.MethodPatch, "/jobs/"+testJob, strings.NewReader(`{}`))
	require.NoError(t, err)
	assert.Equal(http.StatusBadRequest, w.Code)
}
