package board_test

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
	. "github.com/WilliamsStudentsOnline/wso-go/services/board"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func boardRouter(t *testing.T, db *gorm.DB, userID uint, scopes ...string) *gin.Engine {
	router := utils.SetupRouter(scopes...)
	if userID != 0 {
		utils.AddUserContexts(router, userID)
	}
	SetupRouter(router, db, utils.SetupConfig(), zaptest.NewLogger(t).Sugar())
	return router
}

func writerScopes(extra ...string) []string {
	return append([]string{auth.ScopeBulletin, auth.ScopeWriteSelf, auth.ScopeBulletinWrite, auth.ScopeUsers}, extra...)
}

func doJSON(assert *testify.Assertions, router *gin.Engine, method, url string, body interface{}, out interface{}) int {
	var reader *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader(nil)
	}
	w, err := utils.DoHTTPReq(router, method, url, reader)
	assert.NoError(err)
	if out != nil {
		assert.NoError(json.Unmarshal(utils.GetHTTPDataResp(assert, w.Body.Bytes()).Data, out))
	}
	return w.Code
}

func TestBoard_AnonymousRedaction(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	author := models.User{Type: models.UserTypeStudent, UnixID: "author", Name: "Author"}
	other := models.User{Type: models.UserTypeStudent, UnixID: "other", Name: "Other"}
	admin := models.User{Type: models.UserTypeStudent, UnixID: "admin", Name: "Admin"}
	assert.NoError(db.Create(&author).Error)
	assert.NoError(db.Create(&other).Error)
	assert.NoError(db.Create(&admin).Error)

	authorRouter := boardRouter(t, db, author.ID, writerScopes()...)
	otherRouter := boardRouter(t, db, other.ID, writerScopes()...)
	adminRouter := boardRouter(t, db, admin.ID, writerScopes(auth.ScopeAdminAll)...)
	loggedOutRouter := boardRouter(t, db, 0, auth.ScopeBulletin)

	var created models.BoardThread
	code := doJSON(assert, authorRouter, http.MethodPost, "/threads", CreateThreadParams{
		Type:      models.BoardThreadTypeDiscussion,
		Title:     "Secret",
		Body:      "body",
		Anonymous: true,
	}, &created)
	assert.Equal(http.StatusCreated, code)
	assert.True(created.IsAnonymous())
	assert.Zero(created.UserID)
	assert.Empty(created.ExUserName)

	var reply models.BoardPost
	code = doJSON(assert, otherRouter, http.MethodPost, fmt.Sprintf("/threads/%d/replies", created.ID), CreateReplyParams{Content: "named reply"}, &reply)
	assert.Equal(http.StatusCreated, code)
	assert.Equal("Other", reply.ExUserName)
	assert.False(reply.ByOP)
	assert.True(reply.IsMine)

	url := fmt.Sprintf("/threads/%d", created.ID)

	var got models.BoardThread
	doJSON(assert, otherRouter, http.MethodGet, url, nil, &got)
	assert.Zero(got.UserID)
	assert.Nil(got.User)
	assert.Empty(got.ExUserName)
	assert.Len(got.Posts, 2)
	assert.Zero(got.Posts[0].UserID)
	assert.Nil(got.Posts[0].User)
	assert.Empty(got.Posts[0].ExUserName)
	assert.True(got.Posts[0].ByOP)
	assert.False(got.Posts[0].IsMine)
	assert.Equal(other.ID, got.Posts[1].UserID)
	assert.True(got.Posts[1].IsMine)

	got = models.BoardThread{}
	doJSON(assert, authorRouter, http.MethodGet, url, nil, &got)
	assert.Zero(got.UserID)
	assert.True(got.Posts[0].IsMine)
	assert.False(got.Posts[1].IsMine)

	got = models.BoardThread{}
	doJSON(assert, adminRouter, http.MethodGet, url, nil, &got)
	assert.Equal(author.ID, got.UserID)
	assert.Equal("Author", got.ExUserName)
	assert.Equal(author.ID, got.Posts[0].UserID)

	got = models.BoardThread{}
	doJSON(assert, loggedOutRouter, http.MethodGet, url, nil, &got)
	assert.Zero(got.Posts[1].UserID)
	assert.Empty(got.Posts[1].ExUserName)
}

func TestBoard_AnonymousOnlyDiscussions(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	author := models.User{Type: models.UserTypeStudent, UnixID: "author", Name: "Author"}
	other := models.User{Type: models.UserTypeStudent, UnixID: "other", Name: "Other"}
	assert.NoError(db.Create(&author).Error)
	assert.NoError(db.Create(&other).Error)
	authorRouter := boardRouter(t, db, author.ID, writerScopes()...)
	otherRouter := boardRouter(t, db, other.ID, writerScopes()...)

	body, _ := json.Marshal(CreateThreadParams{
		Type:      models.BoardThreadTypeAnnouncement,
		Title:     "Nope",
		Body:      "body",
		Anonymous: true,
	})
	w, err := utils.DoHTTPReq(authorRouter, http.MethodPost, "/threads", bytes.NewReader(body))
	assert.NoError(err)
	utils.CheckRespError(assert, w, lib.ErrorBoardAnonymousNotAllowed)

	var created models.BoardThread
	doJSON(assert, authorRouter, http.MethodPost, "/threads", CreateThreadParams{
		Type:  models.BoardThreadTypeAnnouncement,
		Title: "Named",
		Body:  "body",
	}, &created)

	body, _ = json.Marshal(CreateReplyParams{Content: "hi", Anonymous: true})
	w, err = utils.DoHTTPReq(otherRouter, http.MethodPost, fmt.Sprintf("/threads/%d/replies", created.ID), bytes.NewReader(body))
	assert.NoError(err)
	utils.CheckRespError(assert, w, lib.ErrorBoardAnonymousNotAllowed)
}

func TestBoard_OPRepliesMatchThreadAnonymity(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	author := models.User{Type: models.UserTypeStudent, UnixID: "author", Name: "Author"}
	assert.NoError(db.Create(&author).Error)
	authorRouter := boardRouter(t, db, author.ID, writerScopes()...)

	var anonThread, namedThread models.BoardThread
	doJSON(assert, authorRouter, http.MethodPost, "/threads", CreateThreadParams{
		Type: models.BoardThreadTypeDiscussion, Title: "Anon", Body: "body", Anonymous: true,
	}, &anonThread)
	doJSON(assert, authorRouter, http.MethodPost, "/threads", CreateThreadParams{
		Type: models.BoardThreadTypeDiscussion, Title: "Named", Body: "body",
	}, &namedThread)

	var reply models.BoardPost
	doJSON(assert, authorRouter, http.MethodPost, fmt.Sprintf("/threads/%d/replies", anonThread.ID), CreateReplyParams{Content: "unmask?"}, &reply)
	assert.True(reply.IsAnonymous())
	assert.True(reply.ByOP)
	assert.Zero(reply.UserID)

	reply = models.BoardPost{}
	doJSON(assert, authorRouter, http.MethodPost, fmt.Sprintf("/threads/%d/replies", namedThread.ID), CreateReplyParams{Content: "hide?", Anonymous: true}, &reply)
	assert.False(reply.IsAnonymous())
	assert.Equal(author.ID, reply.UserID)
}

func TestBoard_UserFilterHidesAnonymousThreads(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	author := models.User{Type: models.UserTypeStudent, UnixID: "author", Name: "Author"}
	other := models.User{Type: models.UserTypeStudent, UnixID: "other", Name: "Other"}
	admin := models.User{Type: models.UserTypeStudent, UnixID: "admin", Name: "Admin"}
	assert.NoError(db.Create(&author).Error)
	assert.NoError(db.Create(&other).Error)
	assert.NoError(db.Create(&admin).Error)
	authorRouter := boardRouter(t, db, author.ID, writerScopes()...)
	otherRouter := boardRouter(t, db, other.ID, writerScopes()...)
	adminRouter := boardRouter(t, db, admin.ID, writerScopes(auth.ScopeAdminAll)...)

	doJSON(assert, authorRouter, http.MethodPost, "/threads", CreateThreadParams{
		Type: models.BoardThreadTypeDiscussion, Title: "Anon", Body: "body", Anonymous: true,
	}, nil)
	doJSON(assert, authorRouter, http.MethodPost, "/threads", CreateThreadParams{
		Type: models.BoardThreadTypeDiscussion, Title: "Named", Body: "body",
	}, nil)

	url := fmt.Sprintf("/threads?userID=%d", author.ID)
	var list struct {
		Threads []*models.BoardThread `json:"threads"`
	}
	doJSON(assert, otherRouter, http.MethodGet, url, nil, &list)
	assert.Len(list.Threads, 1)
	assert.Equal("Named", list.Threads[0].Title)

	doJSON(assert, authorRouter, http.MethodGet, url, nil, &list)
	assert.Len(list.Threads, 2)

	doJSON(assert, adminRouter, http.MethodGet, url, nil, &list)
	assert.Len(list.Threads, 2)

	doJSON(assert, otherRouter, http.MethodGet, "/threads", nil, &list)
	assert.Len(list.Threads, 2)
}
