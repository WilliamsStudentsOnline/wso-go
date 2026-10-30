package boardbridge

import (
	"fmt"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// Bridge dual-writes legacy bulletin/discussion/ride mutations into board_* tables
// Failures are logged and swallowed so legacy endpoints stay authoritative
type Bridge struct {
	db      *gorm.DB
	log     *zap.SugaredLogger
	threads *models.BoardThreadModel
	posts   *models.BoardPostModel
}

func New(db *gorm.DB, log *zap.SugaredLogger) *Bridge {
	if log == nil {
		log = zap.NewNop().Sugar()
	}
	return &Bridge{
		db:      db,
		log:     log,
		threads: models.NewBoardThreadModel(db, log),
		posts:   models.NewBoardPostModel(db, log),
	}
}

func (b *Bridge) AfterCreateBulletin(bulletin *models.Bulletin) {
	if err := b.createBulletin(bulletin); err != nil {
		b.log.Warnf("board dual-write create bulletin %d: %v", bulletin.ID, err)
	}
}

func (b *Bridge) AfterUpdateBulletin(bulletin *models.Bulletin) {
	if err := b.updateBulletin(bulletin); err != nil {
		b.log.Warnf("board dual-write update bulletin %d: %v", bulletin.ID, err)
	}
}

func (b *Bridge) AfterDeleteBulletin(bulletin *models.Bulletin) {
	if err := b.deleteMappedThread(models.BoardBackfillSourceBulletins, bulletin.ID); err != nil {
		b.log.Warnf("board dual-write delete bulletin %d: %v", bulletin.ID, err)
	}
}

func (b *Bridge) AfterCreateDiscussion(discussion *models.Discussion) {
	if err := b.createDiscussion(discussion); err != nil {
		b.log.Warnf("board dual-write create discussion %d: %v", discussion.ID, err)
	}
}

func (b *Bridge) AfterDeleteDiscussion(discussion *models.Discussion) {
	if err := b.deleteMappedThread(models.BoardBackfillSourceDiscussions, discussion.ID); err != nil {
		b.log.Warnf("board dual-write delete discussion %d: %v", discussion.ID, err)
	}
}

func (b *Bridge) AfterCreatePost(post *models.Post) {
	if err := b.createPost(post); err != nil {
		b.log.Warnf("board dual-write create post %d: %v", post.ID, err)
	}
}

func (b *Bridge) AfterUpdatePost(post *models.Post) {
	if err := b.updatePost(post); err != nil {
		b.log.Warnf("board dual-write update post %d: %v", post.ID, err)
	}
}

// AfterDeletePost mirrors a legacy post delete
// discussionAlsoDeleted is true when legacy also soft-deleted the discussion
func (b *Bridge) AfterDeletePost(post *models.Post, discussionAlsoDeleted bool) {
	if err := b.deletePost(post, discussionAlsoDeleted); err != nil {
		b.log.Warnf("board dual-write delete post %d: %v", post.ID, err)
	}
}

func (b *Bridge) AfterCreateRide(ride *models.BulletinRide) {
	if err := b.createRide(ride); err != nil {
		b.log.Warnf("board dual-write create ride %d: %v", ride.ID, err)
	}
}

func (b *Bridge) AfterUpdateRide(ride *models.BulletinRide) {
	if err := b.updateRide(ride); err != nil {
		b.log.Warnf("board dual-write update ride %d: %v", ride.ID, err)
	}
}

func (b *Bridge) AfterDeleteRide(ride *models.BulletinRide) {
	if err := b.deleteMappedThread(models.BoardBackfillSourceBulletinRides, ride.ID); err != nil {
		b.log.Warnf("board dual-write delete ride %d: %v", ride.ID, err)
	}
}

func (b *Bridge) createBulletin(bulletin *models.Bulletin) error {
	typ := mapBulletinType(bulletin.Type)
	if typ == "" {
		return fmt.Errorf("unknown bulletin type %q", bulletin.Type)
	}

	name := b.userName(bulletin.UserID, "")
	var startsAt *time.Time
	if !bulletin.StartDate.IsZero() {
		startsAt = &bulletin.StartDate
	}
	thread := &models.BoardThread{
		Type:           typ,
		Title:          bulletin.Title,
		UserID:         bulletin.UserID,
		ExUserName:     name,
		RepliesEnabled: lib.TruePtr(),
		StartsAt:       startsAt,
		EndsAt:         bulletin.EndDate,
		LastActive:     time.Now(),
	}
	if err := b.threads.CreateThread(thread, bulletin.Body, nil); err != nil {
		return err
	}
	return b.recordMap(models.BoardBackfillSourceBulletins, bulletin.ID, thread.ID)
}

func (b *Bridge) updateBulletin(bulletin *models.Bulletin) error {
	threadID, ok, err := b.lookupMappedID(models.BoardBackfillSourceBulletins, bulletin.ID)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	var thread models.BoardThread
	if err := b.db.First(&thread, threadID).Error; err != nil {
		return err
	}
	thread.Title = bulletin.Title
	var startsAt *time.Time
	if !bulletin.StartDate.IsZero() {
		startsAt = &bulletin.StartDate
	}
	thread.StartsAt = startsAt
	thread.EndsAt = bulletin.EndDate
	if err := b.threads.UpdateThread(&thread); err != nil {
		return err
	}
	return b.threads.UpdateOPContent(threadID, bulletin.Body)
}

func (b *Bridge) createDiscussion(discussion *models.Discussion) error {
	posts := discussion.Posts
	if len(posts) == 0 {
		if err := b.db.Where("discussion_id = ?", discussion.ID).Order("created_at asc").Find(&posts).Error; err != nil {
			return err
		}
	}
	body := ""
	var op *models.Post
	if len(posts) > 0 {
		op = posts[0]
		body = op.Content
	}
	name := b.userName(discussion.UserID, discussion.ExUserName)
	thread := &models.BoardThread{
		Type:           models.BoardThreadTypeDiscussion,
		Title:          discussion.Title,
		UserID:         discussion.UserID,
		ExUserName:     name,
		RepliesEnabled: lib.TruePtr(),
		LastActive:     time.Now(),
	}
	if !discussion.LastActive.IsZero() {
		thread.LastActive = discussion.LastActive
	}
	if err := b.threads.CreateThread(thread, body, nil); err != nil {
		return err
	}
	if err := b.recordMap(models.BoardBackfillSourceDiscussions, discussion.ID, thread.ID); err != nil {
		return err
	}
	if op != nil && len(thread.Posts) > 0 {
		return b.recordMap(models.BoardBackfillSourcePosts, op.ID, thread.Posts[0].ID)
	}
	return nil
}

func (b *Bridge) createPost(post *models.Post) error {
	threadID, ok, err := b.lookupMappedID(models.BoardBackfillSourceDiscussions, post.DiscussionID)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	bp := &models.BoardPost{
		UserID:     post.UserID,
		ExUserName: b.userName(post.UserID, post.ExUserName),
		Content:    post.Content,
	}
	if err := b.posts.CreateReply(threadID, bp); err != nil {
		return err
	}
	return b.recordMap(models.BoardBackfillSourcePosts, post.ID, bp.ID)
}

func (b *Bridge) updatePost(post *models.Post) error {
	boardPostID, err := b.findBoardPostID(post)
	if err != nil {
		return err
	}
	if boardPostID == 0 {
		return nil
	}

	var bp models.BoardPost
	if err := b.db.First(&bp, boardPostID).Error; err != nil {
		return err
	}
	bp.Content = post.Content
	return b.posts.UpdatePost(&bp)
}

func (b *Bridge) deletePost(post *models.Post, discussionAlsoDeleted bool) error {
	if discussionAlsoDeleted {
		return b.deleteMappedThread(models.BoardBackfillSourceDiscussions, post.DiscussionID)
	}

	boardPostID, err := b.findBoardPostID(post)
	if err != nil {
		return err
	}
	if boardPostID == 0 {
		return nil
	}

	var bp models.BoardPost
	if err := b.db.First(&bp, boardPostID).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil
		}
		return err
	}

	// Soft-delete the board post without OP→thread cascade (legacy keeps the discussion)
	if bp.IsOP != nil && *bp.IsOP {
		return b.db.Delete(&bp).Error
	}
	return b.posts.DeletePost(&bp)
}

func (b *Bridge) createRide(ride *models.BulletinRide) error {
	name := b.userName(ride.UserID, "")
	offering := false
	if ride.Offer != nil {
		offering = *ride.Offer
	}
	startsAt := ride.Date
	thread := &models.BoardThread{
		Type:           models.BoardThreadTypeRide,
		Title:          fmt.Sprintf("%s to %s", ride.Source, ride.Destination),
		UserID:         ride.UserID,
		ExUserName:     name,
		RepliesEnabled: lib.TruePtr(),
		StartsAt:       &startsAt,
		LastActive:     time.Now(),
	}
	rideInfo := &models.BoardRideInformation{
		OfferingRide: lib.BoolToPtr(offering),
		Source:       ride.Source,
		Destination:  ride.Destination,
	}
	if err := b.threads.CreateThread(thread, ride.Body, rideInfo); err != nil {
		return err
	}
	return b.recordMap(models.BoardBackfillSourceBulletinRides, ride.ID, thread.ID)
}

func (b *Bridge) updateRide(ride *models.BulletinRide) error {
	threadID, ok, err := b.lookupMappedID(models.BoardBackfillSourceBulletinRides, ride.ID)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	var thread models.BoardThread
	if err := b.db.First(&thread, threadID).Error; err != nil {
		return err
	}
	startsAt := ride.Date
	thread.Title = fmt.Sprintf("%s to %s", ride.Source, ride.Destination)
	thread.StartsAt = &startsAt
	if err := b.threads.UpdateThread(&thread); err != nil {
		return err
	}
	if err := b.threads.UpdateOPContent(threadID, ride.Body); err != nil {
		return err
	}

	offering := false
	if ride.Offer != nil {
		offering = *ride.Offer
	}
	return b.db.Model(&models.BoardRideInformation{}).
		Where("thread_id = ?", threadID).
		Updates(map[string]interface{}{
			"offering_ride": offering,
			"source":        ride.Source,
			"destination":   ride.Destination,
		}).Error
}

func (b *Bridge) deleteMappedThread(sourceTable string, sourceID uint) error {
	threadID, ok, err := b.lookupMappedID(sourceTable, sourceID)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	var thread models.BoardThread
	if err := b.db.First(&thread, threadID).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil
		}
		return err
	}
	return b.threads.DeleteThread(&thread)
}

func (b *Bridge) findBoardPostID(post *models.Post) (uint, error) {
	id, ok, err := b.lookupMappedID(models.BoardBackfillSourcePosts, post.ID)
	if err != nil {
		return 0, err
	}
	if ok {
		return id, nil
	}

	// Fallback for backfilled posts: match preserved created_at within the mapped thread
	threadID, ok, err := b.lookupMappedID(models.BoardBackfillSourceDiscussions, post.DiscussionID)
	if err != nil || !ok {
		return 0, err
	}
	var bp models.BoardPost
	err = b.db.Where("thread_id = ? AND user_id = ? AND created_at = ?", threadID, post.UserID, post.CreatedAt).
		First(&bp).Error
	if gorm.IsRecordNotFoundError(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return bp.ID, nil
}

func (b *Bridge) lookupMappedID(sourceTable string, sourceID uint) (uint, bool, error) {
	var m models.BoardBackfillMap
	err := b.db.Where("source_table = ? AND source_id = ?", sourceTable, sourceID).First(&m).Error
	if gorm.IsRecordNotFoundError(err) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return m.ThreadID, true, nil
}

func (b *Bridge) recordMap(sourceTable string, sourceID, threadID uint) error {
	return b.db.Create(&models.BoardBackfillMap{
		SourceTable: sourceTable,
		SourceID:    sourceID,
		ThreadID:    threadID,
	}).Error
}

func (b *Bridge) userName(userID uint, fallback string) string {
	if fallback != "" {
		return fallback
	}
	var u models.User
	if err := b.db.Select("name").First(&u, userID).Error; err != nil {
		return ""
	}
	return u.Name
}

func mapBulletinType(t string) string {
	switch t {
	case models.BulletinTypeAnnouncement:
		return models.BoardThreadTypeAnnouncement
	case models.BulletinTypeJob:
		return models.BoardThreadTypeJob
	case models.BulletinTypeExchange:
		return models.BoardThreadTypeExchange
	case models.BulletinTypeLostAndFound:
		return models.BoardThreadTypeLostAndFound
	default:
		return ""
	}
}
