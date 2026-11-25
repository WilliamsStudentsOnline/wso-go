package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

// Fall 2026 is 1271; 0 when unknown
func strmOf(season *string, year *int) int {
	if season == nil || year == nil || *year == 0 {
		return 0
	}
	switch *season {
	case FactrakSurveySemesterSeasonFall:
		return 1000 + ((*year+1)%100)*10 + 1
	case FactrakSurveySemesterSeasonWinterStudy:
		return 1000 + (*year%100)*10 + 2
	case FactrakSurveySemesterSeasonSpring:
		return 1000 + (*year%100)*10 + 3
	}
	return 0
}

func reviewStrm(db *gorm.DB, s *FactrakSurvey) (int, error) {
	if s.CourseOfferingID != nil {
		var o CourseOffering
		if err := db.Select("id, strm").First(&o, *s.CourseOfferingID).Error; err == nil {
			return o.Strm, nil
		} else if !gorm.IsRecordNotFoundError(err) {
			return 0, err
		}
	}
	return strmOf(s.SemesterSeason, s.SemesterYear), nil
}

// Entries for terms that haven't started get none; db must be a transaction
func linkReviews(db *gorm.DB, uc *UserCourse) error {
	entryStrm := 0
	if uc.CourseInSemesterID != nil {
		var cis CourseInSemester
		if err := db.Select("id, strm").First(&cis, *uc.CourseInSemesterID).Error; err != nil {
			return err
		}
		entryStrm = cis.Strm
		if entryStrm > StartedStrm(time.Now()) {
			return nil
		}
	}
	var reviews []*FactrakSurvey
	err := db.Where("user_id = ? AND course_canonical_id = ? AND user_course_id IS NULL", uc.UserID, uc.CourseCanonicalID).
		Find(&reviews).Error
	if err != nil {
		return err
	}
	for _, s := range reviews {
		strm, err := reviewStrm(db, s)
		if err != nil {
			return err
		}
		if strm != 0 && entryStrm != 0 && strm != entryStrm {
			continue
		}
		if err := db.Model(s).Update("user_course_id", uc.ID).Error; err != nil {
			return err
		}
	}
	return nil
}

func taughtBy(db *gorm.DB, offeringIDs []uint, profID uint) (*uint, error) {
	if len(offeringIDs) == 0 {
		return nil, nil
	}
	var ids []uint
	err := db.Model(&CourseOfferingInstructor{}).Where("course_offering_id IN (?) AND user_id = ?", offeringIDs, profID).
		Order("course_offering_id").Pluck("course_offering_id", &ids).Error
	if err != nil || len(ids) == 0 {
		return nil, err
	}
	return &ids[0], nil
}

// The author chose these sections, so their reviews become author_confirmed
func placeReviews(db *gorm.DB, userCourseID uint, offeringIDs []uint) error {
	var reviews []*FactrakSurvey
	if err := db.Where("user_course_id = ?", userCourseID).Find(&reviews).Error; err != nil {
		return err
	}
	for _, s := range reviews {
		id, err := taughtBy(db, offeringIDs, s.ProfessorID)
		if err != nil {
			return err
		}
		if id == nil || (s.CourseOfferingID != nil && *s.CourseOfferingID == *id) {
			continue
		}
		err = db.Model(s).Updates(map[string]interface{}{"course_offering_id": *id, "course_link_verification_source": CourseLinkVerificationSourceAuthorConfirmed}).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func unplaceReviews(db *gorm.DB, userCourseID uint, removed []uint) error {
	if len(removed) == 0 {
		return nil
	}
	var reviews []*FactrakSurvey
	if err := db.Where("user_course_id = ? AND course_offering_id IN (?)", userCourseID, removed).Find(&reviews).Error; err != nil {
		return err
	}
	var remaining []uint
	if err := db.Model(&UserCourseOffering{}).Where("user_course_id = ?", userCourseID).Pluck("course_offering_id", &remaining).Error; err != nil {
		return err
	}
	for _, s := range reviews {
		id, err := taughtBy(db, remaining, s.ProfessorID)
		if err != nil {
			return err
		}
		var offering interface{} = gorm.Expr("NULL")
		if id != nil {
			offering = *id
		}
		if err := db.Model(s).Update("course_offering_id", offering).Error; err != nil {
			return err
		}
	}
	return nil
}

// An unlinked review of the same course counts once the entry's term has started
func (m *UserCourseModel) ReviewedEntryIDs(userID uint) (map[uint]bool, error) {
	var ids []uint
	err := m.db.Model(&UserCourse{}).Where("user_id = ?", userID).Where(`EXISTS (
SELECT 1 FROM factrak_surveys fs WHERE fs.deleted_at IS NULL AND (fs.user_course_id = user_courses.id
  OR (fs.user_course_id IS NULL AND fs.user_id = user_courses.user_id AND fs.course_canonical_id = user_courses.course_canonical_id
    AND NOT EXISTS (SELECT 1 FROM course_in_semesters cis WHERE cis.id = user_courses.course_in_semester_id AND cis.strm > ?))))`,
		StartedStrm(time.Now())).
		Pluck("id", &ids).Error
	out := make(map[uint]bool, len(ids))
	for _, id := range ids {
		out[id] = true
	}
	return out, err
}
