package factrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

const (
	trustedSurveySections = `FROM factrak_surveys fs
JOIN course_offerings o ON o.id = fs.course_offering_id AND o.deleted_at IS NULL AND o.course_in_semester_id IS NOT NULL`
	trustedSurveyWhere = `WHERE fs.deleted_at IS NULL AND fs.course_link_verification_source IN (?, ?)`
)

// Only adds, so a rerun keeps what the user hid, removed or chose
func SyncReviewedUserCourses(db *gorm.DB, log *zap.SugaredLogger) (courses, sections int64, err error) {
	if log == nil {
		log = zap.NewNop().Sugar()
	}
	sources := []interface{}{models.CourseLinkVerificationSourceVerified, models.CourseLinkVerificationSourceAuthorConfirmed}

	tx := db.Begin()
	if tx.Error != nil {
		return 0, 0, tx.Error
	}
	defer tx.RollbackUnlessCommitted()

	// Entries from here on are this run's
	var before struct{ ID uint }
	if err := tx.Table("user_courses").Select("COALESCE(MAX(id), 0) AS id").Scan(&before).Error; err != nil {
		return 0, 0, err
	}

	res := tx.Exec(`INSERT INTO user_courses (created_at, updated_at, user_id, course_canonical_id, course_in_semester_id, hidden)
SELECT UTC_TIMESTAMP(), UTC_TIMESTAMP(), fs.user_id, o.course_canonical_id, o.course_in_semester_id, 0
`+trustedSurveySections+`
`+trustedSurveyWhere+`
  AND NOT EXISTS (SELECT 1 FROM user_courses uc
    WHERE uc.user_id = fs.user_id AND uc.course_in_semester_id = o.course_in_semester_id)
GROUP BY fs.user_id, o.course_canonical_id, o.course_in_semester_id`, sources...)
	if res.Error != nil {
		return 0, 0, res.Error
	}
	courses = res.RowsAffected

	res = tx.Exec(`INSERT INTO user_course_offerings (created_at, updated_at, user_course_id, course_offering_id)
SELECT DISTINCT UTC_TIMESTAMP(), UTC_TIMESTAMP(), uc.id, o.id
`+trustedSurveySections+`
JOIN user_courses uc ON uc.user_id = fs.user_id AND uc.course_in_semester_id = o.course_in_semester_id
`+trustedSurveyWhere+`
  AND uc.id > ?
  AND NOT EXISTS (SELECT 1 FROM user_course_offerings uco WHERE uco.user_course_id = uc.id AND uco.course_offering_id = o.id)`, append(sources, before.ID)...)
	if res.Error != nil {
		return 0, 0, res.Error
	}
	sections = res.RowsAffected

	res = tx.Exec(`UPDATE factrak_surveys fs SET fs.user_course_id = (
  SELECT uc.id FROM user_courses uc
  LEFT JOIN course_in_semesters cis ON cis.id = uc.course_in_semester_id
  WHERE uc.user_id = fs.user_id AND uc.course_canonical_id = fs.course_canonical_id
  ORDER BY cis.strm = COALESCE(
      (SELECT o.strm FROM course_offerings o WHERE o.id = fs.course_offering_id),
      CASE fs.semester_season
        WHEN ? THEN 1000 + MOD(fs.semester_year + 1, 100) * 10 + 1
        WHEN ? THEN 1000 + MOD(fs.semester_year, 100) * 10 + 2
        WHEN ? THEN 1000 + MOD(fs.semester_year, 100) * 10 + 3
      END) DESC, cis.strm DESC, uc.id DESC
  LIMIT 1)
WHERE fs.user_course_id IS NULL AND fs.course_canonical_id IS NOT NULL AND fs.deleted_at IS NULL
  AND EXISTS (SELECT 1 FROM user_courses uc WHERE uc.user_id = fs.user_id AND uc.course_canonical_id = fs.course_canonical_id)`,
		models.FactrakSurveySemesterSeasonFall, models.FactrakSurveySemesterSeasonWinterStudy, models.FactrakSurveySemesterSeasonSpring)
	if res.Error != nil {
		return 0, 0, res.Error
	}
	linked := res.RowsAffected

	if err := tx.Commit().Error; err != nil {
		return 0, 0, err
	}
	log.Infof("My courses from reviews: added %d courses, %d sections; linked %d reviews", courses, sections, linked)
	return courses, sections, nil
}
