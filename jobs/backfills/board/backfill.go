package board

import (
	"fmt"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// BackfillResult summarizes a copy-only board backfill run.
type BackfillResult struct {
	DiscussionsMigrated int
	BulletinsMigrated   int
	RidesMigrated       int
	Skipped             int
}

// Run copies legacy discussions/posts, bulletins, and bulletin_rides into board_*
// tables with new IDs. Does not modify legacy rows. Idempotent via board_backfill_map.
func Run(db *gorm.DB, log *zap.SugaredLogger) (*BackfillResult, error) {
	if log == nil {
		log = zap.NewNop().Sugar()
	}
	res := &BackfillResult{}

	n, skipped, err := migrateDiscussions(db, log)
	if err != nil {
		return res, err
	}
	res.DiscussionsMigrated = n
	res.Skipped += skipped

	n, skipped, err = migrateBulletins(db, log)
	if err != nil {
		return res, err
	}
	res.BulletinsMigrated = n
	res.Skipped += skipped

	n, skipped, err = migrateRides(db, log)
	if err != nil {
		return res, err
	}
	res.RidesMigrated = n
	res.Skipped += skipped

	return res, nil
}

func alreadyMigrated(db *gorm.DB, sourceTable string, sourceID uint) (bool, error) {
	var count int
	err := db.Model(&models.BoardBackfillMap{}).
		Where("source_table = ? AND source_id = ?", sourceTable, sourceID).
		Count(&count).Error
	return count > 0, err
}

func recordMap(tx *gorm.DB, sourceTable string, sourceID, threadID uint) error {
	return tx.Create(&models.BoardBackfillMap{
		SourceTable: sourceTable,
		SourceID:    sourceID,
		ThreadID:    threadID,
	}).Error
}

func userName(db *gorm.DB, userID uint, fallback string) string {
	if fallback != "" {
		return fallback
	}
	var u models.User
	if err := db.Select("name").First(&u, userID).Error; err != nil {
		return ""
	}
	return u.Name
}

func migrateDiscussions(db *gorm.DB, log *zap.SugaredLogger) (migrated, skipped int, err error) {
	var discussions []*models.Discussion
	if err = db.Find(&discussions).Error; err != nil {
		return
	}

	for _, d := range discussions {
		done, e := alreadyMigrated(db, models.BoardBackfillSourceDiscussions, d.ID)
		if e != nil {
			err = e
			return
		}
		if done {
			skipped++
			continue
		}

		var posts []*models.Post
		if err = db.Where("discussion_id = ?", d.ID).Order("created_at asc").Find(&posts).Error; err != nil {
			return
		}

		tx := db.Begin()
		if err = tx.Error; err != nil {
			return
		}

		name := userName(tx, d.UserID, d.ExUserName)
		thread := &models.BoardThread{
			Type:           models.BoardThreadTypeDiscussion,
			Title:          d.Title,
			UserID:         d.UserID,
			ExUserName:     name,
			RepliesEnabled: true,
			LastActive:     d.LastActive,
			ReplyCount:     0,
		}
		if thread.LastActive.IsZero() {
			thread.LastActive = d.CreatedAt
		}
		if err = tx.Create(thread).Error; err != nil {
			tx.Rollback()
			return
		}

		// Preserve legacy timestamps after create.
		if err = tx.Model(thread).Updates(map[string]interface{}{
			"created_at":  d.CreatedAt,
			"updated_at":  d.UpdatedAt,
			"last_active": thread.LastActive,
		}).Error; err != nil {
			tx.Rollback()
			return
		}

		replyCount := 0
		for i, p := range posts {
			bp := &models.BoardPost{
				ThreadID:   thread.ID,
				UserID:     p.UserID,
				ExUserName: userName(tx, p.UserID, p.ExUserName),
				Content:    p.Content,
				IsOP:       i == 0,
			}
			if err = tx.Create(bp).Error; err != nil {
				tx.Rollback()
				return
			}
			if err = tx.Model(bp).Updates(map[string]interface{}{
				"created_at": p.CreatedAt,
				"updated_at": p.UpdatedAt,
			}).Error; err != nil {
				tx.Rollback()
				return
			}
			if !bp.IsOP {
				replyCount++
			}
		}
		if err = tx.Model(thread).Update("reply_count", replyCount).Error; err != nil {
			tx.Rollback()
			return
		}

		if err = recordMap(tx, models.BoardBackfillSourceDiscussions, d.ID, thread.ID); err != nil {
			tx.Rollback()
			return
		}
		if err = tx.Commit().Error; err != nil {
			return
		}
		migrated++
	}
	log.Infof("migrated %d discussions (%d skipped)", migrated, skipped)
	return
}

func migrateBulletins(db *gorm.DB, log *zap.SugaredLogger) (migrated, skipped int, err error) {
	var bulletins []*models.Bulletin
	if err = db.Find(&bulletins).Error; err != nil {
		return
	}

	for _, b := range bulletins {
		done, e := alreadyMigrated(db, models.BoardBackfillSourceBulletins, b.ID)
		if e != nil {
			err = e
			return
		}
		if done {
			skipped++
			continue
		}

		typ := mapBulletinType(b.Type)
		if typ == "" {
			log.Warnf("skipping bulletin %d with unknown type %q", b.ID, b.Type)
			skipped++
			continue
		}

		tx := db.Begin()
		if err = tx.Error; err != nil {
			return
		}

		name := userName(tx, b.UserID, "")
		var startsAt *time.Time
		if !b.StartDate.IsZero() {
			startsAt = &b.StartDate
		}
		thread := &models.BoardThread{
			Type:           typ,
			Title:          b.Title,
			UserID:         b.UserID,
			ExUserName:     name,
			RepliesEnabled: true,
			StartsAt:       startsAt,
			EndsAt:         b.EndDate,
			LastActive:     b.CreatedAt,
		}
		if thread.LastActive.IsZero() {
			thread.LastActive = time.Now()
		}
		if err = tx.Create(thread).Error; err != nil {
			tx.Rollback()
			return
		}
		if err = tx.Model(thread).Updates(map[string]interface{}{
			"created_at":  b.CreatedAt,
			"updated_at":  b.UpdatedAt,
			"last_active": thread.LastActive,
		}).Error; err != nil {
			tx.Rollback()
			return
		}

		op := &models.BoardPost{
			ThreadID:   thread.ID,
			UserID:     b.UserID,
			ExUserName: name,
			Content:    b.Body,
			IsOP:       true,
		}
		if err = tx.Create(op).Error; err != nil {
			tx.Rollback()
			return
		}
		if err = tx.Model(op).Updates(map[string]interface{}{
			"created_at": b.CreatedAt,
			"updated_at": b.UpdatedAt,
		}).Error; err != nil {
			tx.Rollback()
			return
		}

		if err = recordMap(tx, models.BoardBackfillSourceBulletins, b.ID, thread.ID); err != nil {
			tx.Rollback()
			return
		}
		if err = tx.Commit().Error; err != nil {
			return
		}
		migrated++
	}
	log.Infof("migrated %d bulletins (%d skipped)", migrated, skipped)
	return
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

func migrateRides(db *gorm.DB, log *zap.SugaredLogger) (migrated, skipped int, err error) {
	var rides []*models.BulletinRide
	if err = db.Find(&rides).Error; err != nil {
		return
	}

	for _, r := range rides {
		done, e := alreadyMigrated(db, models.BoardBackfillSourceBulletinRides, r.ID)
		if e != nil {
			err = e
			return
		}
		if done {
			skipped++
			continue
		}

		tx := db.Begin()
		if err = tx.Error; err != nil {
			return
		}

		name := userName(tx, r.UserID, "")
		offering := false
		if r.Offer != nil {
			offering = *r.Offer
		}
		title := fmt.Sprintf("%s to %s", r.Source, r.Destination)
		startsAt := r.Date
		thread := &models.BoardThread{
			Type:           models.BoardThreadTypeRide,
			Title:          title,
			UserID:         r.UserID,
			ExUserName:     name,
			RepliesEnabled: true,
			StartsAt:       &startsAt,
			LastActive:     r.CreatedAt,
		}
		if thread.LastActive.IsZero() {
			thread.LastActive = time.Now()
		}
		if err = tx.Create(thread).Error; err != nil {
			tx.Rollback()
			return
		}
		if err = tx.Model(thread).Updates(map[string]interface{}{
			"created_at":  r.CreatedAt,
			"updated_at":  r.UpdatedAt,
			"last_active": thread.LastActive,
		}).Error; err != nil {
			tx.Rollback()
			return
		}

		op := &models.BoardPost{
			ThreadID:   thread.ID,
			UserID:     r.UserID,
			ExUserName: name,
			Content:    r.Body,
			IsOP:       true,
		}
		if err = tx.Create(op).Error; err != nil {
			tx.Rollback()
			return
		}
		if err = tx.Model(op).Updates(map[string]interface{}{
			"created_at": r.CreatedAt,
			"updated_at": r.UpdatedAt,
		}).Error; err != nil {
			tx.Rollback()
			return
		}

		meta := &models.BoardRideMeta{
			ThreadID:     thread.ID,
			OfferingRide: offering,
			Source:       r.Source,
			Destination:  r.Destination,
		}
		if err = tx.Create(meta).Error; err != nil {
			tx.Rollback()
			return
		}

		if err = recordMap(tx, models.BoardBackfillSourceBulletinRides, r.ID, thread.ID); err != nil {
			tx.Rollback()
			return
		}
		if err = tx.Commit().Error; err != nil {
			return
		}
		migrated++
	}
	log.Infof("migrated %d rides (%d skipped)", migrated, skipped)
	return
}
