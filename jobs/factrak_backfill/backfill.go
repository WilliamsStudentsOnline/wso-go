package factrak_backfill

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

const (
	MatchExactListing  = "exact_listing"
	MatchDisambiguated = "disambiguated_offering"
	MatchAmbiguous     = "ambiguous"
	MatchUnmatched     = "unmatched"
	MatchAlreadyLinked = "already_linked"
)

// Methods record which rule produced a match, so each rule can be audited or undone.
const (
	MethodListing             = "listing"               // subject+number+letter found as-is
	MethodNormalizedCode      = "normalized_code"       // junk suffix dropped ("101LAB", "201-02") or "140/204" split
	MethodTutorialSuffix      = "tutorial_suffix"       // Factrak "127T", catalog lists "127"
	MethodLetteredSection     = "lettered_section"      // Factrak "301", catalog only has "301A".."301M"
	MethodProfessorSameNumber = "professor_same_number" // wrong subject picked; professor taught that number that term
	MethodClassNumber         = "class_number"          // student typed the registration class number ("1804")
	MethodManualOverride      = "manual_override"       // reviewed decision from overrides.json
)

// Result is one survey's backfill outcome.
type Result struct {
	SurveyID          uint   `json:"surveyID"`
	FactrakCourseID   uint   `json:"factrakCourseID"`
	Abbrev            string `json:"abbrev"`
	CatalogSubject    string `json:"catalogSubject"`
	Number            int    `json:"number"`
	Letter            string `json:"letter"`
	RawNumber         string `json:"rawNumber,omitempty"`
	ProfessorID       uint   `json:"professorID"`
	SemesterSeason    string `json:"semesterSeason,omitempty"`
	SemesterYear      *int   `json:"semesterYear,omitempty"`
	Status            string `json:"status"`
	Method            string `json:"method,omitempty"`
	CanonicalCourseID *uint  `json:"canonicalCourseID,omitempty"`
	OfferingID        *uint  `json:"offeringID,omitempty"`
	CandidateCount    int    `json:"candidateCount"`
	Note              string `json:"note,omitempty"`
}

// Report summarizes a dry-run or apply pass.
type Report struct {
	Total         int `json:"total"`
	ExactListing  int `json:"exactListing"`
	Disambiguated int `json:"disambiguated"`
	Ambiguous     int `json:"ambiguous"`
	Unmatched     int `json:"unmatched"`
	AlreadyLinked int `json:"alreadyLinked"`
	Applied       int `json:"applied"`
	// OverridesSkipped counts overrides that didn't fit (filed course changed or CRSE_ID missing).
	OverridesSkipped int               `json:"overridesSkipped"`
	ByMethod         map[string]int    `json:"byMethod"`
	Results          []Result          `json:"results"`
	AliasMap         map[string]string `json:"aliasMap"`
}

// Options controls backfill behavior.
type Options struct {
	DryRun         bool
	SkipLinked     bool       // skip surveys that already have canonical_course_id
	IncludeResults bool       // keep per-survey rows in report (can be large)
	Overrides      []Override // reviewed survey → course decisions; nil uses the embedded overrides.json
}

// CourseCode is one parsed Factrak course number.
type CourseCode struct {
	Number     int
	Letter     string
	Normalized bool // a non-letter suffix was dropped
}

var leadingDigits = regexp.MustCompile(`^(\d+)(.*)$`)
var singleLetter = regexp.MustCompile(`^[A-Z]$`)

// ParseCourseCodes reads Factrak course numbers, which students typed by hand.
// "136" → 136; "301B" → 301/B; "127T" → 127/T; "101LAB" and "201-02" → 101 and 201
// (suffix dropped); "140/204" → both 140 and 204.
func ParseCourseCodes(raw string) ([]CourseCode, error) {
	raw = strings.ToUpper(strings.TrimSpace(raw))
	if raw == "" {
		return nil, fmt.Errorf("empty course number")
	}
	var codes []CourseCode
	for i, part := range strings.Split(raw, "/") {
		m := leadingDigits.FindStringSubmatch(strings.TrimSpace(part))
		if m == nil {
			if i == 0 {
				break // "CHIN 101/102": not a course number
			}
			continue
		}
		n, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}
		suffix := strings.TrimSpace(m[2])
		code := CourseCode{Number: n}
		switch {
		case suffix == "":
		case singleLetter.MatchString(suffix):
			code.Letter = suffix
		default:
			code.Normalized = true
		}
		codes = append(codes, code)
	}
	if len(codes) == 0 {
		return nil, fmt.Errorf("course number has no digits: %q", raw)
	}
	return codes, nil
}

// ParseCourseCode returns the first code from ParseCourseCodes.
func ParseCourseCode(raw string) (number int, letter string, err error) {
	codes, err := ParseCourseCodes(raw)
	if err != nil {
		return 0, "", err
	}
	return codes[0].Number, codes[0].Letter, nil
}

// CatalogYearTerm maps Factrak semester season+year to catalog offering year/term.
// Factrak semesterYear is the calendar year of the semester (Fall 2026 → year 2026).
// Catalog year is the spring calendar year of the AY (Fall 2026 → AY 2026-27 → 2027).
func CatalogYearTerm(season string, semesterYear int) (year int, term string, ok bool) {
	switch strings.ToLower(strings.TrimSpace(season)) {
	case models.FactrakSurveySemesterSeasonFall:
		return semesterYear + 1, "Fall", true
	case models.FactrakSurveySemesterSeasonWinterStudy, "winter":
		return semesterYear, "Winter", true
	case models.FactrakSurveySemesterSeasonSpring:
		return semesterYear, "Spring", true
	default:
		return 0, "", false
	}
}

// LoadAliasMap returns factrak_abbrev → catalog_subject (uppercase).
func LoadAliasMap(db *gorm.DB) (map[string]string, error) {
	var aliases []models.CourseSubjectAlias
	if err := db.Find(&aliases).Error; err != nil {
		return nil, err
	}
	m := make(map[string]string, len(aliases))
	for _, a := range aliases {
		m[strings.ToUpper(a.FactrakAbbrev)] = strings.ToUpper(a.CatalogSubject)
	}
	return m, nil
}

// ResolveSubject applies alias map; always returns uppercase.
func ResolveSubject(abbrev string, aliases map[string]string) string {
	abbrev = strings.ToUpper(strings.TrimSpace(abbrev))
	if mapped, ok := aliases[abbrev]; ok {
		return mapped
	}
	return abbrev
}

// RunBackfill maps Factrak surveys onto courses_canonical / offerings.
func RunBackfill(db *gorm.DB, log *zap.SugaredLogger, opts Options) (*Report, error) {
	if log == nil {
		log = zap.NewNop().Sugar()
	}
	aliases, err := LoadAliasMap(db)
	if err != nil {
		return nil, err
	}
	overrideList := opts.Overrides
	if overrideList == nil {
		if overrideList, err = EmbeddedOverrides(); err != nil {
			return nil, err
		}
	}
	overrides := make(map[uint]Override, len(overrideList))
	for _, o := range overrideList {
		overrides[o.SurveyID] = o
	}

	report := &Report{
		AliasMap: aliases,
		ByMethod: map[string]int{},
		Results:  []Result{},
	}

	q := db.Preload("Course").Preload("Course.AreaOfStudy").
		Where("factrak_surveys.deleted_at IS NULL")
	if opts.SkipLinked {
		q = q.Where("canonical_course_id IS NULL")
	}

	var surveys []models.FactrakSurvey
	if err := q.Find(&surveys).Error; err != nil {
		return nil, err
	}
	report.Total = len(surveys)

	for i := range surveys {
		var res Result
		applied := false
		if o, ok := overrides[surveys[i].ID]; ok && surveys[i].CanonicalCourseID == nil {
			res, applied = applyOverride(db, &surveys[i], o)
			if !applied {
				report.OverridesSkipped++
			}
		}
		if !applied {
			res = matchSurvey(db, &surveys[i], aliases)
		}
		switch res.Status {
		case MatchExactListing:
			report.ExactListing++
		case MatchDisambiguated:
			report.Disambiguated++
		case MatchAmbiguous:
			report.Ambiguous++
		case MatchUnmatched:
			report.Unmatched++
		case MatchAlreadyLinked:
			report.AlreadyLinked++
		}
		if res.Method != "" {
			report.ByMethod[res.Method]++
		}

		if !opts.DryRun && (res.Status == MatchExactListing || res.Status == MatchDisambiguated) {
			updates := map[string]interface{}{
				"canonical_course_id": res.CanonicalCourseID,
			}
			if res.OfferingID != nil {
				updates["offering_id"] = res.OfferingID
			}
			if err := db.Model(&models.FactrakSurvey{}).Where("id = ?", res.SurveyID).Updates(updates).Error; err != nil {
				return report, err
			}
			report.Applied++
		}

		if opts.IncludeResults || res.Status == MatchUnmatched || res.Status == MatchAmbiguous {
			report.Results = append(report.Results, res)
		}
	}

	log.Infof("Factrak backfill: total=%d exact=%d disambiguated=%d ambiguous=%d unmatched=%d already=%d applied=%d byMethod=%v dryRun=%v",
		report.Total, report.ExactListing, report.Disambiguated, report.Ambiguous, report.Unmatched, report.AlreadyLinked, report.Applied, report.ByMethod, opts.DryRun)
	return report, nil
}

func newResult(survey *models.FactrakSurvey) Result {
	res := Result{
		SurveyID:        survey.ID,
		FactrakCourseID: survey.CourseID,
		ProfessorID:     survey.ProfessorID,
		Status:          MatchUnmatched,
	}
	if survey.SemesterSeason != nil {
		res.SemesterSeason = *survey.SemesterSeason
	}
	res.SemesterYear = survey.SemesterYear
	return res
}

func matchSurvey(db *gorm.DB, survey *models.FactrakSurvey, aliases map[string]string) Result {
	res := newResult(survey)

	if survey.CanonicalCourseID != nil {
		res.Status = MatchAlreadyLinked
		res.CanonicalCourseID = survey.CanonicalCourseID
		res.OfferingID = survey.OfferingID
		return res
	}

	if survey.Course == nil || survey.Course.AreaOfStudy == nil {
		res.Note = "missing course or area of study"
		return res
	}

	res.Abbrev = strings.ToUpper(survey.Course.AreaOfStudy.Abbreviation)
	res.CatalogSubject = ResolveSubject(res.Abbrev, aliases)
	res.RawNumber = survey.Course.Number

	codes, err := ParseCourseCodes(survey.Course.Number)
	if err != nil {
		res.Note = err.Error()
		return res
	}
	res.Number = codes[0].Number
	res.Letter = codes[0].Letter

	subjects := []string{res.CatalogSubject}
	if res.Abbrev != res.CatalogSubject {
		subjects = append(subjects, res.Abbrev)
	}

	listings, method, subject, err := findListings(db, subjects, codes)
	if err != nil {
		res.Note = err.Error()
		return res
	}

	if len(listings) == 0 {
		// Students sometimes picked the wrong department (PHYS for PSYC, STS for STAT).
		// Accept the number only if this professor taught exactly one course with it that term.
		numbers := make([]int, 0, len(codes))
		for _, c := range codes {
			numbers = append(numbers, c.Number)
		}
		// Some students typed the class number from registration instead of the course number.
		// Class numbers can repeat across terms, so require the survey's term and professor.
		if offering := matchByClassNumber(db, numbers, survey.ProfessorID, survey.SemesterSeason, survey.SemesterYear); offering != nil {
			res.CanonicalCourseID = &offering.CourseCanonicalID
			res.OfferingID = &offering.ID
			res.Status = MatchDisambiguated
			res.Method = MethodClassNumber
			res.CandidateCount = 1
			return res
		}
		canonicalID, offeringID := matchByProfessorNumber(db, numbers, survey.ProfessorID, survey.SemesterSeason, survey.SemesterYear)
		if canonicalID != nil {
			res.CanonicalCourseID = canonicalID
			res.OfferingID = offeringID
			res.Status = MatchDisambiguated
			res.Method = MethodProfessorSameNumber
			res.CandidateCount = 1
			return res
		}
		res.Note = "no course_listings for subject+number"
		return res
	}
	if subject != res.CatalogSubject {
		res.CatalogSubject = subject
		res.Note = "matched without alias"
	}

	// Unique canonical IDs among listings; optionally filter by validity years.
	candidates := uniqueCanonicalIDs(listings, survey.SemesterSeason, survey.SemesterYear)
	res.CandidateCount = len(candidates)

	if len(candidates) == 1 && method != MethodLetteredSection {
		id := candidates[0]
		res.CanonicalCourseID = &id
		res.Status = MatchExactListing
		res.Method = method
		res.OfferingID = findOfferingID(db, id, survey.ProfessorID, survey.SemesterSeason, survey.SemesterYear)
		return res
	}

	// Disambiguate via professor ∩ offerings. Lettered sections always go through here:
	// a bare "301" says nothing about which section, so the professor has to pick it,
	// and only from the survey's own term.
	lettered := method == MethodLetteredSection
	if lettered && (survey.SemesterSeason == nil || survey.SemesterYear == nil) {
		res.Note = "bare number with lettered sections needs a semester"
		return res
	}
	matched, offeringID, note := disambiguate(db, candidates, survey.ProfessorID, survey.SemesterSeason, survey.SemesterYear, !lettered)
	res.Note = note
	if matched != nil {
		res.CanonicalCourseID = matched
		res.OfferingID = offeringID
		res.Status = MatchDisambiguated
		res.Method = method
		return res
	}
	if note == "ambiguous" {
		res.Status = MatchAmbiguous
		return res
	}
	res.Status = MatchUnmatched
	if res.Note == "" {
		res.Note = "multiple CRSE_IDs; professor did not disambiguate"
	}
	return res
}

// findListings tries each subject and code in order and returns the first listings found.
func findListings(db *gorm.DB, subjects []string, codes []CourseCode) ([]models.CourseListing, string, string, error) {
	for _, subject := range subjects {
		var all []models.CourseListing
		method := ""
		for _, c := range codes {
			found, m, err := findListingsForCode(db, subject, c)
			if err != nil {
				return nil, "", "", err
			}
			if len(found) > 0 && method == "" {
				method = m
			}
			all = append(all, found...)
		}
		if len(all) > 0 {
			if len(codes) > 1 && method == MethodListing {
				method = MethodNormalizedCode
			}
			return all, method, subject, nil
		}
	}
	return nil, "", "", nil
}

func findListingsForCode(db *gorm.DB, subject string, c CourseCode) ([]models.CourseListing, string, error) {
	var listings []models.CourseListing
	if err := db.Where("subject = ? AND number = ? AND letter = ?", subject, c.Number, c.Letter).
		Find(&listings).Error; err != nil {
		return nil, "", err
	}
	if len(listings) > 0 {
		if c.Normalized {
			return listings, MethodNormalizedCode, nil
		}
		return listings, MethodListing, nil
	}
	if c.Letter == "T" {
		if err := db.Where("subject = ? AND number = ? AND letter = ''", subject, c.Number).
			Find(&listings).Error; err != nil {
			return nil, "", err
		}
		if len(listings) > 0 {
			return listings, MethodTutorialSuffix, nil
		}
	}
	if c.Letter == "" {
		if err := db.Where("subject = ? AND number = ? AND letter <> ''", subject, c.Number).
			Find(&listings).Error; err != nil {
			return nil, "", err
		}
		if len(listings) > 0 {
			return listings, MethodLetteredSection, nil
		}
	}
	return nil, "", nil
}

// matchByClassNumber finds the one offering in the survey's term with this class number
// that the professor taught.
func matchByClassNumber(db *gorm.DB, numbers []int, professorID uint, season *string, year *int) *models.Offering {
	if season == nil || year == nil || len(numbers) == 0 {
		return nil
	}
	y, term, ok := CatalogYearTerm(*season, *year)
	if !ok {
		return nil
	}
	var offerings []models.Offering
	if err := db.Table("offerings").
		Select("DISTINCT offerings.*").
		Joins("JOIN offering_instructors ON offering_instructors.offering_id = offerings.id AND offering_instructors.deleted_at IS NULL").
		Where("offerings.deleted_at IS NULL").
		Where("offering_instructors.user_id = ?", professorID).
		Where("offerings.year = ? AND offerings.term = ?", y, term).
		Where("offerings.class_nbr IN (?)", numbers).
		Find(&offerings).Error; err != nil || len(offerings) != 1 {
		return nil
	}
	return &offerings[0]
}

// matchByProfessorNumber looks for a course with one of these numbers (in any subject) that
// the professor taught in the survey's term. It needs a term, and exactly one course.
func matchByProfessorNumber(db *gorm.DB, numbers []int, professorID uint, season *string, year *int) (*uint, *uint) {
	if season == nil || year == nil || len(numbers) == 0 {
		return nil, nil
	}
	y, term, ok := CatalogYearTerm(*season, *year)
	if !ok {
		return nil, nil
	}
	type hit struct {
		CourseCanonicalID uint
		OfferingID        uint
	}
	var hits []hit
	if err := db.Table("offerings").
		Select("DISTINCT offerings.course_canonical_id, offerings.id AS offering_id").
		Joins("JOIN offering_instructors ON offering_instructors.offering_id = offerings.id AND offering_instructors.deleted_at IS NULL").
		Joins("JOIN course_listings ON course_listings.course_canonical_id = offerings.course_canonical_id AND course_listings.deleted_at IS NULL").
		Where("offerings.deleted_at IS NULL").
		Where("offering_instructors.user_id = ?", professorID).
		Where("offerings.year = ? AND offerings.term = ?", y, term).
		Where("course_listings.number IN (?)", numbers).
		Scan(&hits).Error; err != nil || len(hits) == 0 {
		return nil, nil
	}
	canonicalID := hits[0].CourseCanonicalID
	offerings := map[uint]struct{}{}
	for _, h := range hits {
		if h.CourseCanonicalID != canonicalID {
			return nil, nil
		}
		offerings[h.OfferingID] = struct{}{}
	}
	var offeringID *uint
	if len(offerings) == 1 {
		o := hits[0].OfferingID
		offeringID = &o
	}
	return &canonicalID, offeringID
}

func uniqueCanonicalIDs(listings []models.CourseListing, season *string, year *int) []uint {
	seen := map[uint]struct{}{}
	var out []uint
	var catYear int
	var hasYear bool
	if season != nil && year != nil {
		if y, _, ok := CatalogYearTerm(*season, *year); ok {
			catYear = y
			hasYear = true
		}
	}
	for _, l := range listings {
		if hasYear && (catYear < l.FirstValidYear || catYear > l.LastValidYear) {
			continue
		}
		if _, ok := seen[l.CourseCanonicalID]; ok {
			continue
		}
		seen[l.CourseCanonicalID] = struct{}{}
		out = append(out, l.CourseCanonicalID)
	}
	// If year filter wiped everything, fall back to all listings.
	if hasYear && len(out) == 0 {
		for _, l := range listings {
			if _, ok := seen[l.CourseCanonicalID]; ok {
				continue
			}
			seen[l.CourseCanonicalID] = struct{}{}
			out = append(out, l.CourseCanonicalID)
		}
	}
	return out
}

// disambiguate picks the candidate this professor taught, preferring the survey's term.
// With anyTermFallback, it also accepts offerings from other terms.
func disambiguate(db *gorm.DB, candidates []uint, professorID uint, season *string, year *int, anyTermFallback bool) (canonicalID *uint, offeringID *uint, note string) {
	type hit struct {
		CourseCanonicalID uint
		OfferingID        uint
	}
	var hits []hit

	addHits := func(restrictTerm bool) error {
		hits = hits[:0]
		for _, cid := range candidates {
			q := db.Table("offerings").
				Select("offerings.course_canonical_id, offerings.id as offering_id").
				Joins("JOIN offering_instructors ON offering_instructors.offering_id = offerings.id AND offering_instructors.deleted_at IS NULL").
				Where("offerings.deleted_at IS NULL").
				Where("offerings.course_canonical_id = ?", cid).
				Where("offering_instructors.user_id = ?", professorID)
			if restrictTerm && season != nil && year != nil {
				if y, term, ok := CatalogYearTerm(*season, *year); ok {
					q = q.Where("offerings.year = ? AND offerings.term = ?", y, term)
				}
			}
			rows, err := q.Rows()
			if err != nil {
				return err
			}
			for rows.Next() {
				var h hit
				if err := rows.Scan(&h.CourseCanonicalID, &h.OfferingID); err != nil {
					rows.Close()
					return err
				}
				hits = append(hits, h)
			}
			rows.Close()
		}
		return nil
	}

	// Prefer term-restricted intersection when semester info exists.
	if season != nil && year != nil {
		if err := addHits(true); err != nil {
			return nil, nil, err.Error()
		}
	}
	if len(hits) == 0 && anyTermFallback {
		if err := addHits(false); err != nil {
			return nil, nil, err.Error()
		}
	}
	if len(hits) == 0 {
		return nil, nil, "no offering with this professor among candidates"
	}

	canonSet := map[uint]uint{} // canonical -> one offering
	for _, h := range hits {
		if _, ok := canonSet[h.CourseCanonicalID]; !ok {
			canonSet[h.CourseCanonicalID] = h.OfferingID
		}
	}
	if len(canonSet) == 1 {
		for cid, oid := range canonSet {
			c := cid
			o := oid
			return &c, &o, "professor∩offering"
		}
	}
	return nil, nil, "ambiguous"
}

func findOfferingID(db *gorm.DB, canonicalID, professorID uint, season *string, year *int) *uint {
	q := db.Table("offerings").
		Select("offerings.id").
		Joins("JOIN offering_instructors ON offering_instructors.offering_id = offerings.id AND offering_instructors.deleted_at IS NULL").
		Where("offerings.deleted_at IS NULL").
		Where("offerings.course_canonical_id = ?", canonicalID).
		Where("offering_instructors.user_id = ?", professorID)
	if season != nil && year != nil {
		if y, term, ok := CatalogYearTerm(*season, *year); ok {
			q = q.Where("offerings.year = ? AND offerings.term = ?", y, term)
		}
	}
	var ids []uint
	if err := q.Pluck("offerings.id", &ids).Error; err != nil || len(ids) != 1 {
		return nil
	}
	id := ids[0]
	return &id
}

// WriteReportJSON writes the report to w.
func WriteReportJSON(w io.Writer, report *Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}
