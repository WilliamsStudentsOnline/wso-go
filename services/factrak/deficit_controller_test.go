package factrak_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/factrak"
	testify "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestController_CreateDeficitExemption(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	student := models.User{
		Type: models.UserTypeStudent, Name: "Student 1", UnixID: "s1",
		AtWilliams: lib.BoolToPtr(true), OnCampusSemesters: 3,
	}
	require.NoError(t, db.Create(&student).Error)

	router := utils.SetupRouter(auth.ScopeFactrakLimited)
	utils.AddUserContexts(router, student.ID)
	SetupRouter(router, db, utils.SetupConfig(), zaptest.NewLogger(t).Sugar())

	post := func(reason string) *lib.APIError {
		body, _ := json.Marshal(DeficitExemptionCreateParams{Reason: reason})
		w, err := utils.DoHTTPReq(router, http.MethodPost, "/deficit/exemptions", bytes.NewReader(body))
		require.NoError(t, err)
		if w.Code == http.StatusCreated {
			resp := utils.GetHTTPDataResp(assert, w.Body.Bytes())
			var created map[string]interface{}
			assert.NoError(json.Unmarshal(resp.Data, &created))
			assert.NotContains(created, "status")
			return nil
		}
		resp := utils.GetHTTPDataResp(assert, w.Body.Bytes())
		require.NotNil(t, resp.Error)
		return &lib.APIError{Code: resp.Error.ErrorCode, HTTPCode: w.Code}
	}

	assert.Equal(lib.ErrorDeficitExemptionReasonInvalid.Code, post("   ").Code)
	assert.Equal(lib.ErrorDeficitExemptionReasonInvalid.Code, post(strings.Repeat("a", 1001)).Code)

	assert.Nil(post("  studying abroad  "))

	var exemption models.FactrakDeficitExemption
	require.NoError(t, db.Where("user_id = ?", student.ID).First(&exemption).Error)
	assert.Equal("studying abroad", exemption.Reason)
	assert.Equal(models.FactrakDeficitExemptionStatusPending, exemption.Status)

	var task models.ModeratorTask
	require.NoError(t, db.Where("task_type = ?", models.ModeratorTaskTypeFactrakDeficitExemption).First(&task).Error)
	assert.Equal(models.TaskReference(models.ModeratorTaskTypeFactrakDeficitExemption, exemption.ID), task.TaskReference)
	require.NotNil(t, task.RequestingUserID)
	assert.Equal(student.ID, *task.RequestingUserID)

	apiErr := post("another reason")
	require.NotNil(t, apiErr)
	assert.Equal(lib.ErrorDeficitExemptionExists.Code, apiErr.Code)
	assert.Equal(http.StatusConflict, apiErr.HTTPCode)
}
