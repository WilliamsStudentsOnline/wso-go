package models

import (
	"database/sql"
	"strings"

	"github.com/jinzhu/gorm"
)

type FactrakQuestion struct {
	Column string
	Name   string // its JSON name on FactrakSurvey, e.g. "wouldRecommendCourse"
	YesNo  bool   // averaged as a share of yeses
}

var FactrakQuestions = []FactrakQuestion{
	{"would_recommend_course", "wouldRecommendCourse", true},
	{"course_workload", "courseWorkload", false},
	{"course_stimulating", "courseStimulating", false},
	{"would_take_another", "wouldTakeAnother", true},
	{"approachability", "approachability", false},
	{"lead_lecture", "leadLecture", false},
	{"promote_discussion", "promoteDiscussion", false},
	{"outside_helpfulness", "outsideHelpfulness", false},
	{"mental_health_support", "mentalHealthSupport", false},
}

// Values are in FactrakQuestions order; Avg is invalid when nobody answered
type SurveyStats struct {
	ID             uint
	SemesterSeason string
	SemesterYear   int
	Avg            []sql.NullFloat64
	Num            []int
}

func (m *FactrakSurveyModel) DepartmentSurveyStats(departmentID uint) ([]SurveyStats, error) {
	return m.surveyStats([]string{"factrak_surveys.professor_id"}, func(db *gorm.DB) *gorm.DB {
		return db.Joins("JOIN users ON users.id = factrak_surveys.professor_id").
			Where("users.department_id = ?", departmentID)
	})
}

func (m *FactrakSurveyModel) ProfessorTermSurveyStats(profID uint) ([]SurveyStats, error) {
	return m.surveyStats([]string{"factrak_surveys.semester_season", "factrak_surveys.semester_year"}, m.withProfessorID(profID))
}

func (m *FactrakSurveyModel) CourseSurveyStats(canonicalCourseID uint, profID *uint) (SurveyStats, error) {
	stats, err := m.surveyStats([]string{"factrak_surveys.course_canonical_id"}, func(db *gorm.DB) *gorm.DB {
		db = scopeCatalogLinked(db).Where("factrak_surveys.course_canonical_id = ?", canonicalCourseID)
		if profID != nil {
			db = db.Where("factrak_surveys.professor_id = ?", *profID)
		}
		return db
	})
	if err != nil {
		return SurveyStats{}, err
	}
	if len(stats) > 0 {
		return stats[0], nil
	}
	return SurveyStats{ID: canonicalCourseID, Avg: make([]sql.NullFloat64, len(FactrakQuestions)), Num: make([]int, len(FactrakQuestions))}, nil
}

func (m *FactrakSurveyModel) CourseProfessorSurveyStats(canonicalCourseID uint) ([]SurveyStats, error) {
	return m.surveyStats([]string{"factrak_surveys.professor_id"}, func(db *gorm.DB) *gorm.DB {
		return scopeCatalogLinked(db).Where("factrak_surveys.course_canonical_id = ?", canonicalCourseID)
	})
}

func (m *FactrakSurveyModel) SubjectCourseSurveyStats(subjects []string) ([]SurveyStats, error) {
	return m.surveyStats([]string{"factrak_surveys.course_canonical_id"}, func(db *gorm.DB) *gorm.DB {
		return scopeCatalogLinked(db).Where(`factrak_surveys.course_canonical_id IN (SELECT cl.course_canonical_id FROM course_listings cl
  WHERE cl.deleted_at IS NULL AND cl.subject IN (?))`, subjects)
	})
}

func scopeCatalogLinked(db *gorm.DB) *gorm.DB {
	return db.Where("(factrak_surveys.course_link_verification_source IS NULL OR factrak_surveys.course_link_verification_source <> ?)", CourseLinkVerificationSourceMismatch)
}

func (m *FactrakSurveyModel) surveyStats(group []string, scope func(*gorm.DB) *gorm.DB) ([]SurveyStats, error) {
	sel := append([]string{}, group...)
	for _, q := range FactrakQuestions {
		sel = append(sel, "AVG(factrak_surveys."+q.Column+")", "COUNT(factrak_surveys."+q.Column+")")
	}
	rows, err := m.DB.Model(&FactrakSurvey{}).Select(strings.Join(sel, ", ")).
		Scopes(m.scopeCurrent, scope).Group(strings.Join(group, ", ")).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []SurveyStats
	for rows.Next() {
		s := SurveyStats{Avg: make([]sql.NullFloat64, len(FactrakQuestions)), Num: make([]int, len(FactrakQuestions))}
		var season sql.NullString
		var year sql.NullInt64
		dest := []interface{}{&s.ID}
		if len(group) == 2 {
			dest = []interface{}{&season, &year}
		}
		for i := range FactrakQuestions {
			dest = append(dest, &s.Avg[i], &s.Num[i])
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}
		s.SemesterSeason, s.SemesterYear = season.String, int(year.Int64)
		out = append(out, s)
	}
	return out, rows.Err()
}
