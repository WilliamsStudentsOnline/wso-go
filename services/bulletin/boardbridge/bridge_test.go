package boardbridge_test

import (
	"testing"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services/bulletin/boardbridge"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestBridge_BulletinCreateUpdateDelete(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	log := zaptest.NewLogger(t).Sugar()
	bridge := boardbridge.New(db, log)

	u := models.User{Type: models.UserTypeStudent, UnixID: "u1", Name: "User One"}
	assert.NoError(db.Create(&u).Error)

	b := models.Bulletin{
		Type: models.BulletinTypeJob, Title: "Job", Body: "body",
		UserID: u.ID, StartDate: time.Now().Add(-time.Hour),
	}
	assert.NoError(db.Create(&b).Error)
	bridge.AfterCreateBulletin(&b)

	var m models.BoardBackfillMap
	assert.NoError(db.Where("source_table = ? AND source_id = ?", models.BoardBackfillSourceBulletins, b.ID).First(&m).Error)

	var thread models.BoardThread
	assert.NoError(db.First(&thread, m.ThreadID).Error)
	assert.Equal(models.BoardThreadTypeJob, thread.Type)
	assert.Equal("Job", thread.Title)

	var op models.BoardPost
	assert.NoError(db.Where("thread_id = ? AND is_op = ?", thread.ID, true).First(&op).Error)
	assert.Equal("body", op.Content)

	b.Title = "Job updated"
	b.Body = "body updated"
	assert.NoError(db.Save(&b).Error)
	bridge.AfterUpdateBulletin(&b)

	assert.NoError(db.First(&thread, m.ThreadID).Error)
	assert.Equal("Job updated", thread.Title)
	assert.NoError(db.Where("thread_id = ? AND is_op = ?", thread.ID, true).First(&op).Error)
	assert.Equal("body updated", op.Content)

	bridge.AfterDeleteBulletin(&b)
	var count int
	assert.NoError(db.Unscoped().Model(&models.BoardThread{}).Where("id = ? AND deleted_at IS NOT NULL", m.ThreadID).Count(&count).Error)
	assert.Equal(1, count)
}

func TestBridge_DiscussionAndPosts(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	log := zaptest.NewLogger(t).Sugar()
	bridge := boardbridge.New(db, log)

	u := models.User{Type: models.UserTypeStudent, UnixID: "u2", Name: "User Two"}
	assert.NoError(db.Create(&u).Error)

	d := models.Discussion{
		UserID: u.ID, ExUserName: u.Name, Title: "Disc", LastActive: time.Now(),
		Posts: []*models.Post{{
			UserID: u.ID, ExUserName: u.Name, Content: "op content",
		}},
	}
	assert.NoError(db.Create(&d).Error)
	bridge.AfterCreateDiscussion(&d)

	var discMap models.BoardBackfillMap
	assert.NoError(db.Where("source_table = ? AND source_id = ?", models.BoardBackfillSourceDiscussions, d.ID).First(&discMap).Error)

	var postMap models.BoardBackfillMap
	assert.NoError(db.Where("source_table = ? AND source_id = ?", models.BoardBackfillSourcePosts, d.Posts[0].ID).First(&postMap).Error)

	reply := models.Post{
		DiscussionID: d.ID, UserID: u.ID, ExUserName: u.Name, Content: "reply",
	}
	assert.NoError(db.Create(&reply).Error)
	bridge.AfterCreatePost(&reply)

	var replyMap models.BoardBackfillMap
	assert.NoError(db.Where("source_table = ? AND source_id = ?", models.BoardBackfillSourcePosts, reply.ID).First(&replyMap).Error)

	var boardReply models.BoardPost
	assert.NoError(db.First(&boardReply, replyMap.ThreadID).Error)
	assert.Equal("reply", boardReply.Content)
	assert.NotNil(boardReply.IsOP)
	assert.False(*boardReply.IsOP)

	reply.Content = "reply edited"
	assert.NoError(db.Save(&reply).Error)
	bridge.AfterUpdatePost(&reply)
	assert.NoError(db.First(&boardReply, replyMap.ThreadID).Error)
	assert.Equal("reply edited", boardReply.Content)

	bridge.AfterDeletePost(&reply, false)
	var deletedCount int
	assert.NoError(db.Unscoped().Model(&models.BoardPost{}).Where("id = ? AND deleted_at IS NOT NULL", replyMap.ThreadID).Count(&deletedCount).Error)
	assert.Equal(1, deletedCount)

	bridge.AfterDeleteDiscussion(&d)
	var threadCount int
	assert.NoError(db.Unscoped().Model(&models.BoardThread{}).Where("id = ? AND deleted_at IS NOT NULL", discMap.ThreadID).Count(&threadCount).Error)
	assert.Equal(1, threadCount)
}

func TestBridge_RideCreateUpdateDelete(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	log := zaptest.NewLogger(t).Sugar()
	bridge := boardbridge.New(db, log)

	u := models.User{Type: models.UserTypeStudent, UnixID: "u3", Name: "User Three"}
	assert.NoError(db.Create(&u).Error)

	ride := models.BulletinRide{
		Body: "ride body", Date: time.Now().Add(24 * time.Hour),
		Offer: lib.BoolToPtr(true), Source: "Williamstown", Destination: "Boston",
		UserID: u.ID,
	}
	assert.NoError(db.Create(&ride).Error)
	bridge.AfterCreateRide(&ride)

	var m models.BoardBackfillMap
	assert.NoError(db.Where("source_table = ? AND source_id = ?", models.BoardBackfillSourceBulletinRides, ride.ID).First(&m).Error)

	var thread models.BoardThread
	assert.NoError(db.First(&thread, m.ThreadID).Error)
	assert.Equal(models.BoardThreadTypeRide, thread.Type)
	assert.Equal("Williamstown to Boston", thread.Title)

	var rideInfo models.BoardRideInformation
	assert.NoError(db.Where("thread_id = ?", thread.ID).First(&rideInfo).Error)
	assert.NotNil(rideInfo.OfferingRide)
	assert.True(*rideInfo.OfferingRide)

	ride.Body = "updated body"
	ride.Offer = lib.BoolToPtr(false)
	assert.NoError(db.Save(&ride).Error)
	bridge.AfterUpdateRide(&ride)

	var op models.BoardPost
	assert.NoError(db.Where("thread_id = ? AND is_op = ?", thread.ID, true).First(&op).Error)
	assert.Equal("updated body", op.Content)
	assert.NoError(db.Where("thread_id = ?", thread.ID).First(&rideInfo).Error)
	assert.NotNil(rideInfo.OfferingRide)
	assert.False(*rideInfo.OfferingRide)

	bridge.AfterDeleteRide(&ride)
	var count int
	assert.NoError(db.Unscoped().Model(&models.BoardThread{}).Where("id = ? AND deleted_at IS NOT NULL", m.ThreadID).Count(&count).Error)
	assert.Equal(1, count)
}

func TestBridge_UpdateSkipsWhenUnmapped(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	log := zaptest.NewLogger(t).Sugar()
	bridge := boardbridge.New(db, log)

	b := models.Bulletin{
		Type: models.BulletinTypeAnnouncement, Title: "x", Body: "y",
		UserID: 1, StartDate: time.Now(),
	}
	b.ID = 99999
	// Should not panic or create rows
	bridge.AfterUpdateBulletin(&b)
	bridge.AfterDeleteBulletin(&b)

	var count int
	assert.NoError(db.Model(&models.BoardThread{}).Count(&count).Error)
	assert.Equal(0, count)
}
