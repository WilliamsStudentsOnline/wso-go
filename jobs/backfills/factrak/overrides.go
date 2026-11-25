package factrak

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
)

// Keyed by CRSE_ID, which is stable across databases, unlike row ids
type Override struct {
	SurveyID uint `json:"surveyID"`
	// Skipped unless this still matches, so an override can't hit an unrelated survey in a DB with different ids
	FiledAs  string `json:"filedAs"`
	CrseID   string `json:"crseID"`
	Strm     int    `json:"strm,omitempty"`
	ClassNbr int    `json:"classNbr,omitempty"`
	Source   string `json:"source,omitempty"` // who decided, e.g. "llm-review+human"
	Reason   string `json:"reason,omitempty"`
}

//go:embed overrides.json
var embeddedOverrides []byte

func EmbeddedOverrides() ([]Override, error) {
	var out []Override
	if err := json.Unmarshal(embeddedOverrides, &out); err != nil {
		return nil, fmt.Errorf("parse overrides.json: %w", err)
	}
	return out, nil
}

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
	res.CourseCanonicalID = &id
	res.Status = MatchDisambiguated
	res.CandidateCount = 1

	if o.Strm != 0 && o.ClassNbr != 0 {
		var offering models.CourseOffering
		if err := db.Where("strm = ? AND class_nbr = ? AND course_canonical_id = ?", o.Strm, o.ClassNbr, id).
			First(&offering).Error; err == nil {
			oid := offering.ID
			res.CourseOfferingID = &oid
		}
	}
	return res, true
}
