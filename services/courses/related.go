package courses

import (
	"sort"
	"strings"

	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/gin-gonic/gin"
)

const (
	relatedPerSection      = 6
	requirementRecentYears = 2
	maxRelatedAttributeUse = 40
)

var broadAttributePrefixes = []string{"DIV_", "DPE_", "QFR_", "WAC_"}

type CourseRelatedParams struct {
	Strm int `form:"strm"`
}

type RelatedProfessorResponse struct {
	Instructor InstructorResponse        `json:"instructor"`
	Courses    []ProfessorCourseResponse `json:"courses"`
}

type RelatedAttributeResponse struct {
	Code    string               `json:"code"`
	Courses []TermCourseResponse `json:"courses"`
}

// Strm is 0 when the course was never offered
type CourseRelatedResponse struct {
	Strm       int                        `json:"strm"`
	Professors []RelatedProfessorResponse `json:"professors"`
	Attributes []RelatedAttributeResponse `json:"attributes"`
}

// GetCourseRelated godoc
// @Summary Courses related to a course
// @Description What a course page's Related tab shows: for each instructor with a WSO user in the term, up to 6 of their other courses, most recently taught first; and for each of the course's attributes in the term, except distributions and attributes over 40 recent courses have, up to 6 other courses at their most recent offering with it (in the last few catalog years), by code.
// @ID getCatalogCourseRelated
// @Tags courses
// @Produce json
// @Param crseID path string true "Catalog course ID"
// @Param strm query int false "Term code for professors and attributes, e.g. 1271 (Fall 2026); default the latest term the course was offered"
// @Success 200 {object} courses.CourseRelatedResponse
// @Failure 400 {object} services.BaseErrorResponse "1100: could not parse malformed request data"
// @Failure 404 {object} services.BaseErrorResponse "404: record not found"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /courses/catalog/{crseID}/related [get]
func (t *Controller) GetCourseRelated(c *gin.Context) {
	var params CourseRelatedParams
	if err := c.ShouldBindQuery(&params); err != nil {
		t.RespondBadBind(c, err)
		return
	}
	course, err := t.catalog.GetCanonical(c.Param("crseID"))
	if err != nil {
		t.RespondError(c, err)
		return
	}
	resp := CourseRelatedResponse{
		Professors: []RelatedProfessorResponse{},
		Attributes: []RelatedAttributeResponse{},
	}
	cis := relatedTerm(course, params.Strm)
	if cis == nil {
		t.RespondOK(c, resp)
		return
	}
	resp.Strm = cis.Strm
	if resp.Professors, err = t.relatedProfessors(course, cis); err != nil {
		t.RespondError(c, err)
		return
	}
	if resp.Attributes, err = t.relatedAttributes(course, cis); err != nil {
		t.RespondError(c, err)
		return
	}
	t.RespondOK(c, resp)
}

func relatedTerm(course *models.CourseCanonical, strm int) *models.CourseInSemester {
	for _, cis := range course.CoursesInSemester {
		if strm != 0 && cis.Strm == strm {
			return cis
		}
		if strm == 0 && cis.Status == models.CourseOfferingStatusOffered {
			return cis
		}
	}
	return nil
}

func (t *Controller) relatedProfessors(course *models.CourseCanonical, cis *models.CourseInSemester) ([]RelatedProfessorResponse, error) {
	out := []RelatedProfessorResponse{}
	seen := map[uint]bool{}
	for _, o := range cis.Offerings {
		for _, in := range o.Instructors {
			if in.UserID == nil || seen[*in.UserID] {
				continue
			}
			seen[*in.UserID] = true
			taught, err := t.catalog.ProfessorCourses(*in.UserID, models.FactrakReviewWindow(t.now()))
			if err != nil {
				return nil, err
			}
			p := RelatedProfessorResponse{
				Instructor: InstructorResponse{Name: in.NameAsListed, UserID: in.UserID},
				Courses:    make([]ProfessorCourseResponse, 0, relatedPerSection),
			}
			for _, pc := range taught {
				if pc.CrseID == course.CrseID {
					continue
				}
				p.Courses = append(p.Courses, professorCourseResponse(pc))
				if len(p.Courses) == relatedPerSection {
					break
				}
			}
			if len(p.Courses) > 0 {
				out = append(out, p)
			}
		}
	}
	return out, nil
}

func (t *Controller) relatedAttributes(course *models.CourseCanonical, cis *models.CourseInSemester) ([]RelatedAttributeResponse, error) {
	out := []RelatedAttributeResponse{}
	for _, code := range splitAttributes(cis.Attributes) {
		if broadAttribute(code) {
			continue
		}
		year, _ := CurrentTerm(t.now())
		rows, err := t.catalog.LatestCoursesWithAttribute(code, year-requirementRecentYears, maxRelatedAttributeUse)
		if err != nil {
			return nil, err
		}
		courses := make([]TermCourseResponse, 0, len(rows))
		for _, row := range rows {
			if row.CourseCanonicalID != course.ID {
				courses = append(courses, termCourseResponse(row, false))
			}
		}
		if len(courses) == 0 {
			continue
		}
		sort.SliceStable(courses, func(i, j int) bool {
			return firstCode(courses[i].Listings) < firstCode(courses[j].Listings)
		})
		out = append(out, RelatedAttributeResponse{Code: code, Courses: courses[:min(relatedPerSection, len(courses))]})
	}
	return out, nil
}

func broadAttribute(code string) bool {
	for _, p := range broadAttributePrefixes {
		if strings.HasPrefix(code, p) {
			return true
		}
	}
	return false
}

func firstCode(listings []ListingResponse) string {
	if len(listings) == 0 {
		return ""
	}
	return listings[0].Code
}
