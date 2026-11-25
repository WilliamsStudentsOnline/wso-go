package models

import (
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

type ProfessorCourse struct {
	CourseCanonicalID uint
	CrseID            string
	Title             string
	LastTaughtStrm    int // 0 when only reviewed (e.g. before the catalog starts)
	TermsTaught       int
	TaughtStrmList    string // comma-separated term codes, newest first
	TaughtStrms       []int  `gorm:"-"`
	Reviews           int
	Listings          []*CourseListing
}

// Catalog instructors, whatever users.type says (#160)
func scopeCatalogProfessor(db *gorm.DB) *gorm.DB {
	return db.Where(`users.type = ? OR EXISTS (SELECT 1 FROM course_offering_instructors oip
  WHERE oip.user_id = users.id AND oip.deleted_at IS NULL)`, UserTypeProfessor)
}

func (m *CourseCatalogModel) SearchProfessors(query string, limit, offset int) ([]*User, int, error) {
	q := whereNameWords(m.db.Model(&User{}).Scopes(scopeCatalogProfessor), query)
	return professorPage(q, "users.at_williams DESC, users.name", limit, offset)
}

func (m *CourseCatalogModel) AreaProfessors(subjects []string, fromYear, toYear int, query string, limit, offset int) ([]*User, int, error) {
	q := m.db.Model(&User{}).Scopes(scopeCatalogProfessor).Where(`EXISTS (SELECT 1 FROM course_offering_instructors oi
  JOIN course_offerings o ON o.id = oi.course_offering_id AND o.deleted_at IS NULL
  JOIN course_listings cl ON cl.course_canonical_id = o.course_canonical_id AND cl.deleted_at IS NULL
    AND o.year BETWEEN cl.first_valid_year AND cl.last_valid_year
  WHERE oi.user_id = users.id AND oi.deleted_at IS NULL AND o.status = ? AND o.year BETWEEN ? AND ?
    AND cl.subject IN (?))`, CourseOfferingStatusOffered, fromYear, toYear, subjects)
	return professorPage(whereNameWords(q, query), "users.name", limit, offset)
}

func whereNameWords(q *gorm.DB, query string) *gorm.DB {
	for _, word := range strings.Fields(query) {
		q = q.Where("users.name LIKE ?", "%"+word+"%")
	}
	return q
}

func professorPage(q *gorm.DB, order string, limit, offset int) ([]*User, int, error) {
	var total int
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	var profs []*User
	err := q.Preload("Department").Order(order).
		Limit(limit).Offset(offset).Find(&profs).Error
	return profs, total, err
}

func (m *CourseCatalogModel) GetProfessor(userID uint) (*User, error) {
	var prof User
	err := m.db.Preload("Department").Preload("Office").Scopes(scopeCatalogProfessor).First(&prof, userID).Error
	if err != nil {
		return nil, err
	}
	return &prof, nil
}

func (m *CourseCatalogModel) ProfessorsTeachingInYear(userIDs []uint, year int) (map[uint]bool, error) {
	out := map[uint]bool{}
	if len(userIDs) == 0 {
		return out, nil
	}
	var ids []uint
	err := m.db.Table("course_offering_instructors oi").
		Joins("JOIN course_offerings o ON o.id = oi.course_offering_id AND o.deleted_at IS NULL").
		Where("oi.deleted_at IS NULL AND oi.user_id IN (?) AND o.year = ? AND o.status = ?", userIDs, year, CourseOfferingStatusOffered).
		Pluck("DISTINCT oi.user_id", &ids).Error
	for _, id := range ids {
		out[id] = true
	}
	return out, err
}

func (m *CourseCatalogModel) ProfessorCourses(userID uint, reviewsSince time.Time) ([]*ProfessorCourse, error) {
	var courses []*ProfessorCourse
	err := m.db.Raw(`
SELECT c.id AS course_canonical_id, c.crse_id, c.title,
  COALESCE(t.last_strm, 0) AS last_taught_strm, COALESCE(t.terms, 0) AS terms_taught,
  COALESCE(t.strms, '') AS taught_strm_list, COALESCE(r.n, 0) AS reviews
FROM course_canonicals c
LEFT JOIN (
  SELECT o.course_canonical_id AS cid, MAX(o.strm) AS last_strm, COUNT(DISTINCT o.strm) AS terms,
    GROUP_CONCAT(DISTINCT o.strm ORDER BY o.strm DESC) AS strms
  FROM course_offerings o JOIN course_offering_instructors oi ON oi.course_offering_id = o.id AND oi.deleted_at IS NULL
  WHERE oi.user_id = ? AND o.deleted_at IS NULL AND o.status = ?
  GROUP BY o.course_canonical_id
) t ON t.cid = c.id
LEFT JOIN (
  SELECT course_canonical_id AS cid, COUNT(*) AS n FROM factrak_surveys
  WHERE professor_id = ? AND deleted_at IS NULL AND created_at >= ?
    AND (course_link_verification_source IS NULL OR course_link_verification_source <> ?)
  GROUP BY course_canonical_id
) r ON r.cid = c.id
WHERE c.deleted_at IS NULL AND (t.cid IS NOT NULL OR r.cid IS NOT NULL)
ORDER BY last_taught_strm DESC, c.title`,
		userID, CourseOfferingStatusOffered, userID, reviewsSince, CourseLinkVerificationSourceMismatch).Scan(&courses).Error
	if err != nil || len(courses) == 0 {
		return courses, err
	}

	ids := make([]uint, 0, len(courses))
	byID := map[uint]*ProfessorCourse{}
	for _, c := range courses {
		for _, part := range strings.Split(c.TaughtStrmList, ",") {
			if strm, err := strconv.Atoi(strings.TrimSpace(part)); err == nil {
				c.TaughtStrms = append(c.TaughtStrms, strm)
			}
		}
		ids = append(ids, c.CourseCanonicalID)
		byID[c.CourseCanonicalID] = c
	}
	var listings []*CourseListing
	if err := m.db.Where("course_canonical_id IN (?)", ids).Order("subject, number").Find(&listings).Error; err != nil {
		return nil, err
	}
	for _, l := range listings {
		byID[l.CourseCanonicalID].Listings = append(byID[l.CourseCanonicalID].Listings, l)
	}
	return courses, nil
}

// Must match Factrak's own professor ratings
func (m *FactrakSurveyModel) GetSurveyRatingsForProfessor(profID uint, canonicalCourseID *uint, ratings *FactrakSurveyAvgRatings) error {
	scopes := []func(db *gorm.DB) *gorm.DB{m.scopeCurrent, m.withProfessorID(profID)}
	if canonicalCourseID != nil {
		id := *canonicalCourseID
		scopes = append(scopes, func(db *gorm.DB) *gorm.DB {
			return db.Where("factrak_surveys.course_canonical_id = ?", id).
				Where("(factrak_surveys.course_link_verification_source IS NULL OR factrak_surveys.course_link_verification_source <> ?)", CourseLinkVerificationSourceMismatch)
		})
	}
	return m.getSurveyRatings(ratings, scopes...)
}

func FactrakReviewWindow(now time.Time) time.Time {
	return now.AddDate(-5, 0, 0)
}
