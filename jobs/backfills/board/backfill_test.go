package board_test

import (
	"testing"
	"time"

	boardbackfill "github.com/WilliamsStudentsOnline/wso-go/jobs/backfills/board"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestBoardBackfill_CopyOnlyIdempotent(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	log := zaptest.NewLogger(t).Sugar()

	u := models.User{Type: models.UserTypeStudent, UnixID: "u1", Name: "User One"}
	assert.NoError(db.Create(&u).Error)

	d := models.Discussion{
		UserID: u.ID, ExUserName: u.Name, Title: "Disc", LastActive: time.Now(),
	}
	assert.NoError(db.Create(&d).Error)
	assert.NoError(db.Create(&models.Post{
		DiscussionID: d.ID, UserID: u.ID, ExUserName: u.Name, Content: "op content",
	}).Error)
	assert.NoError(db.Create(&models.Post{
		DiscussionID: d.ID, UserID: u.ID, ExUserName: u.Name, Content: "reply",
	}).Error)

	b := models.Bulletin{
		Type: models.BulletinTypeJob, Title: "Job", Body: "job body",
		UserID: u.ID, StartDate: time.Now().Add(-time.Hour),
	}
	assert.NoError(db.Create(&b).Error)

	ride := models.BulletinRide{
		Body: "ride body", Date: time.Now().Add(24 * time.Hour),
		Offer: lib.BoolToPtr(true), Source: "A", Destination: "B", UserID: u.ID,
	}
	assert.NoError(db.Create(&ride).Error)

	res, err := boardbackfill.Run(db, log)
	assert.NoError(err)
	assert.Equal(1, res.DiscussionsMigrated)
	assert.Equal(1, res.BulletinsMigrated)
	assert.Equal(1, res.RidesMigrated)

	var threadCount, postCount, rideMetaCount int
	assert.NoError(db.Model(&models.BoardThread{}).Count(&threadCount).Error)
	assert.NoError(db.Model(&models.BoardPost{}).Count(&postCount).Error)
	assert.NoError(db.Model(&models.BoardRideMeta{}).Count(&rideMetaCount).Error)
	assert.Equal(3, threadCount)
	assert.Equal(4, postCount) // 2 discussion posts + 1 bulletin op + 1 ride op
	assert.Equal(1, rideMetaCount)

	// Legacy untouched
	var discCount, bullCount, rideCount int
	assert.NoError(db.Model(&models.Discussion{}).Count(&discCount).Error)
	assert.NoError(db.Model(&models.Bulletin{}).Count(&bullCount).Error)
	assert.NoError(db.Model(&models.BulletinRide{}).Count(&rideCount).Error)
	assert.Equal(1, discCount)
	assert.Equal(1, bullCount)
	assert.Equal(1, rideCount)

	// Idempotent
	res2, err := boardbackfill.Run(db, log)
	assert.NoError(err)
	assert.Equal(0, res2.DiscussionsMigrated)
	assert.Equal(0, res2.BulletinsMigrated)
	assert.Equal(0, res2.RidesMigrated)
	assert.Equal(3, res2.Skipped)

	assert.NoError(db.Model(&models.BoardThread{}).Count(&threadCount).Error)
	assert.Equal(3, threadCount)
}
