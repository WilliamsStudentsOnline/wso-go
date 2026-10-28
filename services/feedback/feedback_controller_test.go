package feedback_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/feedback"
	testify "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func setupCreate(t *testing.T, userID uint) (*testify.Assertions, *utils.TestEnv) {
	t.Helper()
	assert := testify.New(t)
	env := utils.SetupTest(t, auth.ScopeWriteSelf)
	utils.AddUserContexts(env.Router, userID)
	SetupRouter(env.Router, env.DB, env.Cfg, zaptest.NewLogger(t).Sugar())
	return assert, env
}

func setupAdmin(t *testing.T) (*testify.Assertions, *utils.TestEnv) {
	t.Helper()
	assert := testify.New(t)
	env := utils.SetupTest(t, auth.ScopeAdminAll)
	utils.AddUserContexts(env.Router, 1)
	SetupAdminRouter(env.Router, env.DB, env.Cfg, zaptest.NewLogger(t).Sugar())
	return assert, env
}

func TestCreateFeedbackAnonymous(t *testing.T) {
	assert, env := setupCreate(t, 7)

	name := "Anon"
	url := "https://wso.williams.edu/foo"
	body, _ := json.Marshal(CreateFeedbackParams{
		PreferredName: &name,
		URL:           &url,
		Notes:         "the button is broken",
	})
	w, err := utils.DoHTTPReq(env.Router, http.MethodPost, "/", bytes.NewReader(body))
	assert.NoError(err)
	assert.Equal(http.StatusCreated, w.Code)

	resp := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(resp.Error)

	var created models.Feedback
	assert.NoError(json.Unmarshal(resp.Data, &created))
	assert.Equal("the button is broken", created.Notes)
	assert.Equal(models.FeedbackPriorityNone, created.Priority)
	assert.Nil(created.UserID)
	require.NotNil(t, created.PreferredName)
	assert.Equal("Anon", *created.PreferredName)
	require.NotNil(t, created.URL)
	assert.Equal(url, *created.URL)
}

func TestCreateFeedbackWithSelfUserID(t *testing.T) {
	assert, env := setupCreate(t, 7)

	uid := uint(7)
	body, _ := json.Marshal(CreateFeedbackParams{
		UserID: &uid,
		Notes:  "please contact me",
	})
	w, err := utils.DoHTTPReq(env.Router, http.MethodPost, "/", bytes.NewReader(body))
	assert.NoError(err)
	assert.Equal(http.StatusCreated, w.Code)

	resp := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	var created models.Feedback
	assert.NoError(json.Unmarshal(resp.Data, &created))
	require.NotNil(t, created.UserID)
	assert.Equal(uid, *created.UserID)
}

func TestCreateFeedbackRejectsNonSelfUserID(t *testing.T) {
	assert, env := setupCreate(t, 7)

	uid := uint(99)
	body, _ := json.Marshal(CreateFeedbackParams{
		UserID: &uid,
		Notes:  "spoofed",
	})
	w, err := utils.DoHTTPReq(env.Router, http.MethodPost, "/", bytes.NewReader(body))
	assert.NoError(err)
	utils.CheckRespError(assert, w, lib.ErrorMustBeSelf)
}

func TestCreateFeedbackRequiresNotes(t *testing.T) {
	assert, env := setupCreate(t, 7)

	body, _ := json.Marshal(map[string]interface{}{"url": "https://example.com"})
	w, err := utils.DoHTTPReq(env.Router, http.MethodPost, "/", bytes.NewReader(body))
	assert.NoError(err)
	assert.Equal(http.StatusBadRequest, w.Code)
}

func TestCreateFeedbackRequiresWriteSelf(t *testing.T) {
	assert := testify.New(t)
	env := utils.SetupTest(t) // no write:self
	utils.AddUserContexts(env.Router, 7)
	SetupRouter(env.Router, env.DB, env.Cfg, zaptest.NewLogger(t).Sugar())

	body, _ := json.Marshal(CreateFeedbackParams{Notes: "nope"})
	w, err := utils.DoHTTPReq(env.Router, http.MethodPost, "/", bytes.NewReader(body))
	assert.NoError(err)
	utils.CheckRespError(assert, w, lib.ErrorNoScopeAuthorization)
}

func TestAdminListGetUpdateDelete(t *testing.T) {
	assert, env := setupAdmin(t)

	uid := uint(3)
	seed := models.Feedback{
		UserID:   &uid,
		Notes:    "first",
		Priority: models.FeedbackPriorityNone,
	}
	require.NoError(t, env.DB.Create(&seed).Error)
	seed2 := models.Feedback{Notes: "second", Priority: models.FeedbackPriorityLow}
	require.NoError(t, env.DB.Create(&seed2).Error)

	w, err := utils.DoHTTPReq(env.Router, http.MethodGet, "/", nil)
	assert.NoError(err)
	resp := utils.GetGoodResp(assert, w)
	var listed []*models.Feedback
	assert.NoError(json.Unmarshal(resp.Data, &listed))
	assert.Len(listed, 2)
	assert.Equal(2, resp.PaginationTotal)

	w, err = utils.DoHTTPReq(env.Router, http.MethodGet, "/?priority=priority_low", nil)
	assert.NoError(err)
	resp = utils.GetGoodResp(assert, w)
	assert.NoError(json.Unmarshal(resp.Data, &listed))
	require.Len(t, listed, 1)
	assert.Equal("second", listed[0].Notes)

	w, err = utils.DoHTTPReq(env.Router, http.MethodGet, fmt.Sprintf("/%d", seed.ID), nil)
	assert.NoError(err)
	resp = utils.GetGoodResp(assert, w)
	var one models.Feedback
	assert.NoError(json.Unmarshal(resp.Data, &one))
	assert.Equal(seed.ID, one.ID)

	adminNotes := "looking into it"
	issue := "https://github.com/WilliamsStudentsOnline/wso-go/issues/1"
	priority := models.FeedbackPriorityHigh
	patch, _ := json.Marshal(UpdateFeedbackParams{
		AdminNotes: &adminNotes,
		Issue:      &issue,
		Priority:   &priority,
	})
	w, err = utils.DoHTTPReq(env.Router, http.MethodPatch, fmt.Sprintf("/%d", seed.ID), bytes.NewReader(patch))
	assert.NoError(err)
	resp = utils.GetGoodResp(assert, w)
	assert.NoError(json.Unmarshal(resp.Data, &one))
	assert.Equal(models.FeedbackPriorityHigh, one.Priority)
	require.NotNil(t, one.AdminNotes)
	assert.Equal(adminNotes, *one.AdminNotes)
	require.NotNil(t, one.Issue)
	assert.Equal(issue, *one.Issue)

	w, err = utils.DoHTTPReq(env.Router, http.MethodDelete, fmt.Sprintf("/%d", seed.ID), nil)
	assert.NoError(err)
	resp = utils.GetGoodResp(assert, w)
	var deleted map[string]uint
	assert.NoError(json.Unmarshal(resp.Data, &deleted))
	assert.Equal(seed.ID, deleted["deleted"])

	var count int
	assert.NoError(env.DB.Model(&models.Feedback{}).Where("id = ?", seed.ID).Count(&count).Error)
	assert.Zero(count)
}

func TestAdminUpdateNotFound(t *testing.T) {
	assert, env := setupAdmin(t)

	priority := models.FeedbackPriorityDismissed
	patch, _ := json.Marshal(UpdateFeedbackParams{Priority: &priority})
	w, err := utils.DoHTTPReq(env.Router, http.MethodPatch, "/99999", bytes.NewReader(patch))
	assert.NoError(err)
	assert.Equal(http.StatusNotFound, w.Code)
}

func TestAdminDeleteNotFound(t *testing.T) {
	assert, env := setupAdmin(t)

	w, err := utils.DoHTTPReq(env.Router, http.MethodDelete, "/99999", nil)
	assert.NoError(err)
	assert.Equal(http.StatusNotFound, w.Code)
}
