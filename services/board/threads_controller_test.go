package board_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/board"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestBoard_CreateListGet(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeBulletin, auth.ScopeWriteSelf, auth.ScopeBulletinWrite, auth.ScopeUsers)
	cfg := utils.SetupConfig()

	u1 := models.User{Type: models.UserTypeStudent, UnixID: "u1", Name: "User One"}
	assert.NoError(db.Create(&u1).Error)
	utils.AddUserContexts(router, u1.ID)
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	body, _ := json.Marshal(CreateThreadParams{
		Type:  models.BoardThreadTypeDiscussion,
		Title: "Hello",
		Body:  "First post body",
	})
	w, err := utils.DoHTTPReq(router, http.MethodPost, "/threads", bytes.NewReader(body))
	assert.NoError(err)
	assert.Equal(http.StatusCreated, w.Code)

	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var created models.BoardThread
	assert.NoError(json.Unmarshal(respData.Data, &created))
	assert.Equal("Hello", created.Title)
	assert.Equal(models.BoardThreadTypeDiscussion, created.Type)
	assert.True(created.RepliesEnabled)
	assert.Equal("First post body", created.Body)
	assert.Equal(u1.Name, created.ExUserName)

	w, err = utils.DoHTTPReq(router, http.MethodGet, "/threads", nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)
	respData = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	var list struct {
		Threads []*models.BoardThread `json:"threads"`
	}
	assert.NoError(json.Unmarshal(respData.Data, &list))
	assert.Len(list.Threads, 1)

	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/threads/%d", created.ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)
	respData = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	var got models.BoardThread
	assert.NoError(json.Unmarshal(respData.Data, &got))
	assert.Equal(created.ID, got.ID)
	assert.NotEmpty(got.Posts)
	assert.True(got.Posts[0].IsOP)
}

func TestBoard_RideCreateRequiresMeta(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeBulletin, auth.ScopeWriteSelf, auth.ScopeBulletinWrite, auth.ScopeUsers)
	cfg := utils.SetupConfig()

	u1 := models.User{Type: models.UserTypeStudent, UnixID: "u1", Name: "User One"}
	assert.NoError(db.Create(&u1).Error)
	utils.AddUserContexts(router, u1.ID)
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	body, _ := json.Marshal(map[string]interface{}{
		"type":  models.BoardThreadTypeRide,
		"title": "Ride",
		"body":  "need a ride",
	})
	w, err := utils.DoHTTPReq(router, http.MethodPost, "/threads", bytes.NewReader(body))
	assert.NoError(err)
	utils.CheckRespError(assert, w, lib.ErrorBoardRideMetaRequired)

	body, _ = json.Marshal(map[string]interface{}{
		"type":  models.BoardThreadTypeRide,
		"title": "Ride",
		"body":  "offering seats",
		"ride": map[string]interface{}{
			"offeringRide": true,
			"source":       "Williamstown",
			"destination":  "Boston",
		},
	})
	w, err = utils.DoHTTPReq(router, http.MethodPost, "/threads", bytes.NewReader(body))
	assert.NoError(err)
	assert.Equal(http.StatusCreated, w.Code)
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	var created models.BoardThread
	assert.NoError(json.Unmarshal(respData.Data, &created))
	assert.NotNil(created.RideMeta)
	assert.True(created.RideMeta.OfferingRide)
	assert.Equal("Williamstown", created.RideMeta.Source)
}

func TestBoard_RepliesDisabledAndCascade(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeBulletin, auth.ScopeWriteSelf, auth.ScopeBulletinWrite, auth.ScopeUsers)
	cfg := utils.SetupConfig()

	u1 := models.User{Type: models.UserTypeStudent, UnixID: "u1", Name: "User One"}
	assert.NoError(db.Create(&u1).Error)
	utils.AddUserContexts(router, u1.ID)
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	body, _ := json.Marshal(map[string]interface{}{
		"type":           models.BoardThreadTypeDiscussion,
		"title":          "No replies",
		"body":           "op",
		"repliesEnabled": false,
	})
	w, err := utils.DoHTTPReq(router, http.MethodPost, "/threads", bytes.NewReader(body))
	assert.NoError(err)
	assert.Equal(http.StatusCreated, w.Code)
	var created models.BoardThread
	assert.NoError(json.Unmarshal(utils.GetHTTPDataResp(assert, w.Body.Bytes()).Data, &created))

	replyBody, _ := json.Marshal(CreateReplyParams{Content: "nope"})
	w, err = utils.DoHTTPReq(router, http.MethodPost, fmt.Sprintf("/threads/%d/replies", created.ID), bytes.NewReader(replyBody))
	assert.NoError(err)
	utils.CheckRespError(assert, w, lib.ErrorBoardRepliesDisabled)

	patch, _ := json.Marshal(map[string]interface{}{"repliesEnabled": true})
	w, err = utils.DoHTTPReq(router, http.MethodPatch, fmt.Sprintf("/threads/%d", created.ID), bytes.NewReader(patch))
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	w, err = utils.DoHTTPReq(router, http.MethodPost, fmt.Sprintf("/threads/%d/replies", created.ID), bytes.NewReader(replyBody))
	assert.NoError(err)
	assert.Equal(http.StatusCreated, w.Code)

	var op models.BoardPost
	assert.NoError(db.Where("thread_id = ? AND is_op = ?", created.ID, true).First(&op).Error)
	w, err = utils.DoHTTPReq(router, http.MethodDelete, fmt.Sprintf("/replies/%d", op.ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	var count int
	assert.NoError(db.Model(&models.BoardThread{}).Where("id = ?", created.ID).Count(&count).Error)
	assert.Equal(0, count)
}

func TestBoard_NoSelfFlagAndTypeImmutable(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeBulletin, auth.ScopeWriteSelf, auth.ScopeBulletinWrite, auth.ScopeUsers)
	cfg := utils.SetupConfig()

	u1 := models.User{Type: models.UserTypeStudent, UnixID: "u1", Name: "User One"}
	assert.NoError(db.Create(&u1).Error)
	utils.AddUserContexts(router, u1.ID)
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	body, _ := json.Marshal(map[string]interface{}{
		"type":  models.BoardThreadTypeJob,
		"title": "Job",
		"body":  "hiring",
	})
	w, err := utils.DoHTTPReq(router, http.MethodPost, "/threads", bytes.NewReader(body))
	assert.NoError(err)
	var created models.BoardThread
	assert.NoError(json.Unmarshal(utils.GetHTTPDataResp(assert, w.Body.Bytes()).Data, &created))

	w, err = utils.DoHTTPReq(router, http.MethodPost, fmt.Sprintf("/threads/%d/flag", created.ID), nil)
	assert.NoError(err)
	utils.CheckRespError(assert, w, lib.ErrorBoardCannotFlagSelf)

	patch, _ := json.Marshal(map[string]interface{}{"type": models.BoardThreadTypeExchange})
	w, err = utils.DoHTTPReq(router, http.MethodPatch, fmt.Sprintf("/threads/%d", created.ID), bytes.NewReader(patch))
	assert.NoError(err)
	utils.CheckRespError(assert, w, lib.ErrorBoardTypeImmutable)
}

func TestBoard_FeedExcludesFutureAndExpired(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeBulletin, auth.ScopeWriteSelf, auth.ScopeBulletinWrite, auth.ScopeUsers)
	cfg := utils.SetupConfig()

	u1 := models.User{Type: models.UserTypeStudent, UnixID: "u1", Name: "User One"}
	assert.NoError(db.Create(&u1).Error)

	now := time.Now()
	active := &models.BoardThread{
		Type: models.BoardThreadTypeAnnouncement, Title: "active", UserID: u1.ID,
		ExUserName: u1.Name, RepliesEnabled: true, LastActive: now,
		StartsAt: lib.TimeToPtr(now.Add(-time.Hour)),
		EndsAt:   lib.TimeToPtr(now.Add(time.Hour)),
	}
	future := &models.BoardThread{
		Type: models.BoardThreadTypeAnnouncement, Title: "future", UserID: u1.ID,
		ExUserName: u1.Name, RepliesEnabled: true, LastActive: now,
		StartsAt: lib.TimeToPtr(now.Add(24 * time.Hour)),
	}
	expired := &models.BoardThread{
		Type: models.BoardThreadTypeAnnouncement, Title: "expired", UserID: u1.ID,
		ExUserName: u1.Name, RepliesEnabled: true, LastActive: now,
		StartsAt: lib.TimeToPtr(now.Add(-48 * time.Hour)),
		EndsAt:   lib.TimeToPtr(now.Add(-time.Hour)),
	}
	assert.NoError(db.Create(active).Create(future).Create(expired).Error)
	for _, th := range []*models.BoardThread{active, future, expired} {
		assert.NoError(db.Create(&models.BoardPost{
			ThreadID: th.ID, UserID: u1.ID, ExUserName: u1.Name, Content: "x", IsOP: true,
		}).Error)
	}

	utils.AddUserContexts(router, u1.ID)
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	w, err := utils.DoHTTPReq(router, http.MethodGet, "/threads", nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)
	var list struct {
		Threads []*models.BoardThread `json:"threads"`
	}
	assert.NoError(json.Unmarshal(utils.GetHTTPDataResp(assert, w.Body.Bytes()).Data, &list))
	assert.Len(list.Threads, 1)
	assert.Equal("active", list.Threads[0].Title)
}
