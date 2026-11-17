package courses

import (
	"errors"
	"net/http"
	"sort"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

type MyCourseResponse struct {
	ID       uint              `json:"id"`
	CrseID   string            `json:"crseID"`
	Title    string            `json:"title"`
	Listings []ListingResponse `json:"listings"`
	Term     *CourseTermRef    `json:"term,omitempty"` // empty for a "someday" course
	Hidden   bool              `json:"hidden"`         // listed, but not on the timetable or in distributions
	Sections []SectionResponse `json:"sections"`       // the chosen sections, with meetings for the timetable
	// Components the course offers that term (LEC, LAB, ...) and the ones with no chosen section yet
	Components        []string `json:"components"`
	MissingComponents []string `json:"missingComponents"`
	HasReview         bool     `json:"hasReview"` // the course can't be removed while this is true
	// From the course's term, or for a someday course its newest term
	Div1 bool `json:"div1"`
	Div2 bool `json:"div2"`
	Div3 bool `json:"div3"`
	DPE  bool `json:"dpe"`
	QFR  bool `json:"qfr"`
	WAC  bool `json:"wac"`
}

type ListMyCoursesParams struct {
	Strm int `form:"strm"`
}

type AddMyCourseRequest struct {
	CrseID    string `json:"crseID" binding:"required"`
	Strm      int    `json:"strm"`      // optional; without it the course is "someday"
	ClassNbrs []int  `json:"classNbrs"` // optional sections in that term
	Hidden    *bool  `json:"hidden"`
}

// Omitted fields stay as they are
type UpdateMyCourseRequest struct {
	Hidden    *bool  `json:"hidden"`
	ClassNbrs *[]int `json:"classNbrs"` // replaces the chosen sections; [] clears them
}

func myCourseResponse(uc *models.UserCourse, reviewed map[uint]bool) MyCourseResponse {
	r := MyCourseResponse{
		ID:                uc.ID,
		Hidden:            uc.Hidden,
		Sections:          []SectionResponse{},
		Components:        []string{},
		MissingComponents: []string{},
		HasReview:         reviewed[uc.ID],
	}
	year := 0
	if cis := uc.CourseInSemester; cis != nil {
		ref := termRef(cis)
		r.Term = &ref
		r.Title = cis.Title
		year = cis.Year
	}
	if c := uc.CourseCanonical; c != nil {
		r.CrseID = c.CrseID
		if r.Title == "" {
			r.Title = c.Title
		}
		r.Listings = listingsForYear(c.Listings, year)
	}

	counts := countsFor(uc)
	r.Div1, r.Div2, r.Div3, r.DPE, r.QFR, r.WAC = counts.Div1, counts.Div2, counts.Div3, counts.DPE, counts.QFR, counts.WAC

	chosen := make([]*models.CourseOffering, 0, len(uc.Offerings))
	picked := map[string]bool{}
	for _, uco := range uc.Offerings {
		if uco.CourseOffering != nil {
			chosen = append(chosen, uco.CourseOffering)
			picked[uco.CourseOffering.Component] = true
		}
	}
	sort.Slice(chosen, func(i, j int) bool { return chosen[i].ClassNbr < chosen[j].ClassNbr })
	r.Sections = sectionResponses(chosen)

	if cis := uc.CourseInSemester; cis != nil {
		seen := map[string]bool{}
		for _, o := range cis.Offerings {
			if o.Component == "" || seen[o.Component] {
				continue
			}
			seen[o.Component] = true
			r.Components = append(r.Components, o.Component)
			if !picked[o.Component] {
				r.MissingComponents = append(r.MissingComponents, o.Component)
			}
		}
		sort.Strings(r.Components)
		sort.Strings(r.MissingComponents)
	}
	return r
}

func (t *Controller) respondMyCourse(c *gin.Context, uc *models.UserCourse, status int) {
	reviewed, err := t.mine.ReviewedEntryIDs(uc.UserID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if status == http.StatusCreated {
		t.RespondCreated(c, myCourseResponse(uc, reviewed))
		return
	}
	t.RespondOK(c, myCourseResponse(uc, reviewed))
}

func (t *Controller) respondMyCourseError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, models.ErrUserCourseHasReview):
		t.RespondAPIError(c, lib.ErrorMyCourseHasReview)
	case errors.Is(err, models.ErrUserCourseSection):
		t.RespondAPIError(c, lib.ErrorMyCourseSection)
	default:
		t.RespondError(c, err)
	}
}

// ListMyCourses godoc
// @Summary List my courses
// @Description The signed-in user's courses: taken, current, planned and "someday", newest term first. With strm, only that term (for the timetable). Each course says which components (LEC, LAB, ...) still have no chosen section.
// @ID listMyCourses
// @Tags courses
// @Produce json
// @Param strm query int false "Term code, e.g. 1271"
// @Success 200 {array} courses.MyCourseResponse
// @Failure 400 {object} services.BaseErrorResponse "1100: could not parse malformed request data"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /courses/mine [get]
func (t *Controller) ListMyCourses(c *gin.Context) {
	var params ListMyCoursesParams
	if err := c.ShouldBindQuery(&params); err != nil {
		t.RespondBadBind(c, err)
		return
	}
	userID := services.GetUserID(c)
	list, err := t.mine.ListUserCourses(userID, params.Strm)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	reviewed, err := t.mine.ReviewedEntryIDs(userID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	out := make([]MyCourseResponse, 0, len(list))
	for _, uc := range list {
		out = append(out, myCourseResponse(uc, reviewed))
	}
	t.RespondOK(c, out)
}

// AddMyCourse godoc
// @Summary Add a course to my courses
// @Description Adds a course, optionally in a term (strm) and with sections (classNbrs). If the course is already on the list for that term, the sections are added to it, so a lecture and its lab can be added one at a time.
// @ID addMyCourse
// @Tags courses
// @Accept json
// @Produce json
// @Param body body courses.AddMyCourseRequest true "Course to add"
// @Success 201 {object} courses.MyCourseResponse
// @Failure 400 {object} services.BaseErrorResponse "1100: could not parse malformed request data; 1101: request data validation failed; 2801: section is not part of this course in that term; 2802: sections need a term (strm)"
// @Failure 404 {object} services.BaseErrorResponse "404: record not found"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /courses/mine [post]
func (t *Controller) AddMyCourse(c *gin.Context) {
	var req AddMyCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		t.RespondBadBind(c, err)
		return
	}
	if req.Strm == 0 && len(req.ClassNbrs) > 0 {
		t.RespondAPIError(c, lib.ErrorMyCourseNeedsTerm)
		return
	}
	canonicalID, err := t.catalog.CanonicalID(req.CrseID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	var cis *models.CourseInSemester
	var offerings []*models.CourseOffering
	if req.Strm != 0 {
		cis, offerings, err = t.catalog.CourseOfferingsInTerm(canonicalID, req.Strm, req.ClassNbrs)
		if err != nil {
			t.respondMyCourseError(c, err)
			return
		}
	}
	uc, err := t.mine.AddUserCourse(services.GetUserID(c), canonicalID, cis, offerings, req.Hidden)
	if err != nil {
		t.respondMyCourseError(c, err)
		return
	}
	t.respondMyCourse(c, uc, http.StatusCreated)
}

// UpdateMyCourse godoc
// @Summary Update a course in my courses
// @Description Hides or shows a course, or replaces its chosen sections (class numbers in the course's term).
// @ID updateMyCourse
// @Tags courses
// @Accept json
// @Produce json
// @Param myCourseID path int true "My course ID"
// @Param body body courses.UpdateMyCourseRequest true "Changes"
// @Success 200 {object} courses.MyCourseResponse
// @Failure 400 {object} services.BaseErrorResponse "1100: could not parse malformed request data; 2801: section is not part of this course in that term; 2802: sections need a term (strm)"
// @Failure 404 {object} services.BaseErrorResponse "404: record not found"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /courses/mine/{myCourseID} [patch]
func (t *Controller) UpdateMyCourse(c *gin.Context) {
	id, err := services.GetUIntParam(c, "myCourseID")
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}
	var req UpdateMyCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		t.RespondBadBind(c, err)
		return
	}
	userID := services.GetUserID(c)
	uc, err := t.mine.GetUserCourse(userID, id)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if req.ClassNbrs != nil {
		var offerings []*models.CourseOffering
		if len(*req.ClassNbrs) > 0 {
			if uc.CourseInSemester == nil {
				t.RespondAPIError(c, lib.ErrorMyCourseNeedsTerm)
				return
			}
			_, offerings, err = t.catalog.CourseOfferingsInTerm(uc.CourseCanonicalID, uc.CourseInSemester.Strm, *req.ClassNbrs)
			if err != nil {
				t.respondMyCourseError(c, err)
				return
			}
		}
		if uc, err = t.mine.SetUserCourseOfferings(userID, id, offerings); err != nil {
			t.respondMyCourseError(c, err)
			return
		}
	}
	if req.Hidden != nil {
		if uc, err = t.mine.SetUserCourseHidden(userID, id, *req.Hidden); err != nil {
			t.RespondError(c, err)
			return
		}
	}
	t.respondMyCourse(c, uc, http.StatusOK)
}

// RemoveMySection godoc
// @Summary Remove a section from a course in my courses
// @Description Drops one chosen section (by class number); the course stays on the list.
// @ID removeMyCourseSection
// @Tags courses
// @Produce json
// @Param myCourseID path int true "My course ID"
// @Param classNbr path int true "Class number"
// @Success 200 {object} courses.MyCourseResponse
// @Failure 400 {object} services.BaseErrorResponse "1100: could not parse malformed request data"
// @Failure 404 {object} services.BaseErrorResponse "404: record not found"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /courses/mine/{myCourseID}/sections/{classNbr} [delete]
func (t *Controller) RemoveMySection(c *gin.Context) {
	id, err := services.GetUIntParam(c, "myCourseID")
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}
	classNbr, err := services.GetUIntParam(c, "classNbr")
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}
	userID := services.GetUserID(c)
	uc, err := t.mine.GetUserCourse(userID, id)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	for _, uco := range uc.Offerings {
		if uco.CourseOffering != nil && uco.CourseOffering.ClassNbr == int(classNbr) {
			if uc, err = t.mine.RemoveUserCourseOffering(userID, id, uco.CourseOfferingID); err != nil {
				t.RespondError(c, err)
				return
			}
			t.respondMyCourse(c, uc, http.StatusOK)
			return
		}
	}
	t.RespondAPIError(c, lib.ErrorRecordNotFound)
}

// DeleteMyCourse godoc
// @Summary Remove a course from my courses
// @Description Takes a course (and its chosen sections) off the list. Refused with 409 while you have a Factrak review of the course; delete the review first.
// @ID deleteMyCourse
// @Tags courses
// @Param myCourseID path int true "My course ID"
// @Success 204
// @Failure 400 {object} services.BaseErrorResponse "1100: could not parse malformed request data"
// @Failure 404 {object} services.BaseErrorResponse "404: record not found"
// @Failure 409 {object} services.BaseErrorResponse "2800: delete your review of this course before removing it"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /courses/mine/{myCourseID} [delete]
func (t *Controller) DeleteMyCourse(c *gin.Context) {
	id, err := services.GetUIntParam(c, "myCourseID")
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}
	if err := t.mine.DeleteUserCourse(services.GetUserID(c), id); err != nil {
		t.respondMyCourseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// A cross-listed class counts by the code you register under: AAS 215 is Div II though ENGL 215 is Div I
type entryCounts struct {
	Subject                         string // the registered code's subject, when a section is chosen
	Attributes                      []string
	Div1, Div2, Div3, DPE, QFR, WAC bool
	Known                           bool
}

func countsFor(uc *models.UserCourse) entryCounts {
	var c entryCounts
	attrs := map[string]bool{}
	for _, uco := range uc.Offerings {
		o := uco.CourseOffering
		if o == nil || o.Subject == "" {
			continue
		}
		if !c.Known {
			c.Subject = o.Subject
		}
		c.Known = true
		c.Div1, c.Div2, c.Div3 = c.Div1 || o.Div1, c.Div2 || o.Div2, c.Div3 || o.Div3
		c.DPE, c.QFR, c.WAC = c.DPE || o.DPE, c.QFR || o.QFR, c.WAC || o.WAC
		for _, a := range splitAttributes(o.Attributes) {
			attrs[a] = true
		}
	}
	if c.Known {
		for a := range attrs {
			c.Attributes = append(c.Attributes, a)
		}
		sort.Strings(c.Attributes)
		return c
	}
	cis := uc.CourseInSemester
	if cis == nil && uc.CourseCanonical != nil && len(uc.CourseCanonical.CoursesInSemester) > 0 {
		cis = uc.CourseCanonical.CoursesInSemester[0]
	}
	if cis == nil {
		return c
	}
	return entryCounts{Attributes: splitAttributes(cis.Attributes), Known: true,
		Div1: cis.Div1, Div2: cis.Div2, Div3: cis.Div3, DPE: cis.DPE, QFR: cis.QFR, WAC: cis.WAC}
}
