package factrak_backfill

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

const (
	MatchExactListing   = "exact_listing"
	MatchDisambiguated  = "disambiguated_offering"
	MatchAmbiguous      = "ambiguous"
	MatchUnmatched      = "unmatched"
	MatchAlreadyLinked  = "already_linked"
)

// Result is one survey's backfill outcome.
type Result struct {
	SurveyID          uint   `json:"surveyID"`
	FactrakCourseID   uint   `json:"factrakCourseID"`
	Abbrev            string `json:"abbrev"`
	CatalogSubject    string `json:"catalogSubject"`
	Number            int    `json:"number"`
	Letter            string `json:"letter"`
	ProfessorID       uint   `json:"professorID"`
	SemesterSeason    string `json:"semesterSeason,omitempty"`
	SemesterYear      *int   `json:"semesterYear,omitempty"`
	Status            string `json:"status"`
	CanonicalCourseID *uint  `json:"canonicalCourseID,omitempty"`
	OfferingID        *uint  `json:"offeringID,omitempty"`
	CandidateCount    int    `json:"candidateCount"`
	Note              string `json:"note,omitempty"`
}

// Report summarizes a dry-run or apply pass.
type Report struct {
	Total          int            `json:"total"`
	ExactListing   int            `json:"exactListing"`
	Disambiguated  int            `json:"disambiguated"`
	Ambiguous      int            `json:"ambiguous"`
	Unmatched      int            `json:"unmatched"`
	AlreadyLinked  int            `json:"alreadyLinked"`
	Applied        int            `json:"applied"`
	Results        []Result       `json:"results"`
	AliasMap       map[string]string `json:"aliasMap"`
}

// Options controls backfill behavior.
type Options struct {
	DryRun          bool
	SkipLinked      bool // skip surveys that already have canonical_course_id
	IncludeResults  bool // keep per-survey rows in report (can be large)
}

// ParseCourseCode splits Factrak course numbers like "136" or "301B" into number + letter.
func ParseCourseCode(raw string) (number int, letter string, err error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, "", fmt.Errorf("empty course number")
	}
	i := 0
	for i < len(raw) && unicode.IsDigit(rune(raw[i])) {
		i++
	}
	if i == 0 {
		return 0, "", fmt.Errorf("course number has no digits: %q", raw)
	}
	number, err = strconv.Atoi(raw[:i])
	if err != nil {
		return 0, "", err
	}
	letter = strings.ToUpper(strings.TrimSpace(raw[i:]))
	return number, letter, nil
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

	report := &Report{
		AliasMap: aliases,
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
		res := matchSurvey(db, &surveys[i], aliases)
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

	log.Infof("Factrak backfill: total=%d exact=%d disambiguated=%d ambiguous=%d unmatched=%d already=%d applied=%d dryRun=%v",
		report.Total, report.ExactListing, report.Disambiguated, report.Ambiguous, report.Unmatched, report.AlreadyLinked, report.Applied, opts.DryRun)
	return report, nil
}

func matchSurvey(db *gorm.DB, survey *models.FactrakSurvey, aliases map[string]string) Result {
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

	number, letter, err := ParseCourseCode(survey.Course.Number)
	if err != nil {
		res.Note = err.Error()
		return res
	}
	res.Number = number
	res.Letter = letter

	var listings []models.CourseListing
	q := db.Where("subject = ? AND number = ? AND letter = ?", res.CatalogSubject, number, letter)
	if err := q.Find(&listings).Error; err != nil {
		res.Note = err.Error()
		return res
	}

	// Also try original abbrev if alias differed and found nothing.
	if len(listings) == 0 && res.CatalogSubject != res.Abbrev {
		if err := db.Where("subject = ? AND number = ? AND letter = ?", res.Abbrev, number, letter).
			Find(&listings).Error; err != nil {
			res.Note = err.Error()
			return res
		}
		if len(listings) > 0 {
			res.CatalogSubject = res.Abbrev
			res.Note = "matched without alias"
		}
	}

	if len(listings) == 0 {
		res.Note = "no course_listings for subject+number"
		return res
	}

	// Unique canonical IDs among listings; optionally filter by validity years.
	candidates := uniqueCanonicalIDs(listings, survey.SemesterSeason, survey.SemesterYear)
	res.CandidateCount = len(candidates)

	if len(candidates) == 1 {
		id := candidates[0]
		res.CanonicalCourseID = &id
		res.Status = MatchExactListing
		res.OfferingID = findOfferingID(db, id, survey.ProfessorID, survey.SemesterSeason, survey.SemesterYear)
		return res
	}

	// Disambiguate via professor ∩ offerings.
	matched, offeringID, note := disambiguate(db, candidates, survey.ProfessorID, survey.SemesterSeason, survey.SemesterYear)
	res.Note = note
	if matched != nil {
		res.CanonicalCourseID = matched
		res.OfferingID = offeringID
		res.Status = MatchDisambiguated
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

func disambiguate(db *gorm.DB, candidates []uint, professorID uint, season *string, year *int) (canonicalID *uint, offeringID *uint, note string) {
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
	if len(hits) == 0 {
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
