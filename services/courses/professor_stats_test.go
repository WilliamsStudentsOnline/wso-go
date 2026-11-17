package courses_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/courses"
	"github.com/stretchr/testify/require"
)

func TestProfessorStats(t *testing.T) {
	env, teach, _ := seedProfessors(t, auth.ScopeAllOther, auth.ScopeFactrakFull)
	assert := env.Assert
	var student models.User
	require.NoError(t, env.DB.Where("unix_id = ?", "sd1").First(&student).Error)
	var legacy models.Course
	require.NoError(t, env.DB.First(&legacy).Error)
	review := func(prof uint, lead int, season string, year int) {
		s := models.FactrakSurvey{UserID: student.ID, ProfessorID: prof, CourseID: legacy.ID, LeadLecture: &lead, Comment: "review"}
		if season != "" {
			s.SemesterSeason, s.SemesterYear = &season, &year
		}
		require.NoError(t, env.DB.Create(&s).Error)
	}
	get := func() ProfessorStatsResponse {
		w, err := utils.DoHTTPReq(env.Router, http.MethodGet, fmt.Sprintf("/professors/%d/stats", teach.ID), nil)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var resp ProfessorStatsResponse
		data(t, assert, w.Body.Bytes(), &resp)
		return resp
	}
	question := func(resp ProfessorStatsResponse, name string) ProfessorQuestionStats {
		for _, q := range resp.Questions {
			if q.Question == name {
				return q
			}
		}
		t.Fatalf("no %s", name)
		return ProfessorQuestionStats{}
	}

	// A colleague with five reviews counts, one with four doesn't
	peer := func(name, unix string, n, lead int) {
		p := models.User{Type: models.UserTypeProfessor, Name: name, UnixID: unix, DepartmentID: teach.DepartmentID}
		require.NoError(t, env.DB.Create(&p).Error)
		for i := 0; i < n; i++ {
			review(p.ID, lead, "", 0)
		}
	}
	peer("Peer Five", "pf1", 5, 3)
	peer("Peer Four", "pf4", 4, 1)

	resp := get()
	assert.Equal("Computer Science", resp.Department)
	rec := question(resp, "wouldRecommendCourse")
	assert.True(rec.Percent)
	assert.Equal(2, rec.Count)
	assert.InDelta(50, rec.Value, 0.01)
	assert.Nil(rec.Department)

	review(teach.ID, 4, models.FactrakSurveySemesterSeasonFall, 2025)
	review(teach.ID, 4, models.FactrakSurveySemesterSeasonFall, 2025)
	review(teach.ID, 2, models.FactrakSurveySemesterSeasonSpring, 2026)
	lead := question(get(), "leadLecture")
	assert.False(lead.Percent)
	assert.Equal(5, lead.Count)
	assert.InDelta(4.0, lead.Value, 0.01) // (5+5+4+4+2)/5
	assert.Equal([]float64{3}, lead.Department)
	require.Len(t, lead.Terms, 2)
	assert.Equal(ProfessorTermStat{Strm: 1261, Label: "Fall 2025", Value: 4, Count: 2}, lead.Terms[0])
	assert.Equal(ProfessorTermStat{Strm: 1263, Label: "Spring 2026", Value: 2, Count: 1}, lead.Terms[1])
}

func TestCourseStats(t *testing.T) {
	env, teach, gone := seedProfessors(t, auth.ScopeAllOther, auth.ScopeFactrakFull)
	assert := env.Assert
	var student models.User
	require.NoError(t, env.DB.Where("unix_id = ?", "sd1").First(&student).Error)
	var legacy models.Course
	require.NoError(t, env.DB.First(&legacy).Error)
	var course models.CourseCanonical
	require.NoError(t, env.DB.Where("crse_id = ?", "100136").First(&course).Error)
	review := func(courseID, prof uint, lead int, source string) {
		s := models.FactrakSurvey{UserID: student.ID, ProfessorID: prof, CourseID: legacy.ID, CourseCanonicalID: &courseID, LeadLecture: &lead, Comment: "review", CourseLinkVerificationSource: source}
		require.NoError(t, env.DB.Create(&s).Error)
	}
	peer := func(crseID string, number, n, lead int) {
		c := models.CourseCanonical{CrseID: crseID, Title: crseID}
		require.NoError(t, env.DB.Create(&c).Error)
		require.NoError(t, env.DB.Create(&models.CourseListing{CourseCanonicalID: c.ID, Subject: "CSCI", Number: number, FirstValidYear: 2000, LastValidYear: 2100}).Error)
		for i := 0; i < n; i++ {
			review(c.ID, gone.ID, lead, "")
		}
	}
	peer("100256", 256, 3, 3)
	peer("100334", 334, 2, 1)
	for i := 0; i < 3; i++ {
		review(course.ID, gone.ID, 2, "")
	}
	review(course.ID, gone.ID, 1, models.CourseLinkVerificationSourceMismatch)

	get := func(q string) CourseQuestionStats {
		w, err := utils.DoHTTPReq(env.Router, http.MethodGet, "/catalog/100136/stats"+q, nil)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var resp CourseStatsResponse
		data(t, assert, w.Body.Bytes(), &resp)
		if q == "" {
			assert.Equal("Computer Science", resp.Department)
		}
		for _, s := range resp.Questions {
			if s.Question == "leadLecture" {
				return s
			}
		}
		t.Fatal("no leadLecture")
		return CourseQuestionStats{}
	}

	all := get("")
	assert.Equal(5, all.Count)
	assert.InDelta(3.2, all.Value, 0.01)
	assert.Equal([]float64{3}, all.Department)

	one := get(fmt.Sprintf("?professorID=%d", teach.ID))
	assert.Equal(2, one.Count)
	assert.InDelta(5.0, one.Value, 0.01)
	assert.Equal([]float64{2}, one.Department)
}
