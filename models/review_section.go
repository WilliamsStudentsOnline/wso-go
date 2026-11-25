package models

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

var (
	ErrReviewSectionNotFound       = errors.New("no offered section with that term and class number")
	ErrReviewSectionNotStarted     = errors.New("section's term hasn't started")
	ErrReviewSectionProfessor      = errors.New("professor did not teach that section")
	ErrReviewSectionNoLegacyCourse = errors.New("no Factrak area for this course's codes")
)

type ReviewSection struct {
	CourseOffering *CourseOffering
	Season         string // FactrakSurveySemesterSeason*
	CalendarYear   int    // Factrak's semester year: Fall 2026 is 2026
	InstructorIDs  []uint // unmatched instructors are left out
	LegacyCourseID uint
}

// Over the summer, the coming fall
func CurrentStrm(now time.Time) int {
	year, term := CurrentCatalogYear(now), 1
	switch m := now.Month(); {
	case m == time.January:
		term = 2
	case m >= time.February && m < time.June:
		term = 3
	}
	return 1000 + (year%100)*10 + term
}

// Over the summer, the spring just ended
func StartedStrm(now time.Time) int {
	strm := CurrentStrm(now)
	if m := now.Month(); m >= time.June && m < time.September {
		return strm - 10 + 2
	}
	return strm
}

// With no matched instructors, accept any professor, since matching can miss people
func (s *ReviewSection) TaughtBy(profID uint) bool {
	if len(s.InstructorIDs) == 0 {
		return true
	}
	for _, id := range s.InstructorIDs {
		if id == profID {
			return true
		}
	}
	return false
}

func (m *CourseCatalogModel) GetReviewSection(strm, classNbr int, now time.Time) (*ReviewSection, error) {
	var o CourseOffering
	err := m.db.Preload("Instructors").
		Where("strm = ? AND class_nbr = ? AND status = ?", strm, classNbr, CourseOfferingStatusOffered).
		First(&o).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, ErrReviewSectionNotFound
		}
		return nil, err
	}
	if o.Strm > StartedStrm(now) {
		return nil, ErrReviewSectionNotStarted
	}
	s := &ReviewSection{CourseOffering: &o, CalendarYear: o.Year}
	switch o.Term {
	case "Fall":
		s.Season, s.CalendarYear = FactrakSurveySemesterSeasonFall, o.Year-1
	case "Winter":
		s.Season = FactrakSurveySemesterSeasonWinterStudy
	default:
		s.Season = FactrakSurveySemesterSeasonSpring
	}
	for _, in := range o.Instructors {
		if in.UserID != nil {
			s.InstructorIDs = append(s.InstructorIDs, *in.UserID)
		}
	}
	if s.LegacyCourseID, err = m.legacyCourseFor(o.CourseCanonicalID, o.Year); err != nil {
		if errors.Is(err, ErrReviewSectionNoLegacyCourse) {
			return s, err // still usable for linking an existing review
		}
		return nil, err
	}
	return s, nil
}

// Prefers the course most existing reviews use, else creates one as Factrak's own form does
func (m *CourseCatalogModel) legacyCourseFor(canonicalID uint, year int) (uint, error) {
	var used []uint
	err := m.db.Model(&FactrakSurvey{}).Where("course_canonical_id = ?", canonicalID).
		Group("course_id").Order("COUNT(*) DESC").Limit(1).Pluck("course_id", &used).Error
	if err != nil {
		return 0, err
	}
	if len(used) > 0 {
		return used[0], nil
	}

	var listings []*CourseListing
	err = m.db.Where("course_canonical_id = ? AND first_valid_year <= ? AND last_valid_year >= ?", canonicalID, year, year).
		Order("id").Find(&listings).Error
	if err != nil {
		return 0, err
	}
	for _, l := range listings {
		var area AreaOfStudy
		err := m.db.Where("abbrev = ?", strings.ToUpper(l.Subject)).First(&area).Error
		if gorm.IsRecordNotFoundError(err) {
			err = areaRenamedTo(m.db, strings.ToUpper(l.Subject), &area)
		}
		if gorm.IsRecordNotFoundError(err) {
			continue
		}
		if err != nil {
			return 0, err
		}
		course := Course{Number: strconv.Itoa(l.Number) + l.Letter, AreaOfStudyID: &area.ID}
		if err := m.db.FirstOrCreate(&course, Course{Number: course.Number, AreaOfStudyID: &area.ID}).Error; err != nil {
			return 0, err
		}
		return course.ID, nil
	}
	return 0, ErrReviewSectionNoLegacyCourse
}

func (m *CourseCatalogModel) LinkSurveyToSection(survey *FactrakSurvey, s *ReviewSection) error {
	tx := m.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer tx.RollbackUnlessCommitted()
	err := tx.Model(survey).Updates(map[string]interface{}{
		"course_canonical_id":             s.CourseOffering.CourseCanonicalID,
		"course_offering_id":              s.CourseOffering.ID,
		"course_link_verification_source": CourseLinkVerificationSourceAuthorConfirmed,
		"semester_season":                 s.Season,
		"semester_year":                   s.CalendarYear,
	}).Error
	if err != nil {
		return err
	}
	if s.CourseOffering.CourseInSemesterID != nil {
		var cis CourseInSemester
		if err := tx.First(&cis, *s.CourseOffering.CourseInSemesterID).Error; err != nil {
			return err
		}
		mine := &UserCourseModel{db: tx}
		uc, err := mine.addInTx(survey.UserID, s.CourseOffering.CourseCanonicalID, &cis, []*CourseOffering{s.CourseOffering})
		if err != nil {
			return err
		}
		if err := tx.Model(survey).Update("user_course_id", uc.ID).Error; err != nil {
			return err
		}
	}
	return tx.Commit().Error
}

func (m *CourseCatalogModel) ListUserSurveys(userID uint) ([]*FactrakSurvey, error) {
	var out []*FactrakSurvey
	err := m.db.Preload("Professor").Preload("Course").Preload("Course.AreaOfStudy").
		Preload("CourseCanonical").Preload("CourseOffering").Preload("CourseOffering.CourseInSemester").Preload("CourseOffering.Meetings").
		Where("user_id = ?", userID).Order("created_at DESC, id DESC").Find(&out).Error
	return out, err
}

func (m *CourseCatalogModel) SectionsTaughtBy(canonicalID, profID uint, now time.Time, limit int) ([]*CourseOffering, error) {
	var out []*CourseOffering
	err := m.db.Preload("CourseInSemester").Preload("Meetings").
		Joins("JOIN course_offering_instructors oi ON oi.course_offering_id = course_offerings.id AND oi.deleted_at IS NULL").
		Where("course_offerings.course_canonical_id = ? AND oi.user_id = ? AND course_offerings.status = ? AND course_offerings.strm <= ?",
			canonicalID, profID, CourseOfferingStatusOffered, StartedStrm(now)).
		Order("course_offerings.strm DESC, course_offerings.class_nbr").Limit(limit).Find(&out).Error
	return out, err
}
