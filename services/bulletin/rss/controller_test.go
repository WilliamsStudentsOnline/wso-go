package rss_test

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/bulletin/rss"
	"github.com/gin-gonic/gin"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestRSS_DiscussionHidesAnonymousAuthors(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	u := models.User{Type: models.UserTypeStudent, UnixID: "author", Name: "Zane Rogahn"}
	assert.NoError(db.Create(&u).Error)

	tm := models.NewBoardThreadModel(db, zaptest.NewLogger(t).Sugar())
	for _, anon := range []bool{true, false} {
		title := "named"
		if anon {
			title = "anon"
		}
		assert.NoError(tm.CreateThread(&models.BoardThread{
			Type: models.BoardThreadTypeDiscussion, Title: title, UserID: u.ID, ExUserName: u.Name,
			RepliesEnabled: lib.TruePtr(), LastActive: time.Now(), Anonymous: &anon,
		}, "body", nil))
	}

	router := gin.New()
	SetupRouter(router, db, utils.SetupConfig(), zaptest.NewLogger(t).Sugar())
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/discussion", nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)
	assert.Contains(w.Body.String(), "Anonymous")
	assert.Equal(1, strings.Count(w.Body.String(), "Zane R"))
}
