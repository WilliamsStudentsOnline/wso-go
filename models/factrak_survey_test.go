package models_test

import (
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	. "github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestFactrakSurveyModel_GetSurveyRatingsByProfessorOrCourse(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewFactrakSurveyModel(db, zaptest.NewLogger(t).Sugar())

	student := User{
		Type:   UserTypeStudent,
		UnixID: "s1",
		Name:   "Student 1",
	}

	course := Course{
		Number: "256",
		AreaOfStudy: &AreaOfStudy{
			Name:         "Computer Science",
			Abbreviation: "CSCI",
			Department: &Department{
				Name: "Computer Science",
			},
		},
	}

	prof := User{
		Type:   UserTypeProfessor,
		UnixID: "p1",
		Name:   "Professor 1",
	}

	surveys := []*FactrakSurvey{
		{
			User:                 &student,
			Course:               &course,
			Professor:            &prof,
			WouldRecommendCourse: lib.BoolToPtr(true),
			CourseWorkload:       lib.IntToPtr(5),
			CourseStimulating:    lib.IntToPtr(3),
			WouldTakeAnother:     lib.BoolToPtr(true),
			Approachability:      lib.IntToPtr(7),
			LeadLecture:          lib.IntToPtr(2),
			PromoteDiscussion:    lib.IntToPtr(5),
			OutsideHelpfulness:   lib.IntToPtr(6),
		},
		{
			User:                 &student,
			Course:               &course,
			Professor:            &prof,
			WouldRecommendCourse: lib.BoolToPtr(false),
			CourseWorkload:       lib.IntToPtr(5),
			CourseStimulating:    lib.IntToPtr(1),
			WouldTakeAnother:     lib.BoolToPtr(false),
			Approachability:      lib.IntToPtr(1),
			LeadLecture:          lib.IntToPtr(4),
			PromoteDiscussion:    lib.IntToPtr(2),
			OutsideHelpfulness:   lib.IntToPtr(6),
		},
		{
			User:                 &student,
			Course:               &course,
			Professor:            &prof,
			WouldRecommendCourse: lib.BoolToPtr(true),
			CourseWorkload:       lib.IntToPtr(4),
			CourseStimulating:    lib.IntToPtr(6),
			WouldTakeAnother:     lib.BoolToPtr(false),
			Approachability:      lib.IntToPtr(6),
			LeadLecture:          lib.IntToPtr(4),
			PromoteDiscussion:    lib.IntToPtr(5),
			OutsideHelpfulness:   lib.IntToPtr(2),
		},
	}

	assert.NoError(m.DB.Create(&student).Create(&course).Create(&prof).Error)

	for i := range surveys {
		assert.NoError(db.Create(&surveys[i]).Error)
	}

	var res FactrakSurveyAvgRatings
	assert.NoError(m.GetSurveyRatingsByProfessorOrCourse(&prof.ID, nil, nil, &res))

	assert.Equal(float64(2)/float64(3), res.AvgWouldRecommendCourse)
	assert.Equal(float64(5+5+4)/float64(3), res.AvgCourseWorkload)
	assert.Equal(float64(3+1+6)/float64(3), res.AvgCourseStimulating)
	assert.Equal(float64(1)/float64(3), res.AvgWouldTakeAnother)
	assert.Equal(float64(7+1+6)/float64(3), res.AvgApproachability)
	assert.Equal(float64(2+4+4)/float64(3), res.AvgLeadLecture)
	assert.Equal(float64(5+2+5)/float64(3), res.AvgPromoteDiscussion)
	assert.Equal(float64(6+6+2)/float64(3), res.AvgOutsideHelpfulness)
}
