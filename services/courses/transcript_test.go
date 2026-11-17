package courses_test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	. "github.com/WilliamsStudentsOnline/wso-go/services/courses"
	"github.com/stretchr/testify/require"
)

func TestImportTranscriptResetsCoveredTermsOnly(t *testing.T) {
	assert, env, _ := seedMine(t)
	code, body := send(t, env, http.MethodPost, "/mine", AddMyCourseRequest{CrseID: "100136", Strm: 1271, ClassNbrs: []int{2001}})
	require.Equal(t, http.StatusCreated, code, string(body))
	code, body = send(t, env, http.MethodPost, "/mine", AddMyCourseRequest{CrseID: "100136", Strm: 1251})
	require.Equal(t, http.StatusCreated, code, string(body))
	code, body = send(t, env, http.MethodPost, "/mine", AddMyCourseRequest{CrseID: "200105"})
	require.Equal(t, http.StatusCreated, code, string(body))

	code, body = send(t, env, http.MethodPut, "/mine/transcript", ImportTranscriptRequest{
		Courses: []TranscriptCourseRef{{CrseID: "100136", Strm: 1271}},
	})
	require.Equal(t, http.StatusOK, code, string(body))
	var res ImportTranscriptResponse
	data(t, assert, body, &res)
	assert.Equal(ImportTranscriptResponse{Through: 1271, Kept: 1, Removed: 1}, res)

	code, body = send(t, env, http.MethodGet, "/mine", nil)
	require.Equal(t, http.StatusOK, code)
	var list []MyCourseResponse
	data(t, assert, body, &list)
	require.Len(t, list, 2)
	assert.Equal(1271, list[0].Term.Strm)
	assert.Len(list[0].Sections, 1)
	assert.Equal("200105", list[1].CrseID)
	assert.Nil(list[1].Term)
}

func TestImportTranscriptUnknownCourse(t *testing.T) {
	_, env, _ := seedMine(t)
	code, _ := send(t, env, http.MethodPut, "/mine/transcript", ImportTranscriptRequest{
		Courses: []TranscriptCourseRef{{CrseID: "100136", Strm: 1263}},
	})
	require.Equal(t, http.StatusNotFound, code)
}

func TestPreviewTranscriptRejectsNonPDF(t *testing.T) {
	assert, env, _ := seedMine(t)
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", "transcript.pdf")
	require.NoError(t, err)
	_, err = part.Write([]byte("not a pdf"))
	require.NoError(t, err)
	require.NoError(t, w.Close())

	req := httptest.NewRequest(http.MethodPost, "/mine/transcript/preview", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	env.Router.ServeHTTP(rec, req)
	assert.Equal(http.StatusBadRequest, rec.Code)
	assert.Contains(rec.Body.String(), lib.ErrorTranscriptUnreadable.Message)
}
