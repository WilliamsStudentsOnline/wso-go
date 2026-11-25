package courses_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/jobs/catalog_update"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/courses"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestReviewFollowsMyCourses(t *testing.T) {
	assert, env, student := seedMine(t)
	section := func(classNbr int, section, component string) catalog_update.RawCourse {
		return catalog_update.RawCourse{
			AcademicYear: 2025, Offered: "Y", Semester: 1251, CourseID: "100136", Subject: "CSCI",
			CatalogNumber: 136, ClassSection: section, ClassNumber: classNbr, SSRComponent: component,
			CourseTitleLong: "Data Structures", FirstName1: "Bill", LastName1: "Jannen",
		}
	}
	require.NoError(t, catalog_update.IngestCatalog(env.DB, []catalog_update.RawCourse{
		section(1001, "01", "LEC"), section(1002, "L1", "LAB"),
	}, zap.NewNop().Sugar()))
	prof := models.User{Type: models.UserTypeProfessor, Name: "Bill Jannen", UnixID: "bj1"}
	other := models.User{Type: models.UserTypeProfessor, Name: "Someone Else", UnixID: "se1"}
	for _, u := range []*models.User{&prof, &other} {
		require.NoError(t, env.DB.Create(u).Error)
	}
	teach := func(classNbr int, user uint) {
		require.NoError(t, env.DB.Exec(`UPDATE course_offering_instructors SET user_id = ?
WHERE course_offering_id = (SELECT id FROM course_offerings WHERE class_nbr = ?)`, user, classNbr).Error)
	}
	teach(1001, prof.ID)
	teach(1002, other.ID)
	offeringID := func(classNbr int) uint {
		var o models.CourseOffering
		require.NoError(t, env.DB.Where("class_nbr = ?", classNbr).First(&o).Error)
		return o.ID
	}

	dept := models.Department{Name: "CS"}
	require.NoError(t, env.DB.Create(&dept).Error)
	area := models.AreaOfStudy{Name: "Computer Science", Abbreviation: "CSCI", DepartmentID: &dept.ID}
	require.NoError(t, env.DB.Create(&area).Error)
	legacy := models.Course{Number: "136", AreaOfStudyID: &area.ID}
	require.NoError(t, env.DB.Create(&legacy).Error)
	var canonical models.CourseCanonical
	require.NoError(t, env.DB.Where("crse_id = ?", "100136").First(&canonical).Error)
	newReview := func(year int) models.FactrakSurvey {
		fall := models.FactrakSurveySemesterSeasonFall
		r := models.FactrakSurvey{UserID: student.ID, ProfessorID: prof.ID, CourseID: legacy.ID, CourseCanonicalID: &canonical.ID,
			CourseLinkVerificationSource: models.CourseLinkVerificationSourceConfirmed, SemesterSeason: &fall, SemesterYear: &year, Comment: "Good."}
		require.NoError(t, env.DB.Create(&r).Error)
		return r
	}
	review, older := newReview(2024), newReview(2022)
	reload := func(r models.FactrakSurvey) models.FactrakSurvey {
		var out models.FactrakSurvey
		require.NoError(t, env.DB.First(&out, r.ID).Error)
		return out
	}

	code, body := send(t, env, http.MethodPost, "/mine", AddMyCourseRequest{CrseID: "100136", Strm: 1251})
	require.Equal(t, http.StatusCreated, code, string(body))
	var mc MyCourseResponse
	data(t, assert, body, &mc)
	assert.True(mc.HasReview)
	require.NotNil(t, reload(review).UserCourseID)
	assert.Equal(mc.ID, *reload(review).UserCourseID)
	assert.Nil(reload(older).UserCourseID)
	code, body = send(t, env, http.MethodDelete, fmt.Sprintf("/mine/%d", mc.ID), nil)
	assert.Equal(http.StatusConflict, code)
	assert.Contains(string(body), fmt.Sprint(lib.ErrorMyCourseHasReview.Code))

	sections := func(classNbrs ...int) {
		if classNbrs == nil {
			classNbrs = []int{} // clear them, not leave them
		}
		code, body := send(t, env, http.MethodPatch, fmt.Sprintf("/mine/%d", mc.ID), UpdateMyCourseRequest{ClassNbrs: &classNbrs})
		require.Equal(t, http.StatusOK, code, string(body))
	}
	// A section the professor didn't teach leaves the review alone
	sections(1002)
	assert.Nil(reload(review).CourseOfferingID)
	assert.Equal(models.CourseLinkVerificationSourceConfirmed, reload(review).CourseLinkVerificationSource)
	// One they taught places it, as the author's call
	sections(1002, 1001)
	require.NotNil(t, reload(review).CourseOfferingID)
	assert.Equal(offeringID(1001), *reload(review).CourseOfferingID)
	assert.Equal(models.CourseLinkVerificationSourceAuthorConfirmed, reload(review).CourseLinkVerificationSource)

	// Moves the review to another section the professor taught, else unplaces it
	teach(1002, prof.ID)
	code, body = send(t, env, http.MethodDelete, fmt.Sprintf("/mine/%d/sections/1001", mc.ID), nil)
	require.Equal(t, http.StatusOK, code, string(body))
	require.NotNil(t, reload(review).CourseOfferingID)
	assert.Equal(offeringID(1002), *reload(review).CourseOfferingID)
	sections()
	assert.Nil(reload(review).CourseOfferingID)
}
