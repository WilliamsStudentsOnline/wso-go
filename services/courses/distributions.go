package courses

import (
	"strings"

	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

type PlanDistributions struct {
	Div1 int `json:"div1"`
	Div2 int `json:"div2"`
	Div3 int `json:"div3"`
	WS   int `json:"ws"`
	QFR  int `json:"qfr"`
	DPE  int `json:"dpe"`
}

type DistributionsResponse struct {
	Earned      PlanDistributions `json:"earned"`
	WithPlanned PlanDistributions `json:"withPlanned"`
	Needed      PlanDistributions `json:"needed"`
}

// Div I-III count at most two courses per subject
var distributionsNeeded = PlanDistributions{Div1: 3, Div2: 3, Div3: 3, WS: 2, QFR: 1, DPE: 1}

// GetMyDistributions godoc
// @Summary My distribution requirements
// @Description Div I/II/III (at most two courses per subject count toward a division), writing skills, QFR and DPE over the signed-in user's courses (not hidden): earned through the current term, and including planned and someday courses. No class year needed.
// @ID getMyDistributions
// @Tags courses
// @Produce json
// @Success 200 {object} courses.DistributionsResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /courses/mine/distributions [get]
func (t *Controller) GetMyDistributions(c *gin.Context) {
	entries, err := t.mine.PlanCourses(services.GetUserID(c))
	if err != nil {
		t.RespondError(c, err)
		return
	}
	current := models.CurrentStrm(t.now())
	var earned, all []entryCounts
	for _, uc := range entries {
		dc, ok := distributionCourse(uc)
		if !ok {
			continue
		}
		all = append(all, dc)
		if uc.CourseInSemester != nil && uc.CourseInSemester.Strm <= current {
			earned = append(earned, dc)
		}
	}
	t.RespondOK(c, DistributionsResponse{
		Earned:      tally(earned),
		WithPlanned: tally(all),
		Needed:      distributionsNeeded,
	})
}

func entryTerm(uc *models.UserCourse) *models.CourseInSemester {
	cis := uc.CourseInSemester
	if cis == nil && uc.CourseCanonical != nil && len(uc.CourseCanonical.CoursesInSemester) > 0 {
		cis = uc.CourseCanonical.CoursesInSemester[0] // newest: what a someday course would count as
	}
	return cis
}

func entryCodes(uc *models.UserCourse) []string {
	if uc.CourseCanonical == nil {
		return nil
	}
	year := 0
	if cis := entryTerm(uc); cis != nil {
		year = cis.Year
	}
	var codes []string
	for _, l := range listingsForYear(uc.CourseCanonical.Listings, year) {
		codes = append(codes, l.Code)
	}
	return codes
}

func distributionCourse(uc *models.UserCourse) (entryCounts, bool) {
	c := countsFor(uc)
	if c.Subject == "" {
		if codes := entryCodes(uc); len(codes) > 0 {
			c.Subject = strings.SplitN(codes[0], " ", 2)[0]
		}
	}
	return c, c.Known
}

func tally(courses []entryCounts) PlanDistributions {
	var d PlanDistributions
	per := map[int]map[string]int{1: {}, 2: {}, 3: {}}
	count := func(div int, subject string) int {
		per[div][subject]++
		if per[div][subject] <= 2 {
			return 1
		}
		return 0
	}
	for _, c := range courses {
		if c.Div1 {
			d.Div1 += count(1, c.Subject)
		}
		if c.Div2 {
			d.Div2 += count(2, c.Subject)
		}
		if c.Div3 {
			d.Div3 += count(3, c.Subject)
		}
		if c.WAC {
			d.WS++
		}
		if c.QFR {
			d.QFR++
		}
		if c.DPE {
			d.DPE++
		}
	}
	return d
}
