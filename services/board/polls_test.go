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

func pollParams(options []string, closesAt *time.Time) map[string]interface{} {
	return map[string]interface{}{
		"type":  models.BoardThreadTypeDiscussion,
		"title": "Poll",
		"body":  "vote",
		"poll":  map[string]interface{}{"options": options, "closesAt": closesAt},
	}
}

func TestBoard_PollVoting(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	author := models.User{Type: models.UserTypeStudent, UnixID: "author", Name: "Author"}
	voter := models.User{Type: models.UserTypeStudent, UnixID: "voter", Name: "Voter"}
	assert.NoError(db.Create(&author).Error)
	assert.NoError(db.Create(&voter).Error)
	authorRouter := boardRouter(t, db, author.ID, writerScopes()...)
	voterRouter := boardRouter(t, db, voter.ID, writerScopes()...)

	var created models.BoardThread
	code := doJSON(assert, authorRouter, http.MethodPost, "/threads", pollParams([]string{" Yes ", "No"}, nil), &created)
	assert.Equal(http.StatusCreated, code)
	assert.NotNil(created.Poll)
	assert.Len(created.Poll.Options, 2)
	assert.Equal("Yes", created.Poll.Options[0].Label)

	threadURL := fmt.Sprintf("/threads/%d", created.ID)
	voteURL := threadURL + "/vote"
	yes, no := created.Poll.Options[0].ID, created.Poll.Options[1].ID

	var got models.BoardThread
	doJSON(assert, voterRouter, http.MethodGet, threadURL, nil, &got)
	assert.Nil(got.Poll.TotalVotes)
	assert.Nil(got.Poll.Options[0].Votes)
	assert.Nil(got.Poll.MyVote)

	var poll models.BoardPoll
	code = doJSON(assert, voterRouter, http.MethodPut, voteURL, VoteParams{OptionID: yes}, &poll)
	assert.Equal(http.StatusOK, code)
	assert.Equal(yes, *poll.MyVote)
	assert.Equal(1, *poll.TotalVotes)
	assert.Equal(1, *poll.Options[0].Votes)
	assert.Equal(0, *poll.Options[1].Votes)

	poll = models.BoardPoll{}
	doJSON(assert, voterRouter, http.MethodPut, voteURL, VoteParams{OptionID: no}, &poll)
	assert.Equal(no, *poll.MyVote)
	assert.Equal(1, *poll.TotalVotes)
	assert.Equal(0, *poll.Options[0].Votes)
	assert.Equal(1, *poll.Options[1].Votes)

	got = models.BoardThread{}
	doJSON(assert, authorRouter, http.MethodGet, threadURL, nil, &got)
	assert.Nil(got.Poll.TotalVotes)

	doJSON(assert, authorRouter, http.MethodPut, voteURL, VoteParams{OptionID: yes}, nil)
	got = models.BoardThread{}
	doJSON(assert, authorRouter, http.MethodGet, threadURL, nil, &got)
	assert.Equal(2, *got.Poll.TotalVotes)

	var other models.BoardThread
	doJSON(assert, authorRouter, http.MethodPost, "/threads", pollParams([]string{"A", "B"}, nil), &other)
	body, _ := json.Marshal(VoteParams{OptionID: other.Poll.Options[0].ID})
	w, err := utils.DoHTTPReq(voterRouter, http.MethodPut, voteURL, bytes.NewReader(body))
	assert.NoError(err)
	utils.CheckRespError(assert, w, lib.ErrorBoardInvalidPollOption)

	var list []*models.BoardThread
	doJSON(assert, voterRouter, http.MethodGet, "/threads", nil, &list)
	assert.Len(list, 2)
	for _, th := range list {
		assert.NotNil(th.Poll)
	}
}

func TestBoard_PollClosed(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	author := models.User{Type: models.UserTypeStudent, UnixID: "author", Name: "Author"}
	assert.NoError(db.Create(&author).Error)
	authorRouter := boardRouter(t, db, author.ID, writerScopes()...)
	loggedOutRouter := boardRouter(t, db, 0, auth.ScopeBulletin)

	closesAt := time.Now().Add(time.Hour)
	var created models.BoardThread
	doJSON(assert, authorRouter, http.MethodPost, "/threads", pollParams([]string{"A", "B"}, &closesAt), &created)
	assert.False(created.Poll.Closed)

	doJSON(assert, authorRouter, http.MethodPut, fmt.Sprintf("/threads/%d/vote", created.ID), VoteParams{OptionID: created.Poll.Options[0].ID}, nil)
	assert.NoError(db.Model(&models.BoardPoll{}).Where("id = ?", created.Poll.ID).
		Update("closes_at", time.Now().Add(-time.Minute)).Error)

	body, _ := json.Marshal(VoteParams{OptionID: created.Poll.Options[1].ID})
	w, err := utils.DoHTTPReq(authorRouter, http.MethodPut, fmt.Sprintf("/threads/%d/vote", created.ID), bytes.NewReader(body))
	assert.NoError(err)
	utils.CheckRespError(assert, w, lib.ErrorBoardPollClosed)

	var got models.BoardThread
	doJSON(assert, loggedOutRouter, http.MethodGet, fmt.Sprintf("/threads/%d", created.ID), nil, &got)
	assert.True(got.Poll.Closed)
	assert.Equal(1, *got.Poll.TotalVotes)
	assert.Equal(1, *got.Poll.Options[0].Votes)

	var list []*models.BoardThread
	doJSON(assert, loggedOutRouter, http.MethodGet, "/threads", nil, &list)
	assert.Len(list, 1)
}

func TestBoard_PollValidation(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	author := models.User{Type: models.UserTypeStudent, UnixID: "author", Name: "Author"}
	assert.NoError(db.Create(&author).Error)
	authorRouter := boardRouter(t, db, author.ID, writerScopes()...)

	past := time.Now().Add(-time.Hour)
	cases := []map[string]interface{}{
		pollParams([]string{"only one"}, nil),
		pollParams([]string{"A", "  "}, nil),
		pollParams([]string{"1", "2", "3", "4", "5"}, nil),
		pollParams([]string{"A", "B"}, &past),
	}
	for _, params := range cases {
		body, _ := json.Marshal(params)
		w, err := utils.DoHTTPReq(authorRouter, http.MethodPost, "/threads", bytes.NewReader(body))
		assert.NoError(err)
		utils.CheckRespError(assert, w, lib.ErrorBoardInvalidPoll)
	}

	params := pollParams([]string{"A", "B"}, nil)
	params["type"] = models.BoardThreadTypeAnnouncement
	body, _ := json.Marshal(params)
	w, err := utils.DoHTTPReq(authorRouter, http.MethodPost, "/threads", bytes.NewReader(body))
	assert.NoError(err)
	utils.CheckRespError(assert, w, lib.ErrorBoardPollNotAllowed)
}

func TestBoard_DeleteThreadDeletesPoll(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	author := models.User{Type: models.UserTypeStudent, UnixID: "author", Name: "Author"}
	assert.NoError(db.Create(&author).Error)
	authorRouter := boardRouter(t, db, author.ID, writerScopes()...)

	var keep, gone models.BoardThread
	doJSON(assert, authorRouter, http.MethodPost, "/threads", pollParams([]string{"A", "B"}, nil), &keep)
	doJSON(assert, authorRouter, http.MethodPost, "/threads", pollParams([]string{"C", "D"}, nil), &gone)
	doJSON(assert, authorRouter, http.MethodPut, fmt.Sprintf("/threads/%d/vote", keep.ID), VoteParams{OptionID: keep.Poll.Options[0].ID}, nil)
	doJSON(assert, authorRouter, http.MethodPut, fmt.Sprintf("/threads/%d/vote", gone.ID), VoteParams{OptionID: gone.Poll.Options[0].ID}, nil)

	code := doJSON(assert, authorRouter, http.MethodDelete, fmt.Sprintf("/threads/%d", gone.ID), nil, nil)
	assert.Equal(http.StatusOK, code)

	var polls, options, votes int
	assert.NoError(db.Model(&models.BoardPoll{}).Count(&polls).Error)
	assert.NoError(db.Model(&models.BoardPollOption{}).Count(&options).Error)
	assert.NoError(db.Model(&models.BoardPollVote{}).Count(&votes).Error)
	assert.Equal(1, polls)
	assert.Equal(2, options)
	assert.Equal(1, votes)

	assert.NoError(db.Unscoped().Model(&models.BoardPollVote{}).Count(&votes).Error)
	assert.Equal(2, votes)

	body, _ := json.Marshal(VoteParams{OptionID: gone.Poll.Options[1].ID})
	w, err := utils.DoHTTPReq(authorRouter, http.MethodPut, fmt.Sprintf("/threads/%d/vote", gone.ID), bytes.NewReader(body))
	assert.NoError(err)
	utils.CheckRespError(assert, w, lib.ErrorBoardPollNotFound)
}
