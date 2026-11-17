package factrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// Depends on catalog instructors, so rerun after catalog loads
func AssignCourseLinkVerificationSources(db *gorm.DB, log *zap.SugaredLogger) (map[string]int, error) {
	if log == nil {
		log = zap.NewNop().Sugar()
	}
	const (
		linked       = "course_canonical_id IS NOT NULL"
		notAuthor    = "(course_link_verification_source IS NULL OR course_link_verification_source <> '" + models.CourseLinkVerificationSourceAuthorConfirmed + "')"
		taughtCourse = `EXISTS (SELECT 1 FROM course_offerings o
  JOIN course_offering_instructors oi ON oi.course_offering_id = o.id AND oi.deleted_at IS NULL
  WHERE o.course_canonical_id = factrak_surveys.course_canonical_id AND o.deleted_at IS NULL
    AND oi.user_id = factrak_surveys.professor_id)`
		inCatalog = `EXISTS (SELECT 1 FROM course_offering_instructors oi
  WHERE oi.user_id = factrak_surveys.professor_id AND oi.deleted_at IS NULL)`
	)
	steps := []struct {
		source interface{}
		where  string
	}{
		{nil, "course_canonical_id IS NULL"},
		{models.CourseLinkVerificationSourceMismatch, linked + " AND " + notAuthor},
		{models.CourseLinkVerificationSourceHistorical, linked + " AND " + notAuthor + " AND NOT " + inCatalog},
		{models.CourseLinkVerificationSourceConfirmed, linked + " AND " + notAuthor + " AND " + taughtCourse},
		{models.CourseLinkVerificationSourceVerified, "course_offering_id IS NOT NULL AND " + notAuthor},
	}

	tx := db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer tx.RollbackUnlessCommitted()
	for _, s := range steps {
		if err := tx.Exec("UPDATE factrak_surveys SET course_link_verification_source = ? WHERE "+s.where, s.source).Error; err != nil {
			return nil, err
		}
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	var rows []struct {
		CourseLinkVerificationSource *string
		N                            int
	}
	if err := db.Table("factrak_surveys").Select("course_link_verification_source, COUNT(*) AS n").
		Where("deleted_at IS NULL").Group("course_link_verification_source").Scan(&rows).Error; err != nil {
		return nil, err
	}
	counts := map[string]int{}
	for _, r := range rows {
		key := "unlinked"
		if r.CourseLinkVerificationSource != nil {
			key = *r.CourseLinkVerificationSource
		}
		counts[key] = r.N
	}
	log.Infof("Factrak link sources: %v", counts)
	return counts, nil
}
