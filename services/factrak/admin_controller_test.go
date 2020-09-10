package factrak_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/factrak"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestController_ListFlaggedSurveys(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf, auth.ScopeFactrakAdmin)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	c1 := models.Course{
		Number: "c1",
		AreaOfStudy: &models.AreaOfStudy{
			Name:         "Computer Science",
			Abbreviation: "CSCI",
			Department: &models.Department{
				Name: "Computer Science",
			},
		},
	}

	p1 := models.User{
		Type:   models.UserTypeProfessor,
		Name:   "Professor 1",
		UnixID: "p1",
	}
	s1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "Student",
		UnixID: "s1",
	}
	assert.NoError(db.Create(&p1).Create(&s1).Error)

	fs1 := models.FactrakSurvey{
		User:      &s1,
		Professor: &p1,
		Course:    &c1,
		Comment:   "Survey 1",
		Flagged:   true,
	}
	fs2 := models.FactrakSurvey{
		User:      &s1,
		Professor: &p1,
		Course:    &c1,
		Comment:   "Survey 2",
	}
	fs3 := models.FactrakSurvey{
		User:      &s1,
		Professor: &p1,
		Course:    &c1,
		Comment:   "Survey 2",
		Flagged:   true,
	}

	assert.NoError(db.Create(&fs1).Create(&fs2).Create(&fs3).Error)

	// Test 1: Get flagged surveys, when unauthorized (should fail)
	noAdminR := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf)
	SetupRouter(noAdminR, db, cfg, zaptest.NewLogger(t).Sugar())
	w, err := utils.DoHTTPReq(noAdminR, http.MethodGet, "/admin/surveys", nil)
	assert.NoError(err)
	assert.Equal(http.StatusForbidden, w.Code)

	// Test 2: Get flagged surveys, when authorized (should succeed)
	w, err = utils.DoHTTPReq(router, http.MethodGet, "/admin/surveys", nil)
	assert.NoError(err)

	// Check if correct surveys (reverse chronological order)
	resp := GetSurveysFromResp(assert, w)
	assert.NoError(EqualSurveyIDs([]models.FactrakSurvey{fs3, fs1}, resp))
}

func TestController_UnflagSurvey(t *testing.T) {
	// Setup
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	// Populate the database
	s1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "Student 1",
		UnixID: "s1",
	}
	survey := models.FactrakSurvey{
		User: &s1,
		Professor: &models.User{
			Type:   models.UserTypeProfessor,
			Name:   "Professor 1",
			UnixID: "p1",
		},
		Course: &models.Course{
			Number: "c1",
			AreaOfStudy: &models.AreaOfStudy{
				Name:         "Computer Science",
				Abbreviation: "CSCI",
				Department: &models.Department{
					Name: "Computer Science",
				},
			},
		},
		Comment: generateSurveyTestComment(),
		Flagged: true,
	}
	assert.NoError(db.Create(&s1).Create(&survey).Error)

	// Setup router
	router := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf, auth.ScopeFactrakAdmin)
	utils.AddUserContexts(router, s1.ID)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	// First, we run tests on validations

	// Test 1: Fail on no admin
	noAdminR := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf)
	SetupRouter(noAdminR, db, cfg, zaptest.NewLogger(t).Sugar())
	w, err := utils.DoHTTPReq(noAdminR, http.MethodDelete,
		fmt.Sprintf("/admin/surveys/%d/flag", survey.ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusForbidden, w.Code)

	// Test 2: error on bad survey
	w, err = utils.DoHTTPReq(router, http.MethodDelete, fmt.Sprintf("/admin/surveys/%d/flag", 42), nil)
	assert.NoError(err)
	// Assert correct error
	apiErr := lib.ErrorRecordNotFound
	utils.CheckRespError(assert, w, apiErr)

	// Test 3: actually remove flag
	w, err = utils.DoHTTPReq(router, http.MethodDelete, fmt.Sprintf("/admin/surveys/%d/flag", survey.ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)

	// Check to make sure it updated in the db
	// Assert for database entry
	var surveyInDB models.FactrakSurvey
	assert.NoError(db.First(&surveyInDB, survey.ID).Error)
	// Assertions
	assert.False(surveyInDB.Flagged)
}
