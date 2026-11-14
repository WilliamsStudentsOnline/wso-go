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
		ExUserName: u.Name, RepliesEnabled: lib.TruePtr(), LastActive: time.Now(),
	}
	assert.NoError(tm.CreateThread(open, "body", nil))

	resolved := &models.BoardThread{
		Type: models.BoardThreadTypeExchange, Title: "done", UserID: u.ID,
		ExUserName: u.Name, RepliesEnabled: lib.TruePtr(), LastActive: time.Now(),
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

func TestBoardThreadModel_ListShowsUpcomingRides(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	log := zaptest.NewLogger(t).Sugar()
	tm := models.NewBoardThreadModel(db, log)

	u := models.User{Type: models.UserTypeStudent, UnixID: "u-ride", Name: "R"}
	assert.NoError(db.Create(&u).Error)

	future := time.Now().Add(48 * time.Hour)
	past := time.Now().Add(-48 * time.Hour)

	upcoming := &models.BoardThread{
		Type: models.BoardThreadTypeRide, Title: "upcoming", UserID: u.ID,
		ExUserName: u.Name, RepliesEnabled: lib.TruePtr(), LastActive: time.Now(),
		StartsAt: &future,
	}
	assert.NoError(tm.CreateThread(upcoming, "body", &models.BoardRideInformation{
		OfferingRide: lib.BoolToPtr(true), Source: "A", Destination: "B",
	}))

	gone := &models.BoardThread{
		Type: models.BoardThreadTypeRide, Title: "past", UserID: u.ID,
		ExUserName: u.Name, RepliesEnabled: lib.TruePtr(), LastActive: time.Now(),
		StartsAt: &past,
	}
	assert.NoError(tm.CreateThread(gone, "body", &models.BoardRideInformation{
		OfferingRide: lib.BoolToPtr(false), Source: "C", Destination: "D",
	}))

	scheduled := &models.BoardThread{
		Type: models.BoardThreadTypeAnnouncement, Title: "later", UserID: u.ID,
		ExUserName: u.Name, RepliesEnabled: lib.TruePtr(), LastActive: time.Now(),
		StartsAt: &future,
	}
	assert.NoError(tm.CreateThread(scheduled, "body", nil))

	var threads []*models.BoardThread
	assert.NoError(tm.GetThreads(&threads, &models.GetBoardThreadsOptions{}))
	titles := make([]string, 0, len(threads))
	for _, th := range threads {
		titles = append(titles, th.Title)
	}
	assert.Contains(titles, "upcoming")
	assert.NotContains(titles, "past")
	assert.NotContains(titles, "later")
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
		ExUserName: u.Name, RepliesEnabled: lib.FalsePtr(), LastActive: time.Now(),
	}
	assert.NoError(tm.CreateThread(th, "body", nil))
	assert.NotNil(th.RepliesEnabled)
	assert.False(*th.RepliesEnabled)

	var loaded models.BoardThread
	assert.NoError(db.First(&loaded, th.ID).Error)
	assert.NotNil(loaded.RepliesEnabled)
	assert.False(*loaded.RepliesEnabled)
}
