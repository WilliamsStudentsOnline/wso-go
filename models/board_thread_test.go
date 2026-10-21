package models_test

import (
	"testing"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestBoardThreadModel_CreateAndFilterResolved(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	log := zaptest.NewLogger(t).Sugar()
	tm := models.NewBoardThreadModel(db, log)

	u := models.User{Type: models.UserTypeStudent, UnixID: "u1", Name: "U"}
	assert.NoError(db.Create(&u).Error)

	open := &models.BoardThread{
		Type: models.BoardThreadTypeExchange, Title: "open", UserID: u.ID,
		ExUserName: u.Name, RepliesEnabled: true, LastActive: time.Now(),
	}
	assert.NoError(tm.CreateThread(open, "body", nil))

	resolved := &models.BoardThread{
		Type: models.BoardThreadTypeExchange, Title: "done", UserID: u.ID,
		ExUserName: u.Name, RepliesEnabled: true, LastActive: time.Now(),
		Resolved: lib.BoolToPtr(true),
	}
	assert.NoError(tm.CreateThread(resolved, "body2", nil))

	var threads []*models.BoardThread
	assert.NoError(tm.GetThreads(&threads, &models.GetBoardThreadsOptions{
		Resolved: lib.BoolToPtr(false),
	}))
	assert.Len(threads, 1)
	assert.Equal("open", threads[0].Title)

	threads = nil
	assert.NoError(tm.GetThreads(&threads, &models.GetBoardThreadsOptions{
		Resolved: lib.BoolToPtr(true),
	}))
	assert.Len(threads, 1)
	assert.Equal("done", threads[0].Title)
}

func TestBoardThreadModel_RepliesEnabledFalsePersists(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	log := zaptest.NewLogger(t).Sugar()
	tm := models.NewBoardThreadModel(db, log)

	u := models.User{Type: models.UserTypeStudent, UnixID: "u2", Name: "U2"}
	assert.NoError(db.Create(&u).Error)

	th := &models.BoardThread{
		Type: models.BoardThreadTypeDiscussion, Title: "x", UserID: u.ID,
		ExUserName: u.Name, RepliesEnabled: false, LastActive: time.Now(),
	}
	assert.NoError(tm.CreateThread(th, "body", nil))
	assert.False(th.RepliesEnabled)

	var loaded models.BoardThread
	assert.NoError(db.First(&loaded, th.ID).Error)
	assert.False(loaded.RepliesEnabled)
}
