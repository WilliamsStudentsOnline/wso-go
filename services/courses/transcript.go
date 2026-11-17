package courses

// The PDF is read in memory and dropped, never stored

import (
	"io"
	"strconv"
	"strings"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/transcriptpdf"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type TranscriptRow struct {
	Code      string `json:"code"`  // as printed, e.g. "CSCI 136"
	Title     string `json:"title"` // as printed, abbreviations and all
	Strm      int    `json:"strm"`
	TermLabel string `json:"termLabel"` // e.g. "Fall 2026"
	// The catalog course it matched in that term; empty when our catalog doesn't have it
	CrseID       string `json:"crseID,omitempty"`
	CatalogTitle string `json:"catalogTitle,omitempty"`
	Status       string `json:"status"` // add, keep (already on the list) or unmatched
}

type TranscriptPreviewResponse struct {
	Format  string          `json:"format"`  // wmsr102a (progress report) or wmsr102b (transcript)
	Through int             `json:"through"` // the last term on it; later terms aren't touched
	Rows    []TranscriptRow `json:"rows"`
	// Courses in those terms (or someday courses now taken) that would come off; reviewed ones stay
	Remove   []MyCourseResponse `json:"remove"`
	Reviewed []MyCourseResponse `json:"reviewed"`
}

type TranscriptCourseRef struct {
	CrseID string `json:"crseID" binding:"required"`
	Strm   int    `json:"strm" binding:"required"`
}

type ImportTranscriptRequest struct {
	Courses []TranscriptCourseRef `json:"courses" binding:"required,min=1,max=200,dive"`
}

type ImportTranscriptResponse struct {
	Through  int `json:"through"`
	Added    int `json:"added"`
	Kept     int `json:"kept"`
	Removed  int `json:"removed"`
	Reviewed int `json:"reviewed"` // not on the transcript, but kept for their reviews
}

func strmLabel(strm int) string {
	year, _ := catalogYearOfStrm(strm)
	return TermLabel(year, termNameOfStrm(strm))
}

func splitNumber(s string) (int, string) {
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	n, _ := strconv.Atoi(s[:i])
	return n, s[i:]
}

// PreviewTranscript godoc
// @Summary Preview importing a transcript PDF
// @Description Reads a Williams academic progress report or internal transcript PDF (one page, at most 2 MB) and matches its courses to the catalog. Nothing is saved, the PDF included. The response says which courses would be added, which are already on the list, which we couldn't match, and which courses in the transcript's terms would come off (reviewed courses never do). Import with PUT /courses/mine/transcript.
// @ID previewMyTranscript
// @Tags courses
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Progress report or transcript PDF"
// @Success 200 {object} courses.TranscriptPreviewResponse
// @Failure 400 {object} services.BaseErrorResponse "2820: couldn't read that PDF: upload a one-page Williams academic progress report or internal transcript"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /courses/mine/transcript/preview [post]
func (t *Controller) PreviewTranscript(c *gin.Context) {
	formFile, err := c.FormFile("file")
	if err != nil {
		t.RespondAPIError(c, lib.ErrorTranscriptUnreadable)
		return
	}
	if formFile.Size > transcriptpdf.MaxBytes {
		t.RespondAPIError(c, lib.ErrorTranscriptUnreadable)
		return
	}
	file, err := formFile.Open()
	if err != nil {
		t.RespondAPIError(c, lib.ErrorTranscriptUnreadable)
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, transcriptpdf.MaxBytes+1))
	if err != nil {
		t.RespondAPIError(c, lib.ErrorTranscriptUnreadable)
		return
	}
	report, err := transcriptpdf.Read(data)
	if err != nil {
		t.RespondAPIError(c, lib.ErrorTranscriptUnreadable)
		return
	}

	rows := make([]TranscriptRow, 0, len(report.Courses))
	var matched []*models.CourseInSemester
	for _, pc := range report.Courses {
		row := TranscriptRow{
			Code:      pc.Subject + " " + pc.Number,
			Title:     pc.Title,
			Strm:      pc.Strm(),
			TermLabel: strmLabel(pc.Strm()),
			Status:    "unmatched",
		}
		number, letter := splitNumber(pc.Number)
		cis, err := t.catalog.TermCourseByCode(pc.Subject, number, strings.TrimSuffix(letter, "T"), row.Strm)
		switch {
		case gorm.IsRecordNotFoundError(err):
		case err != nil:
			t.RespondError(c, err)
			return
		default:
			matched = append(matched, cis)
			if cis.CourseCanonical != nil {
				row.CrseID = cis.CourseCanonical.CrseID
			}
			row.CatalogTitle = cis.Title
		}
		rows = append(rows, row)
	}

	userID := services.GetUserID(c)
	out := TranscriptPreviewResponse{Format: report.Format, Rows: rows, Remove: []MyCourseResponse{}, Reviewed: []MyCourseResponse{}}
	if len(matched) > 0 {
		plan, err := t.mine.ImportTranscript(userID, matched, true)
		if err != nil {
			t.RespondError(c, err)
			return
		}
		out.Through = plan.Through
		kept := map[uint]bool{}
		for _, uc := range plan.Keep {
			kept[*uc.CourseInSemesterID] = true
		}
		mi := 0
		for i := range out.Rows {
			if out.Rows[i].CrseID == "" {
				continue
			}
			if kept[matched[mi].ID] {
				out.Rows[i].Status = "keep"
			} else {
				out.Rows[i].Status = "add"
			}
			mi++
		}
		reviewed, err := t.mine.ReviewedEntryIDs(userID)
		if err != nil {
			t.RespondError(c, err)
			return
		}
		for _, uc := range plan.Remove {
			out.Remove = append(out.Remove, myCourseResponse(uc, reviewed))
		}
		for _, uc := range plan.Reviewed {
			out.Reviewed = append(out.Reviewed, myCourseResponse(uc, reviewed))
		}
	}
	t.RespondOK(c, out)
}

// ImportTranscript godoc
// @Summary Import a transcript into my courses
// @Description Resets every term up to the last one given to exactly these courses (from a transcript preview): they're added, or kept with their sections, and other courses in those terms come off, as do someday entries for courses now taken. Later terms are left alone, and courses you reviewed are never removed.
// @ID importMyTranscript
// @Tags courses
// @Accept json
// @Produce json
// @Param body body courses.ImportTranscriptRequest true "Matched courses"
// @Success 200 {object} courses.ImportTranscriptResponse
// @Failure 400 {object} services.BaseErrorResponse "1100: could not parse malformed request data; 1101: request data validation failed"
// @Failure 404 {object} services.BaseErrorResponse "404: record not found"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /courses/mine/transcript [put]
func (t *Controller) ImportTranscript(c *gin.Context) {
	var req ImportTranscriptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		t.RespondBadBind(c, err)
		return
	}
	courses := make([]*models.CourseInSemester, 0, len(req.Courses))
	for _, ref := range req.Courses {
		canonicalID, err := t.catalog.CanonicalID(ref.CrseID)
		if err != nil {
			t.RespondError(c, err)
			return
		}
		cis, _, err := t.catalog.CourseOfferingsInTerm(canonicalID, ref.Strm, nil)
		if err != nil {
			t.RespondError(c, err)
			return
		}
		courses = append(courses, cis)
	}
	plan, err := t.mine.ImportTranscript(services.GetUserID(c), courses, false)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	t.RespondOK(c, ImportTranscriptResponse{
		Through:  plan.Through,
		Added:    len(plan.Add),
		Kept:     len(plan.Keep),
		Removed:  len(plan.Remove),
		Reviewed: len(plan.Reviewed),
	})
}
