package models

import (
	"database/sql"
	"regexp"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type CourseCatalogModel struct {
	db  *gorm.DB
	log *zap.SugaredLogger
}

func NewCourseCatalogModel(db *gorm.DB, log *zap.SugaredLogger) *CourseCatalogModel {
	return &CourseCatalogModel{db: db, log: log}
}

type CatalogTerm struct {
	Strm        int
	Year        int
	Term        string
	CourseCount int
}

// Terms with only recovered (last_offered) history are too partial to browse
func (m *CourseCatalogModel) ListTerms() ([]CatalogTerm, error) {
	var terms []CatalogTerm
	err := m.db.Model(&CourseInSemester{}).
		Select("strm, year, term, COUNT(*) AS course_count").
		Where("source = ?", CatalogSourceCatalog).
		Group("strm, year, term").
		Order("strm DESC").
		Scan(&terms).Error
	return terms, err
}

func (m *CourseCatalogModel) TermVersion(strm int) (time.Time, error) {
	var version time.Time
	for _, table := range []string{"course_in_semesters", "course_offerings"} {
		var t sql.NullTime
		if err := m.db.Table(table).Where("strm = ?", strm).Select("MAX(updated_at)").Row().Scan(&t); err != nil {
			return time.Time{}, err
		}
		if t.Valid && t.Time.After(version) {
			version = t.Time
		}
	}
	return version, nil
}

func (m *CourseCatalogModel) withSections(db *gorm.DB) *gorm.DB {
	return db.
		Preload("CourseCanonical").
		Preload("CourseCanonical.Listings").
		Preload("Offerings", func(db *gorm.DB) *gorm.DB { return db.Order("section, class_nbr") }).
		Preload("Offerings.Meetings").
		Preload("Offerings.Instructors")
}

func (m *CourseCatalogModel) TermCourses(strm int) ([]*CourseInSemester, error) {
	var courses []*CourseInSemester
	err := m.withSections(m.db).
		Where("strm = ? AND source = ?", strm, CatalogSourceCatalog).
		Order("title").
		Find(&courses).Error
	return courses, err
}

// Includes recovered history, so links to old terms still resolve
func (m *CourseCatalogModel) GetCourseInSemester(strm int, crseID string) (*CourseInSemester, error) {
	var cis CourseInSemester
	err := m.withSections(m.db).
		Joins("JOIN course_canonicals ON course_canonicals.id = course_in_semesters.course_canonical_id AND course_canonicals.deleted_at IS NULL").
		Where("course_in_semesters.strm = ? AND course_canonicals.crse_id = ?", strm, crseID).
		First(&cis).Error
	if err != nil {
		return nil, err
	}
	return &cis, nil
}

func (m *CourseCatalogModel) GetCanonical(crseID string) (*CourseCanonical, error) {
	var course CourseCanonical
	err := m.db.
		Preload("Listings").
		Preload("CoursesInSemester", func(db *gorm.DB) *gorm.DB { return db.Order("strm DESC") }).
		Preload("CoursesInSemester.Offerings", func(db *gorm.DB) *gorm.DB { return db.Order("section, class_nbr") }).
		Preload("CoursesInSemester.Offerings.Instructors").
		Where("crse_id = ?", crseID).
		First(&course).Error
	if err != nil {
		return nil, err
	}
	return &course, nil
}

// Any option within a group, every group must match
type CatalogSearchOptions struct {
	Query         string
	Subjects      []string
	Levels        []int    // 0 (winter study), 1..4 (100..400)
	Divisions     []string // div1, div2, div3, in any term
	Distributions []string // dpe, qfr, wac, in any term
	OfferedSince  int      // catalog year (0 = any)
	Limit         int
	Offset        int
}

var catalogFlagColumns = map[string]bool{"div1": true, "div2": true, "div3": true, "dpe": true, "qfr": true, "wac": true}

func (m *CourseCatalogModel) SearchCanonical(opts CatalogSearchOptions) ([]*CourseCanonical, int, error) {
	q := m.db.Model(&CourseCanonical{}).Where("course_canonicals.deleted_at IS NULL")

	var subjects []string
	for _, s := range opts.Subjects {
		if s = strings.ToUpper(strings.TrimSpace(s)); s != "" {
			subjects = append(subjects, s)
		}
	}
	if len(subjects) > 0 {
		q = q.Where("EXISTS (SELECT 1 FROM course_listings cl WHERE cl.course_canonical_id = course_canonicals.id AND cl.deleted_at IS NULL AND cl.subject IN (?))", subjects)
	}
	if query := strings.TrimSpace(opts.Query); query != "" {
		q = whereQuery(q, query)
	}
	if len(opts.Levels) > 0 {
		q = q.Where("EXISTS (SELECT 1 FROM course_listings cl WHERE cl.course_canonical_id = course_canonicals.id AND cl.deleted_at IS NULL AND FLOOR(cl.number / 100) IN (?))", opts.Levels)
	}
	for _, group := range [][]string{opts.Divisions, opts.Distributions} {
		if cond := flagCondition(group); cond != "" {
			q = q.Where("EXISTS (SELECT 1 FROM course_in_semesters cis WHERE cis.course_canonical_id = course_canonicals.id AND cis.deleted_at IS NULL AND (" + cond + "))")
		}
	}
	if opts.OfferedSince > 0 {
		q = q.Where("EXISTS (SELECT 1 FROM course_in_semesters cis WHERE cis.course_canonical_id = course_canonicals.id AND cis.deleted_at IS NULL AND cis.status = ? AND cis.year >= ?)",
			CourseOfferingStatusOffered, opts.OfferedSince)
	}

	var total int
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	var courses []*CourseCanonical
	order := "course_canonicals.title"
	if query := strings.TrimSpace(opts.Query); query != "" {
		rank, args := queryRank(query)
		q = q.Select("course_canonicals.*, "+rank+" AS search_rank", args...)
		order = "search_rank DESC, course_canonicals.title"
	}
	err := q.
		Preload("Listings").
		Preload("CoursesInSemester", func(db *gorm.DB) *gorm.DB { return db.Order("strm DESC") }).
		Order(order).
		Limit(limit).
		Offset(opts.Offset).
		Find(&courses).Error
	return courses, total, err
}

func whereQuery(q *gorm.DB, query string) *gorm.DB {
	var conds []string
	var args []interface{}
	for _, word := range strings.Fields(query) {
		like := "%" + word + "%"
		conds = append(conds, `(course_canonicals.title LIKE ? OR course_canonicals.description LIKE ?
  OR EXISTS (SELECT 1 FROM course_listings cl WHERE cl.course_canonical_id = course_canonicals.id AND cl.deleted_at IS NULL
    AND CONCAT(cl.subject, ' ', cl.number, cl.letter) LIKE ?)
  OR EXISTS (SELECT 1 FROM course_offerings o JOIN course_offering_instructors oi ON oi.course_offering_id = o.id AND oi.deleted_at IS NULL
    WHERE o.course_canonical_id = course_canonicals.id AND o.deleted_at IS NULL AND oi.name_as_listed LIKE ?))`)
		args = append(args, like, like, like, like)
	}
	cond := strings.Join(conds, " AND ")
	if subject, number, ok := parseCodeQuery(query); ok {
		cond = `EXISTS (SELECT 1 FROM course_listings cl WHERE cl.course_canonical_id = course_canonicals.id AND cl.deleted_at IS NULL
  AND cl.subject = ? AND CAST(cl.number AS CHAR) LIKE ?) OR (` + cond + `)`
		args = append([]interface{}{subject, number + "%"}, args...)
	}
	return q.Where(cond, args...)
}

var rankStopWords = map[string]bool{"a": true, "an": true, "and": true, "&": true, "the": true,
	"of": true, "in": true, "on": true, "to": true, "for": true, "with": true}

// Short words must be whole title words or title initials, so "ai" isn't "AIDS" or "Pain"
func queryRank(query string) (string, []interface{}) {
	var terms []string
	var args []interface{}
	for _, word := range strings.Fields(query) {
		if rankStopWords[strings.ToLower(word)] {
			continue
		}
		start := `\b` + regexp.QuoteMeta(word)
		if len(word) <= 3 {
			start += `\b`
		}
		terms = append(terms, "4 * (COALESCE(course_canonicals.title, '') REGEXP ?)")
		args = append(args, start)
		if acronym := acronymPattern(word); acronym != "" {
			terms = append(terms, "4 * (COALESCE(course_canonicals.title, '') REGEXP ?)")
			args = append(args, acronym)
		}
		if len(word) >= 4 {
			terms = append(terms, "(course_canonicals.title LIKE ?)")
			args = append(args, "%"+word+"%")
		}
		terms = append(terms, "(COALESCE(course_canonicals.description, '') REGEXP ?)")
		args = append(args, start)
	}
	rank := "0"
	if len(terms) > 0 {
		rank = "(" + strings.Join(terms, " + ") + ")"
	}
	if subject, number, ok := parseCodeQuery(query); ok {
		rank = `(CASE WHEN EXISTS (SELECT 1 FROM course_listings cl WHERE cl.course_canonical_id = course_canonicals.id
  AND cl.deleted_at IS NULL AND cl.subject = ? AND CAST(cl.number AS CHAR) LIKE ?) THEN 1000 ELSE ` + rank + ` END)`
		args = append([]interface{}{subject, number + "%"}, args...)
	}
	return rank, args
}

// "ai" matches "Artificial Intelligence", not "and Immortals"
func acronymPattern(word string) string {
	if len(word) < 2 || len(word) > 4 {
		return ""
	}
	parts := make([]string, 0, len(word))
	for _, r := range strings.ToLower(word) {
		if r < 'a' || r > 'z' {
			return ""
		}
		parts = append(parts, `\b`+string(r)+`\w{3,}`)
	}
	return strings.Join(parts, `\W+`)
}

func flagCondition(flags []string) string {
	var cols []string
	for _, f := range flags {
		if catalogFlagColumns[f] {
			cols = append(cols, "cis."+f)
		}
	}
	return strings.Join(cols, " OR ")
}

func parseCodeQuery(q string) (subject, number string, ok bool) {
	q = strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(q), " ", ""))
	i := 0
	for i < len(q) && q[i] >= 'A' && q[i] <= 'Z' {
		i++
	}
	if i == 0 || i == len(q) {
		return "", "", false
	}
	for _, r := range q[i:] {
		if r < '0' || r > '9' {
			return "", "", false
		}
	}
	return q[:i], q[i:], true
}

// Keep in step with jobs/catalog_update ParseCatalog
func GradingBasisOptions(basis string) (desc string, passFail, fifthCourse bool) {
	switch strings.TrimSpace(basis) {
	case "WPP":
		return "Winter Study", false, false
	case "GRD":
		return "No Pass/Fail and No Fifth Course", false, false
	case "OPT":
		return "Pass/Fail Available, Fifth Course Available", true, true
	case "OPX":
		return "Pass/Fail Unavailable, Fifth Course Available", false, true
	case "OPP":
		return "Pass/Fail Available, Fifth Course Unavailable", true, false
	case "PNP":
		return "Pass/Fail Option Only", true, true
	case "PF4":
		return "Pass/Fail Option Only", true, false
	case "NON":
		return "Non-Graded", false, false
	default:
		return "", false, false
	}
}

func (m *CourseCatalogModel) ListCourseAttributes() ([]*CourseAttribute, error) {
	var attrs []*CourseAttribute
	err := m.db.Preload("AreaOfStudy").Order("code").Find(&attrs).Error
	return attrs, err
}

func (m *CourseCatalogModel) ListAreasOfStudy() ([]*AreaOfStudy, []*CourseAttribute, error) {
	var areas []*AreaOfStudy
	if err := m.db.Order("name").Find(&areas).Error; err != nil {
		return nil, nil, err
	}
	var attrs []*CourseAttribute
	if err := m.db.Where("area_of_study_id IS NOT NULL").Order("code").Find(&attrs).Error; err != nil {
		return nil, nil, err
	}
	return areas, attrs, nil
}

// Factrak area abbrev → the catalog subject that replaced it
var RenamedSubjects = map[string]string{"WGST": "WGSS", "ASST": "ASIA"}

func areaRenamedTo(db *gorm.DB, subject string, area *AreaOfStudy) error {
	var abbrevs []string
	for abbrev, s := range RenamedSubjects {
		if s == subject {
			abbrevs = append(abbrevs, abbrev)
		}
	}
	if len(abbrevs) == 0 {
		return gorm.ErrRecordNotFound
	}
	return db.Where("abbrev IN (?)", abbrevs).First(area).Error
}

type CourseFactrakSummary struct {
	CrseID            string
	Reviews           int
	Recommends        int
	CurrentReviews    int
	CurrentRecommends int
}

// Follows Factrak's own visibility rules (see FactrakSurveyModel scopeCurrent and scopeProfAtWilliams)
func (m *CourseCatalogModel) TermFactrakSummaries(strm int, includeMismatch bool) ([]CourseFactrakSummary, error) {
	sourceFilter := ""
	if !includeMismatch {
		sourceFilter = " AND (fs.course_link_verification_source IS NULL OR fs.course_link_verification_source <> '" + CourseLinkVerificationSourceMismatch + "')"
	}
	var out []CourseFactrakSummary
	err := m.db.Raw(`
SELECT c.crse_id,
  COUNT(fs.would_recommend_course) AS reviews,
  COALESCE(SUM(fs.would_recommend_course), 0) AS recommends,
  COUNT(CASE WHEN cur.user_id IS NOT NULL THEN fs.would_recommend_course END) AS current_reviews,
  COALESCE(SUM(CASE WHEN cur.user_id IS NOT NULL THEN fs.would_recommend_course END), 0) AS current_recommends
FROM course_in_semesters cis
JOIN course_canonicals c ON c.id = cis.course_canonical_id AND c.deleted_at IS NULL
JOIN factrak_surveys fs ON fs.course_canonical_id = cis.course_canonical_id AND fs.deleted_at IS NULL`+sourceFilter+`
  AND fs.created_at >= ?
JOIN users prof ON prof.id = fs.professor_id AND (prof.at_williams = 1 OR `+teachingThisYearSQL("prof.id")+`)
LEFT JOIN (
  SELECT DISTINCT o.course_in_semester_id, oi.user_id
  FROM course_offerings o
  JOIN course_offering_instructors oi ON oi.course_offering_id = o.id AND oi.deleted_at IS NULL
  WHERE o.strm = ? AND o.deleted_at IS NULL AND oi.user_id IS NOT NULL
) cur ON cur.course_in_semester_id = cis.id AND cur.user_id = fs.professor_id
WHERE cis.strm = ? AND cis.source = ? AND cis.deleted_at IS NULL
GROUP BY c.crse_id
HAVING reviews > 0`, time.Now().AddDate(-5, 0, 0), CurrentCatalogYear(time.Now()), CourseOfferingStatusOffered, strm, strm, CatalogSourceCatalog).Scan(&out).Error
	return out, err
}

// The lowest class number in the latest term is the primary cross-listing
func (m *CourseCatalogModel) CourseArea(canonicalID uint) (*AreaOfStudy, error) {
	var subjects []string
	err := m.db.Model(&CourseOffering{}).Where("course_canonical_id = ? AND subject <> ''", canonicalID).
		Order("strm DESC, class_nbr").Limit(1).Pluck("subject", &subjects).Error
	if err != nil {
		return nil, err
	}
	if len(subjects) == 0 {
		err = m.db.Model(&CourseListing{}).Where("course_canonical_id = ?", canonicalID).
			Order("last_valid_year DESC, subject").Limit(1).Pluck("subject", &subjects).Error
		if err != nil || len(subjects) == 0 {
			return nil, err
		}
	}
	subject := strings.ToUpper(subjects[0])
	var area AreaOfStudy
	err = m.db.Where("abbrev = ?", subject).First(&area).Error
	if gorm.IsRecordNotFoundError(err) {
		err = areaRenamedTo(m.db, subject, &area)
	}
	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &area, nil
}

func (m *CourseCatalogModel) CanonicalID(crseID string) (uint, error) {
	var course CourseCanonical
	err := m.db.Select("id").Where("crse_id = ?", crseID).First(&course).Error
	return course.ID, err
}

func (m *CourseCatalogModel) CanonicalCoursesByID(ids []uint) (map[uint]*CourseCanonical, error) {
	var courses []*CourseCanonical
	if err := m.db.Select("id, crse_id, title").Where("id IN (?)", ids).Find(&courses).Error; err != nil {
		return nil, err
	}
	out := make(map[uint]*CourseCanonical, len(courses))
	for _, c := range courses {
		out[c.ID] = c
	}
	return out, nil
}

func (m *CourseCatalogModel) LatestCoursesWithAttribute(attr string, since, limit int) ([]*CourseInSemester, error) {
	latest := m.db.Model(&CourseInSemester{}).Select("course_canonical_id, MAX(strm)").
		Where("source = ? AND year >= ? AND FIND_IN_SET(?, attributes) > 0", CatalogSourceCatalog, since, attr).
		Group("course_canonical_id").SubQuery()
	q := m.db.Model(&CourseInSemester{}).Where("source = ? AND (course_canonical_id, strm) IN ?", CatalogSourceCatalog, latest)
	var total int
	if err := q.Count(&total).Error; err != nil || total == 0 || total > limit {
		return nil, err
	}
	var courses []*CourseInSemester
	err := m.withSections(q).Find(&courses).Error
	return courses, err
}
