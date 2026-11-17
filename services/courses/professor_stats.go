package courses

import (
	"sort"

	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

// Also the minimum for a colleague to count in the comparison
const minStatsReviews = 5

type ProfessorStatsResponse struct {
	Department string                   `json:"department,omitempty"`
	Questions  []ProfessorQuestionStats `json:"questions"`
}

// Values are averages, or for yes/no questions (percent) the share of yeses from 0 to 100
type ProfessorQuestionStats struct {
	Question string  `json:"question"` // the FactrakSurvey field, e.g. wouldRecommendCourse
	Percent  bool    `json:"percent"`
	Value    float64 `json:"value"` // 0 when count is 0
	Count    int     `json:"count"`
	// Other professors' values in the department, ascending; everyone compared has at least 5 answers
	Department []float64           `json:"department,omitempty"`
	Terms      []ProfessorTermStat `json:"terms"` // terms with answers, oldest first
}

type ProfessorTermStat struct {
	Strm  int     `json:"strm"`
	Label string  `json:"label"`
	Value float64 `json:"value"`
	Count int     `json:"count"`
}

func surveyTerm(season string, year int) (strm int, label string, ok bool) {
	catalogYear, term, digit := year, "", 0
	switch season {
	case models.FactrakSurveySemesterSeasonFall:
		catalogYear, term, digit = year+1, "Fall", 1
	case models.FactrakSurveySemesterSeasonWinterStudy:
		term, digit = "Winter", 2
	case models.FactrakSurveySemesterSeasonSpring:
		term, digit = "Spring", 3
	default:
		return 0, "", false
	}
	if year == 0 {
		return 0, "", false
	}
	return 1000 + (catalogYear%100)*10 + digit, TermLabel(catalogYear, term), true
}

func statValue(avg float64, q models.FactrakQuestion) float64 {
	if q.YesNo {
		return avg * 100
	}
	return avg
}

// GetProfessorStats godoc
// @Summary Get a professor's Factrak stats
// @Description For each numeric or yes/no Factrak question, the professor's average (yes/no questions as a percentage) and answer count, every other professor's value in their department (each with at least 5 answers; only when this professor has at least 5), and the professor's value per term. Counts the reviews the professor's ratings do (the last 5 years). Needs full Factrak access.
// @ID getCourseProfessorStats
// @Tags courses
// @Produce json
// @Param professorID path int true "Professor (WSO user) ID"
// @Success 200 {object} courses.ProfessorStatsResponse
// @Failure 400 {object} services.BaseErrorResponse "1100: could not parse malformed request data"
// @Failure 404 {object} services.BaseErrorResponse "404: record not found"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /courses/professors/{professorID}/stats [get]
func (t *Controller) GetProfessorStats(c *gin.Context) {
	profID, err := services.GetUIntParam(c, "professorID")
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}
	prof, err := t.catalog.GetProfessor(profID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	terms, err := t.surveys.ProfessorTermSurveyStats(profID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	var dept []models.SurveyStats
	resp := ProfessorStatsResponse{Questions: make([]ProfessorQuestionStats, len(models.FactrakQuestions))}
	if prof.DepartmentID != nil {
		if dept, err = t.surveys.DepartmentSurveyStats(*prof.DepartmentID); err != nil {
			t.RespondError(c, err)
			return
		}
		if prof.Department != nil {
			resp.Department = prof.Department.Name
		}
	}

	for i, q := range models.FactrakQuestions {
		qs := ProfessorQuestionStats{Question: q.Name, Percent: q.YesNo, Terms: []ProfessorTermStat{}}
		qs.Value, qs.Count = pooledStat(terms, i, q)
		for _, s := range terms {
			if s.Num[i] == 0 {
				continue
			}
			if strm, label, ok := surveyTerm(s.SemesterSeason, s.SemesterYear); ok {
				qs.Terms = append(qs.Terms, ProfessorTermStat{Strm: strm, Label: label, Value: statValue(s.Avg[i].Float64, q), Count: s.Num[i]})
			}
		}
		sort.Slice(qs.Terms, func(a, b int) bool { return qs.Terms[a].Strm < qs.Terms[b].Strm })
		if qs.Count >= minStatsReviews {
			qs.Department = peerStats(dept, profID, i, q, minStatsReviews)
		}
		resp.Questions[i] = qs
	}
	t.RespondOK(c, resp)
}

func pooledStat(stats []models.SurveyStats, i int, q models.FactrakQuestion) (float64, int) {
	var sum float64
	var n int
	for _, s := range stats {
		sum += s.Avg[i].Float64 * float64(s.Num[i])
		n += s.Num[i]
	}
	if n == 0 {
		return 0, 0
	}
	return statValue(sum/float64(n), q), n
}

func peerStats(stats []models.SurveyStats, self uint, i int, q models.FactrakQuestion, minCount int) []float64 {
	out := []float64{}
	for _, s := range stats {
		if s.ID != self && s.Num[i] >= minCount {
			out = append(out, statValue(s.Avg[i].Float64, q))
		}
	}
	sort.Float64s(out)
	return out
}

const minCourseStatsReviews = 3

type GetCourseStatsParams struct {
	ProfessorID *uint `form:"professorID"`
}

type CourseStatsResponse struct {
	Department string                `json:"department,omitempty"`
	Questions  []CourseQuestionStats `json:"questions"`
}

// Department is ascending: other courses in the area, or for one professor, the course's other professors (3+ answers each)
type CourseQuestionStats struct {
	Question   string    `json:"question"`
	Percent    bool      `json:"percent"`
	Value      float64   `json:"value"`
	Count      int       `json:"count"`
	Department []float64 `json:"department,omitempty"`
}

// GetCourseStats godoc
// @Summary Get a course's Factrak stats
// @Description For each numeric or yes/no Factrak question, a catalog course's average (yes/no questions as a percentage) and answer count over the reviews its ratings count (all cross-listings, the last 5 years, without mismatched links). For the whole course, every other course's value in the area of its primary subject (each with at least 3 answers); with professorID, that professor's sections and every other professor's value for this course (each with at least 3 answers). Needs full Factrak access.
// @ID getCatalogCourseStats
// @Tags courses
// @Produce json
// @Param crseID path string true "Catalog course ID"
// @Param professorID query int false "Only reviews of this professor"
// @Success 200 {object} courses.CourseStatsResponse
// @Failure 400 {object} services.BaseErrorResponse "1100: could not parse malformed request data"
// @Failure 404 {object} services.BaseErrorResponse "404: record not found"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /courses/catalog/{crseID}/stats [get]
func (t *Controller) GetCourseStats(c *gin.Context) {
	var params GetCourseStatsParams
	if err := c.ShouldBindQuery(&params); err != nil {
		t.RespondBadBind(c, err)
		return
	}
	courseID, err := t.catalog.CanonicalID(c.Param("crseID"))
	if err != nil {
		t.RespondError(c, err)
		return
	}
	own, err := t.surveys.CourseSurveyStats(courseID, params.ProfessorID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	resp := CourseStatsResponse{Questions: make([]CourseQuestionStats, len(models.FactrakQuestions))}
	var dept, profs []models.SurveyStats
	if params.ProfessorID != nil {
		if profs, err = t.surveys.CourseProfessorSurveyStats(courseID); err != nil {
			t.RespondError(c, err)
			return
		}
	} else {
		area, err := t.catalog.CourseArea(courseID)
		if err != nil {
			t.RespondError(c, err)
			return
		}
		if area != nil {
			resp.Department = area.Name
			if dept, err = t.surveys.SubjectCourseSurveyStats(areaCatalogSubjects(area.Abbreviation)); err != nil {
				t.RespondError(c, err)
				return
			}
		}
	}

	for i, q := range models.FactrakQuestions {
		qs := CourseQuestionStats{Question: q.Name, Percent: q.YesNo}
		qs.Value, qs.Count = pooledStat([]models.SurveyStats{own}, i, q)
		switch {
		case qs.Count == 0:
		case params.ProfessorID != nil:
			qs.Department = peerStats(profs, *params.ProfessorID, i, q, minCourseStatsReviews)
		case dept != nil:
			qs.Department = peerStats(dept, courseID, i, q, minCourseStatsReviews)
		}
		resp.Questions[i] = qs
	}
	t.RespondOK(c, resp)
}
