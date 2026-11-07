package factrak_backfill

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
)

// Override is a reviewed decision for one survey the rules couldn't match.
// It names the course by catalog CRSE_ID (stable across databases), not by row ID.
type Override struct {
	SurveyID uint `json:"surveyID"`
	// FiledAs is the survey's legacy course ("PHYS 354"). The override is skipped unless it
	// still matches, so it can't land on an unrelated survey in a DB with different IDs.
	FiledAs string `json:"filedAs"`
	CrseID  string `json:"crseID"`
	// Optional: pin the offering too (catalog term + class number).
	Strm     int    `json:"strm,omitempty"`
	ClassNbr int    `json:"classNbr,omitempty"`
	Source   string `json:"source,omitempty"` // who decided, e.g. "llm-review+human"
	Reason   string `json:"reason,omitempty"`
}

//go:embed overrides.json
var embeddedOverrides []byte

// EmbeddedOverrides returns the overrides checked into overrides.json.
func EmbeddedOverrides() ([]Override, error) {
	var out []Override
	if err := json.Unmarshal(embeddedOverrides, &out); err != nil {
		return nil, fmt.Errorf("parse overrides.json: %w", err)
	}
	return out, nil
}

// applyOverride returns false when the override doesn't fit this survey; the caller then
// falls back to the normal rules.
func applyOverride(db *gorm.DB, survey *models.FactrakSurvey, o Override) (Result, bool) {
	res := newResult(survey)
	if survey.Course == nil || survey.Course.AreaOfStudy == nil {
		return res, false
	}
	res.Abbrev = strings.ToUpper(survey.Course.AreaOfStudy.Abbreviation)
	res.RawNumber = survey.Course.Number
	filedAs := res.Abbrev + " " + strings.TrimSpace(survey.Course.Number)
	if !strings.EqualFold(filedAs, strings.TrimSpace(o.FiledAs)) {
		return res, false
	}
	res.Method = MethodManualOverride
	res.Note = o.Reason

	var canonical models.CourseCanonical
	if err := db.Where("crse_id = ?", o.CrseID).First(&canonical).Error; err != nil {
		return res, false
	}
	id := canonical.ID
	res.CanonicalCourseID = &id
	res.Status = MatchDisambiguated
	res.CandidateCount = 1

	if o.Strm != 0 && o.ClassNbr != 0 {
		var offering models.Offering
		if err := db.Where("strm = ? AND class_nbr = ? AND course_canonical_id = ?", o.Strm, o.ClassNbr, id).
			First(&offering).Error; err == nil {
			oid := offering.ID
			res.OfferingID = &oid
		}
	}
	return res, true
}
