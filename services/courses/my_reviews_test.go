package courses_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/courses"
	"github.com/stretchr/testify/require"
)

func TestMyReviewsSuggestAndLink(t *testing.T) {
	assert, env, student := seedMine(t)

	prof := models.User{Type: models.UserTypeProfessor, Name: "Bill Jannen", UnixID: "bj1"}
	require.NoError(t, env.DB.Create(&prof).Error)
	other := models.User{Type: models.UserTypeProfessor, Name: "Someone Else", UnixID: "se1"}
	require.NoError(t, env.DB.Create(&other).Error)
	setInstructor := func(classNbr int, user uint) {
		require.NoError(t, env.DB.Exec(`UPDATE course_offering_instructors SET user_id = ?
WHERE course_offering_id = (SELECT id FROM course_offerings WHERE class_nbr = ?)`, user, classNbr).Error)
	}
	setInstructor(1001, prof.ID)
	setInstructor(2001, prof.ID)
	setInstructor(2002, prof.ID)

	dept := models.Department{Name: "CS"}
	require.NoError(t, env.DB.Create(&dept).Error)
	area := models.AreaOfStudy{Name: "Computer Science", Abbreviation: "CSCI", DepartmentID: &dept.ID}
	require.NoError(t, env.DB.Create(&area).Error)
	legacy := models.Course{Number: "136", AreaOfStudyID: &area.ID}
	require.NoError(t, env.DB.Create(&legacy).Error)
	var canonical models.CourseCanonical
	require.NoError(t, env.DB.Where("crse_id = ?", "100136").First(&canonical).Error)
	spring := models.FactrakSurveySemesterSeasonSpring
	review := models.FactrakSurvey{UserID: student.ID, ProfessorID: prof.ID, CourseID: legacy.ID,
		CourseCanonicalID: &canonical.ID, CourseLinkVerificationSource: models.CourseLinkVerificationSourceConfirmed, Comment: "Good.",
		SemesterSeason: &spring, SemesterYear: lib.IntToPtr(2025)}
	require.NoError(t, env.DB.Create(&review).Error)

	code, body := send(t, env, http.MethodGet, "/mine/reviews", nil)
	require.Equal(t, http.StatusOK, code, string(body))
	var mine []MyReviewResponse
	data(t, assert, body, &mine)
	require.Len(t, mine, 1)
	assert.Equal("CSCI 136", mine[0].Course)
	assert.Equal("100136", mine[0].CrseID)
	assert.Equal("Bill Jannen", mine[0].Professor)
	assert.Equal("Spring 2025", mine[0].Semester)
	assert.Nil(mine[0].Linked)
	require.Len(t, mine[0].Suggestions, 3)
	assert.Equal(2001, mine[0].Suggestions[0].ClassNbr)
	assert.Equal(1001, mine[0].Suggestions[2].ClassNbr)
	assert.Equal("Fall 2024", mine[0].Suggestions[2].TermLabel)

	link := func(strm, classNbr int) (int, []byte) {
		return send(t, env, http.MethodPost, fmt.Sprintf("/mine/reviews/%d/link", review.ID),
			LinkMyReviewRequest{Strm: strm, ClassNbr: classNbr})
	}
	code, body = link(1201, 1089)
	assert.Equal(http.StatusBadRequest, code)
	assert.Contains(string(body), fmt.Sprint(lib.ErrorReviewOtherCourse.Code))

	setInstructor(1001, other.ID)
	code, body = link(1251, 1001)
	assert.Equal(http.StatusBadRequest, code)
	assert.Contains(string(body), fmt.Sprint(lib.ErrorSurveySectionProfessor.Code))
	setInstructor(1001, prof.ID)

	code, body = link(1251, 1001)
	require.Equal(t, http.StatusOK, code, string(body))
	var linked MyReviewResponse
	data(t, assert, body, &linked)
	require.NotNil(t, linked.Linked)
	assert.Equal(1001, linked.Linked.ClassNbr)
	assert.Equal(models.CourseLinkVerificationSourceAuthorConfirmed, linked.CourseLinkVerificationSource)

	var saved models.FactrakSurvey
	require.NoError(t, env.DB.First(&saved, review.ID).Error)
	require.NotNil(t, saved.SemesterSeason)
	assert.Equal(models.FactrakSurveySemesterSeasonFall, *saved.SemesterSeason)
	assert.Equal(2024, *saved.SemesterYear)
	assert.NotNil(saved.UserCourseID) // the My courses entry the section went on

	code, body = send(t, env, http.MethodGet, "/mine", nil)
	require.Equal(t, http.StatusOK, code)
	var list []MyCourseResponse
	data(t, assert, body, &list)
	require.Len(t, list, 1)
	assert.True(list[0].HasReview)
	require.Len(t, list[0].Sections, 1)
	assert.Equal(1001, list[0].Sections[0].ClassNbr)
}

func TestMyReviewsOwnerOnly(t *testing.T) {
	assert, env, _ := seedMine(t)
	someone := models.User{Type: models.UserTypeStudent, Name: "Other", UnixID: "ot1"}
	require.NoError(t, env.DB.Create(&someone).Error)
	prof := models.User{Type: models.UserTypeProfessor, Name: "P", UnixID: "p1"}
	require.NoError(t, env.DB.Create(&prof).Error)
	dept := models.Department{Name: "CS"}
	require.NoError(t, env.DB.Create(&dept).Error)
	area := models.AreaOfStudy{Name: "Computer Science", Abbreviation: "CSCI", DepartmentID: &dept.ID}
	require.NoError(t, env.DB.Create(&area).Error)
	legacy := models.Course{Number: "136", AreaOfStudyID: &area.ID}
	require.NoError(t, env.DB.Create(&legacy).Error)
	theirs := models.FactrakSurvey{UserID: someone.ID, ProfessorID: prof.ID, CourseID: legacy.ID}
	require.NoError(t, env.DB.Create(&theirs).Error)

	code, body := send(t, env, http.MethodGet, "/mine/reviews", nil)
	require.Equal(t, http.StatusOK, code)
	var mine []MyReviewResponse
	data(t, assert, body, &mine)
	assert.Empty(mine)
	code, _ = send(t, env, http.MethodPost, fmt.Sprintf("/mine/reviews/%d/link", theirs.ID), LinkMyReviewRequest{Strm: 1271, ClassNbr: 2001})
	assert.Equal(http.StatusNotFound, code)
}
