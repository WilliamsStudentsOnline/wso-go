package models

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm"
)

var (
	ErrUserCourseHasReview = errors.New("user has a review of this course")
	ErrUserCourseSection   = errors.New("section is not part of this course in that term")
)

type UserCourseModel struct {
	db *gorm.DB
}

func NewUserCourseModel(db *gorm.DB) *UserCourseModel {
	return &UserCourseModel{db: db}
}

func (m *UserCourseModel) preloaded() *gorm.DB {
	return m.db.
		Preload("CourseCanonical").
		Preload("CourseCanonical.Listings").
		// A someday course shows its newest term's distributions
		Preload("CourseCanonical.CoursesInSemester", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, course_canonical_id, strm, div1, div2, div3, dpe, qfr, wac").
				Where("source = ?", CatalogSourceCatalog).Order("strm DESC")
		}).
		Preload("CourseInSemester").
		Preload("CourseInSemester.Offerings", "status = ?", CourseOfferingStatusOffered).
		Preload("Offerings").
		Preload("Offerings.CourseOffering").
		Preload("Offerings.CourseOffering.Meetings").
		Preload("Offerings.CourseOffering.Instructors")
}

func (m *UserCourseModel) ListUserCourses(userID uint, strm int) ([]*UserCourse, error) {
	q := m.preloaded().
		Joins("LEFT JOIN course_in_semesters cis ON cis.id = user_courses.course_in_semester_id").
		Where("user_courses.user_id = ?", userID)
	if strm > 0 {
		q = q.Where("cis.strm = ?", strm)
	}
	var out []*UserCourse
	err := q.Order("cis.strm IS NULL, cis.strm DESC, user_courses.id").Find(&out).Error
	return out, err
}

func (m *UserCourseModel) GetUserCourse(userID, id uint) (*UserCourse, error) {
	var uc UserCourse
	if err := m.preloaded().Where("user_id = ?", userID).First(&uc, id).Error; err != nil {
		return nil, err
	}
	return &uc, nil
}

// Adds sections to an existing entry, so a lecture and its lab can be added one at a time; caller checks offerings belong to cis
func (m *UserCourseModel) AddUserCourse(userID, canonicalID uint, cis *CourseInSemester, offerings []*CourseOffering, hidden *bool) (*UserCourse, error) {
	tx := m.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer tx.RollbackUnlessCommitted()
	inTx := &UserCourseModel{db: tx}
	uc, err := inTx.addInTx(userID, canonicalID, cis, offerings)
	if err != nil {
		return nil, err
	}
	if hidden != nil && uc.Hidden != *hidden {
		if err := tx.Model(uc).Update("hidden", *hidden).Error; err != nil {
			return nil, err
		}
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return m.GetUserCourse(userID, uc.ID)
}

// m.db must be a transaction
func (m *UserCourseModel) addInTx(userID, canonicalID uint, cis *CourseInSemester, offerings []*CourseOffering) (*UserCourse, error) {
	var uc UserCourse
	q := m.db.Set("gorm:query_option", "FOR UPDATE").
		Where("user_id = ? AND course_canonical_id = ?", userID, canonicalID)
	if cis != nil {
		q = q.Where("course_in_semester_id = ?", cis.ID)
	} else {
		q = q.Where("course_in_semester_id IS NULL")
	}
	err := q.First(&uc).Error
	if gorm.IsRecordNotFoundError(err) {
		uc = UserCourse{UserID: userID, CourseCanonicalID: canonicalID}
		if cis != nil {
			uc.CourseInSemesterID = &cis.ID
		}
		if err = m.db.Create(&uc).Error; err == nil {
			err = linkReviews(m.db, &uc)
		}
	}
	if err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(offerings))
	for _, o := range offerings {
		if err := addOffering(m.db, uc.ID, o.ID); err != nil {
			return nil, err
		}
		ids = append(ids, o.ID)
	}
	if err := placeReviews(m.db, uc.ID, ids); err != nil {
		return nil, err
	}
	return &uc, nil
}

func addOffering(tx *gorm.DB, userCourseID, offeringID uint) error {
	return tx.Exec(`INSERT INTO user_course_offerings (created_at, updated_at, user_course_id, course_offering_id)
SELECT UTC_TIMESTAMP(), UTC_TIMESTAMP(), ?, ? FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM user_course_offerings WHERE user_course_id = ? AND course_offering_id = ?)`,
		userCourseID, offeringID, userCourseID, offeringID).Error
}

func (m *UserCourseModel) SetUserCourseHidden(userID, id uint, hidden bool) (*UserCourse, error) {
	res := m.db.Model(&UserCourse{}).Where("id = ? AND user_id = ?", id, userID).Update("hidden", hidden)
	if res.Error != nil {
		return nil, res.Error
	}
	return m.GetUserCourse(userID, id)
}

// Caller checks the offerings belong to the entry's term
func (m *UserCourseModel) SetUserCourseOfferings(userID, id uint, offerings []*CourseOffering) (*UserCourse, error) {
	tx := m.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer tx.RollbackUnlessCommitted()
	var uc UserCourse
	if err := tx.Where("user_id = ?", userID).First(&uc, id).Error; err != nil {
		return nil, err
	}
	var before []uint
	if err := tx.Model(&UserCourseOffering{}).Where("user_course_id = ?", uc.ID).Pluck("course_offering_id", &before).Error; err != nil {
		return nil, err
	}
	if err := tx.Unscoped().Where("user_course_id = ?", uc.ID).Delete(&UserCourseOffering{}).Error; err != nil {
		return nil, err
	}
	chosen := map[uint]bool{}
	ids := make([]uint, 0, len(offerings))
	for _, o := range offerings {
		if err := addOffering(tx, uc.ID, o.ID); err != nil {
			return nil, err
		}
		chosen[o.ID] = true
		ids = append(ids, o.ID)
	}
	var removed []uint
	for _, id := range before {
		if !chosen[id] {
			removed = append(removed, id)
		}
	}
	if err := unplaceReviews(tx, uc.ID, removed); err != nil {
		return nil, err
	}
	if err := placeReviews(tx, uc.ID, ids); err != nil {
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return m.GetUserCourse(userID, id)
}

func (m *UserCourseModel) RemoveUserCourseOffering(userID, id, offeringID uint) (*UserCourse, error) {
	uc, err := m.GetUserCourse(userID, id)
	if err != nil {
		return nil, err
	}
	tx := m.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer tx.RollbackUnlessCommitted()
	err = tx.Unscoped().Where("user_course_id = ? AND course_offering_id = ?", uc.ID, offeringID).
		Delete(&UserCourseOffering{}).Error
	if err == nil {
		err = unplaceReviews(tx, uc.ID, []uint{offeringID})
	}
	if err == nil {
		err = tx.Commit().Error
	}
	if err != nil {
		return nil, err
	}
	return m.GetUserCourse(userID, id)
}

func (m *UserCourseModel) DeleteUserCourse(userID, id uint) error {
	tx := m.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer tx.RollbackUnlessCommitted()
	var uc UserCourse
	if err := tx.Where("user_id = ?", userID).First(&uc, id).Error; err != nil {
		return err
	}
	var reviews int
	err := tx.Model(&FactrakSurvey{}).
		Where(`user_course_id = ? OR (user_course_id IS NULL AND user_id = ? AND course_canonical_id = ?
  AND NOT EXISTS (SELECT 1 FROM course_in_semesters cis WHERE cis.id = ? AND cis.strm > ?))`,
			uc.ID, userID, uc.CourseCanonicalID, uc.CourseInSemesterID, StartedStrm(time.Now())).
		Count(&reviews).Error
	if err != nil {
		return err
	}
	if reviews > 0 {
		return ErrUserCourseHasReview
	}
	if err := tx.Unscoped().Where("user_course_id = ?", uc.ID).Delete(&UserCourseOffering{}).Error; err != nil {
		return err
	}
	if err := tx.Unscoped().Delete(&uc).Error; err != nil {
		return err
	}
	return tx.Commit().Error
}

func (m *CourseCatalogModel) CourseOfferingsInTerm(canonicalID uint, strm int, classNbrs []int) (*CourseInSemester, []*CourseOffering, error) {
	var cis CourseInSemester
	if err := m.db.Where("course_canonical_id = ? AND strm = ?", canonicalID, strm).First(&cis).Error; err != nil {
		return nil, nil, err
	}
	if len(classNbrs) == 0 {
		return &cis, nil, nil
	}
	var offerings []*CourseOffering
	err := m.db.Where("course_in_semester_id = ? AND strm = ? AND class_nbr IN (?)", cis.ID, strm, classNbrs).
		Find(&offerings).Error
	if err != nil {
		return nil, nil, err
	}
	found := map[int]bool{}
	for _, o := range offerings {
		found[o.ClassNbr] = true
	}
	for _, n := range classNbrs {
		if !found[n] {
			return nil, nil, ErrUserCourseSection
		}
	}
	return &cis, offerings, nil
}

func (m *UserCourseModel) PlanCourses(userID uint) ([]*UserCourse, error) {
	var out []*UserCourse
	err := m.db.
		Preload("CourseCanonical").
		Preload("CourseCanonical.Listings").
		Preload("CourseCanonical.CoursesInSemester", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, course_canonical_id, strm, year, term, status, source, title, components, attributes, div1, div2, div3, dpe, qfr, wac").
				Where("source = ?", CatalogSourceCatalog).Order("strm DESC")
		}).
		Preload("CourseInSemester").
		Preload("Offerings").
		Preload("Offerings.CourseOffering").
		Where("user_id = ? AND hidden = ?", userID, false).
		Order("id").Find(&out).Error
	return out, err
}
