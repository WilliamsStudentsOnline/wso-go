package factrak

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

const (
	MethodListing             = "listing"
	MethodNormalizedCode      = "normalized_code"       // junk suffix dropped ("101LAB", "201-02") or "140/204" split
	MethodTutorialSuffix      = "tutorial_suffix"       // Factrak "127T", catalog lists "127"
	MethodLetteredSection     = "lettered_section"      // Factrak "301", catalog only has "301A".."301M"
	MethodProfessorSameNumber = "professor_same_number" // wrong subject picked; professor taught that number that term
	MethodClassNumber         = "class_number"          // student typed the registration class number ("1804")
	MethodManualOverride      = "manual_override"
)

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
	CourseCanonicalID *uint  `json:"canonicalCourseID,omitempty"`
	CourseOfferingID  *uint  `json:"offeringID,omitempty"`
	CandidateCount    int    `json:"candidateCount"`
	Note              string `json:"note,omitempty"`
}

type Report struct {
	Total              int               `json:"total"`
	ExactListing       int               `json:"exactListing"`
	Disambiguated      int               `json:"disambiguated"`
	Ambiguous          int               `json:"ambiguous"`
	Unmatched          int               `json:"unmatched"`
	AlreadyLinked      int               `json:"alreadyLinked"`
	Applied            int               `json:"applied"`
	OverridesSkipped   int               `json:"overridesSkipped"`
	ByMethod           map[string]int    `json:"byMethod"`
	LinkSources        map[string]int    `json:"linkSources,omitempty"`
	UserCourses        int64             `json:"userCourses,omitempty"`
	UserCourseSections int64             `json:"userCourseSections,omitempty"`
	Results            []Result          `json:"results"`
	AliasMap           map[string]string `json:"aliasMap"`
}

type Options struct {
	DryRun         bool
	SkipLinked     bool
	IncludeResults bool
	Overrides      []Override
}

type CourseCode struct {
	Number     int
	Letter     string
	Normalized bool
}

var leadingDigits = regexp.MustCompile(`^(\d+)(.*)$`)
var singleLetter = regexp.MustCompile(`^[A-Z]$`)

// Factrak numbers are typed by hand: "301B" → 301/B; "101LAB" and "201-02" drop the suffix; "140/204" → both
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

// Factrak years are calendar years, catalog years are the academic year's spring (Fall 2026 → 2027)
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

func ResolveSubject(abbrev string, aliases map[string]string) string {
	abbrev = strings.ToUpper(strings.TrimSpace(abbrev))
	if mapped, ok := aliases[abbrev]; ok {
		return mapped
	}
	return abbrev
}

func RunBackfill(db *gorm.DB, log *zap.SugaredLogger, opts Options) (*Report, error) {
	if log == nil {
		log = zap.NewNop().Sugar()
	}
	aliases := models.RenamedSubjects
	overrideList := opts.Overrides
	if overrideList == nil {
		var err error
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
		q = q.Where("course_canonical_id IS NULL")
	}

	var surveys []models.FactrakSurvey
	if err := q.Find(&surveys).Error; err != nil {
		return nil, err
	}
	report.Total = len(surveys)

	for i := range surveys {
		var res Result
		applied := false
		if o, ok := overrides[surveys[i].ID]; ok && surveys[i].CourseCanonicalID == nil {
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
				"course_canonical_id": res.CourseCanonicalID,
			}
			if res.CourseOfferingID != nil {
				updates["course_offering_id"] = res.CourseOfferingID
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

	if !opts.DryRun {
		sources, err := AssignCourseLinkVerificationSources(db, log)
		if err != nil {
			return report, err
		}
		report.LinkSources = sources
		if report.UserCourses, report.UserCourseSections, err = SyncReviewedUserCourses(db, log); err != nil {
			return report, err
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

	if survey.CourseCanonicalID != nil {
		res.Status = MatchAlreadyLinked
		res.CourseCanonicalID = survey.CourseCanonicalID
		res.CourseOfferingID = survey.CourseOfferingID
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
		numbers := make([]int, 0, len(codes))
		for _, c := range codes {
			numbers = append(numbers, c.Number)
		}
		// Students sometimes typed the class number; these repeat across terms, so require term and professor
		if offering := matchByClassNumber(db, numbers, survey.ProfessorID, survey.SemesterSeason, survey.SemesterYear); offering != nil {
			res.CourseCanonicalID = &offering.CourseCanonicalID
			res.CourseOfferingID = &offering.ID
			res.Status = MatchDisambiguated
			res.Method = MethodClassNumber
			res.CandidateCount = 1
			return res
		}
		// Students sometimes picked the wrong department (PHYS for PSYC, STS for STAT)
		canonicalID, offeringID := matchByProfessorNumber(db, numbers, survey.ProfessorID, survey.SemesterSeason, survey.SemesterYear)
		if canonicalID != nil {
			res.CourseCanonicalID = canonicalID
			res.CourseOfferingID = offeringID
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

	candidates := uniqueCanonicalIDs(listings, survey.SemesterSeason, survey.SemesterYear)
	res.CandidateCount = len(candidates)

	if len(candidates) == 1 && method != MethodLetteredSection {
		id := candidates[0]
		res.CourseCanonicalID = &id
		res.Status = MatchExactListing
		res.Method = method
		res.CourseOfferingID = findOfferingID(db, id, survey.ProfessorID, survey.SemesterSeason, survey.SemesterYear)
		return res
	}

	// A bare "301" doesn't say which lettered section, so the professor must pick it, only from the survey's own term
	lettered := method == MethodLetteredSection
	if lettered && (survey.SemesterSeason == nil || survey.SemesterYear == nil) {
		res.Note = "bare number with lettered sections needs a semester"
		return res
	}
	matched, offeringID, note := disambiguate(db, candidates, survey.ProfessorID, survey.SemesterSeason, survey.SemesterYear, !lettered)
	res.Note = note
	if matched != nil {
		res.CourseCanonicalID = matched
		res.CourseOfferingID = offeringID
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

func matchByClassNumber(db *gorm.DB, numbers []int, professorID uint, season *string, year *int) *models.CourseOffering {
	if season == nil || year == nil || len(numbers) == 0 {
		return nil
	}
	y, term, ok := CatalogYearTerm(*season, *year)
	if !ok {
		return nil
	}
	var offerings []models.CourseOffering
	if err := db.Table("course_offerings").
		Select("DISTINCT course_offerings.*").
		Joins("JOIN course_offering_instructors ON course_offering_instructors.course_offering_id = course_offerings.id AND course_offering_instructors.deleted_at IS NULL").
		Where("course_offerings.deleted_at IS NULL").
		Where("course_offering_instructors.user_id = ?", professorID).
		Where("course_offerings.year = ? AND course_offerings.term = ?", y, term).
		Where("course_offerings.class_nbr IN (?)", numbers).
		Find(&offerings).Error; err != nil || len(offerings) != 1 {
		return nil
	}
	return &offerings[0]
}

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
		CourseOfferingID  uint
	}
	var hits []hit
	if err := db.Table("course_offerings").
		Select("DISTINCT course_offerings.course_canonical_id, course_offerings.id AS course_offering_id").
		Joins("JOIN course_offering_instructors ON course_offering_instructors.course_offering_id = course_offerings.id AND course_offering_instructors.deleted_at IS NULL").
		Joins("JOIN course_listings ON course_listings.course_canonical_id = course_offerings.course_canonical_id AND course_listings.deleted_at IS NULL").
		Where("course_offerings.deleted_at IS NULL").
		Where("course_offering_instructors.user_id = ?", professorID).
		Where("course_offerings.year = ? AND course_offerings.term = ?", y, term).
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
		offerings[h.CourseOfferingID] = struct{}{}
	}
	var offeringID *uint
	if len(offerings) == 1 {
		o := hits[0].CourseOfferingID
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

func disambiguate(db *gorm.DB, candidates []uint, professorID uint, season *string, year *int, anyTermFallback bool) (canonicalID *uint, offeringID *uint, note string) {
	type hit struct {
		CourseCanonicalID uint
		CourseOfferingID  uint
	}
	var hits []hit

	addHits := func(restrictTerm bool) error {
		hits = hits[:0]
		for _, cid := range candidates {
			q := db.Table("course_offerings").
				Select("course_offerings.course_canonical_id, course_offerings.id as course_offering_id").
				Joins("JOIN course_offering_instructors ON course_offering_instructors.course_offering_id = course_offerings.id AND course_offering_instructors.deleted_at IS NULL").
				Where("course_offerings.deleted_at IS NULL").
				Where("course_offerings.course_canonical_id = ?", cid).
				Where("course_offering_instructors.user_id = ?", professorID)
			if restrictTerm && season != nil && year != nil {
				if y, term, ok := CatalogYearTerm(*season, *year); ok {
					q = q.Where("course_offerings.year = ? AND course_offerings.term = ?", y, term)
				}
			}
			rows, err := q.Rows()
			if err != nil {
				return err
			}
			for rows.Next() {
				var h hit
				if err := rows.Scan(&h.CourseCanonicalID, &h.CourseOfferingID); err != nil {
					rows.Close()
					return err
				}
				hits = append(hits, h)
			}
			rows.Close()
		}
		return nil
	}

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

	canonSet := map[uint]uint{}
	for _, h := range hits {
		if _, ok := canonSet[h.CourseCanonicalID]; !ok {
			canonSet[h.CourseCanonicalID] = h.CourseOfferingID
		}
	}
	if len(canonSet) == 1 {
		for cid, oid := range canonSet {
			return &cid, &oid, "professor∩offering"
		}
	}
	return nil, nil, "ambiguous"
}

func findOfferingID(db *gorm.DB, canonicalID, professorID uint, season *string, year *int) *uint {
	q := db.Table("course_offerings").
		Select("course_offerings.id").
		Joins("JOIN course_offering_instructors ON course_offering_instructors.course_offering_id = course_offerings.id AND course_offering_instructors.deleted_at IS NULL").
		Where("course_offerings.deleted_at IS NULL").
		Where("course_offerings.course_canonical_id = ?", canonicalID).
		Where("course_offering_instructors.user_id = ?", professorID)
	if season != nil && year != nil {
		if y, term, ok := CatalogYearTerm(*season, *year); ok {
			q = q.Where("course_offerings.year = ? AND course_offerings.term = ?", y, term)
		}
	}
	var ids []uint
	if err := q.Pluck("course_offerings.id", &ids).Error; err != nil || len(ids) != 1 {
		return nil
	}
	id := ids[0]
	return &id
}

func WriteReportJSON(w io.Writer, report *Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}
