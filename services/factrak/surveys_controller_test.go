package factrak_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/factrak"
	"github.com/gin-gonic/gin"
	testify "github.com/stretchr/testify/assert"
)

func TestController_ListSurveys(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf)
	SetupRouter(router, db)

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
		Type:       models.UserTypeProfessor,
		Name:       "Professor 1",
		UnixID:     "p1",
		Visible:    true,
		AtWilliams: true,
	}
	s1 := models.User{
		Type:       models.UserTypeStudent,
		Name:       "Student",
		UnixID:     "s1",
		Visible:    true,
		AtWilliams: true,
	}
	assert.NoError(db.Create(&p1).Create(&s1).Error)

	fs1 := models.FactrakSurvey{
		User:      &s1,
		Professor: &p1,
		Course:    &c1,
		Comment:   "Survey 1",
	}
	fs2 := models.FactrakSurvey{
		User:      &s1,
		Professor: &p1,
		Course:    &c1,
		Comment:   "Survey 2",
	}

	assert.NoError(db.Create(&fs1).Create(&fs2).Error)

	// Get test surveys
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/surveys", nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp []models.FactrakSurvey
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct surveys (ordered by date)
	assert.Len(resp, 2)
	assert.Equal(fs2.Comment, resp[0].Comment)
	assert.Equal(fs1.Comment, resp[1].Comment)

	// Make sure anonymous
	assert.Zero(resp[0].UserID)
	assert.Nil(resp[0].User)
}

func TestController_GetSurvey(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf)
	SetupRouter(router, db)

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
		Type:       models.UserTypeProfessor,
		Name:       "Professor 1",
		UnixID:     "p1",
		Visible:    true,
		AtWilliams: true,
	}
	s1 := models.User{
		Type:       models.UserTypeStudent,
		Name:       "Student",
		UnixID:     "s1",
		Visible:    true,
		AtWilliams: true,
	}
	assert.NoError(db.Create(&p1).Create(&s1).Error)

	fs1 := models.FactrakSurvey{
		User:      &s1,
		Professor: &p1,
		Course:    &c1,
		Comment:   "Survey 1",
	}
	fs2 := models.FactrakSurvey{
		User:      &s1,
		Professor: &p1,
		Course:    &c1,
		Comment:   "Survey 2",
	}

	assert.NoError(db.Create(&fs1).Create(&fs2).Error)

	// Get test survey 1
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/surveys/%d", fs1.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.FactrakSurvey
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct surveys
	assert.Equal(fs1.Comment, resp.Comment)

	// Make sure anonymous
	assert.Zero(resp.UserID)
	assert.Nil(resp.User)

	// Get bad survey (expect failure)
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/surveys/%d", 42), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusNotFound, w.Code)
}

func TestController_CreateSurvey(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	// Populate the database
	s1 := models.User{
		Type:       models.UserTypeStudent,
		Name:       "Student 1",
		UnixID:     "s1",
		Visible:    true,
		AtWilliams: true,
	}
	s2 := models.User{
		Type:       models.UserTypeStudent,
		Name:       "Student 2",
		UnixID:     "s2",
		Visible:    true,
		AtWilliams: true,
	}
	p1 := models.User{
		Type:       models.UserTypeProfessor,
		Name:       "Professor 1",
		UnixID:     "p1",
		Visible:    true,
		AtWilliams: true,
	}
	p2 := models.User{
		Type:       models.UserTypeProfessor,
		Name:       "Professor 2",
		UnixID:     "p2",
		Visible:    true,
		AtWilliams: true,
	}
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
	c2 := models.Course{
		Number: "c2",
		AreaOfStudy: &models.AreaOfStudy{
			Name:         "Economics",
			Abbreviation: "ECON",
			Department: &models.Department{
				Name: "Economics",
			},
		},
	}
	a1 := models.AreaOfStudy{
		Name:         "Mathematics",
		Abbreviation: "MATH",
		Department: &models.Department{
			Name: "Mathematics & Statistics",
		},
	}
	assert.NoError(db.Create(&s1).Create(&s2).Create(&p1).Create(&p2).Create(&c1).Create(&c2).Create(&a1).Error)

	// Setup router
	router := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf)
	utils.AddUserContexts(router, s1.ID)
	SetupRouter(router, db)

	// First, we run tests on validations

	// Test 1: error on missing course parameters
	params := SurveyCreateParams{ProfessorID: &p1.ID, Comment: "c"}
	createSurveyExpectError(assert, router, params, lib.ErrorSurveyMissingCourseParams)

	// Test 2: error on missing course parameters (where course number is empty)
	params = SurveyCreateParams{ProfessorID: &p1.ID, Comment: "c",
		AreaOfStudyAbbreviation: lib.StrToPtr(""), CourseNumber: lib.StrToPtr("")}
	createSurveyExpectError(assert, router, params, lib.ErrorSurveyMissingCourseParams)

	// Test 3: error on small survey comment
	params = SurveyCreateParams{ProfessorID: &p1.ID, CourseID: &c1.ID,
		Comment: "comment is too small"}
	createSurveyExpectError(assert, router, params, lib.ErrorSurveyCommentTooSmall)

	// Test 4: error on bad student
	params = SurveyCreateParams{ProfessorID: &p1.ID, CourseID: &c1.ID, Comment: generateSurveyTestComment()}
	// Setup bad student router
	r1 := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf)
	utils.AddUserContexts(r1, p2.ID)
	SetupRouter(r1, db)
	createSurveyExpectError(assert, r1, params, lib.ErrorSurveyStudentNotFound)

	// Test 5: error on student prefrosh
	assert.NoError(db.Model(&s1).Update("class_year", models.NewStudentModel(db).SeniorYear()+4).Error)
	params = SurveyCreateParams{CourseID: &c1.ID, Comment: generateSurveyTestComment(), ProfessorID: &p1.ID}
	createSurveyExpectError(assert, router, params, lib.ErrorUserCannotBePrefrosh)
	// Revert
	assert.NoError(db.Model(&s1).Update("class_year", nil).Error)

	// Test 6: error on professor not at williams
	assert.NoError(db.Model(&p2).Update("at_williams", false).Error)
	params = SurveyCreateParams{CourseID: &c1.ID, Comment: generateSurveyTestComment(),
		ProfessorID: &p2.ID}
	createSurveyExpectError(assert, router, params, lib.ErrorSurveyProfessorNotFound)
	// Revert
	assert.NoError(db.Model(&p2).Update("at_williams", true).Error)

	// Test 7: error on bad professor
	params = SurveyCreateParams{CourseID: &c1.ID, Comment: generateSurveyTestComment(),
		ProfessorID: &s2.ID}
	createSurveyExpectError(assert, router, params, lib.ErrorSurveyProfessorNotFound)

	// Test 8: error on bad area of study abbrev
	params = SurveyCreateParams{ProfessorID: &p1.ID, Comment: generateSurveyTestComment(),
		CourseNumber:            lib.StrToPtr("201"),
		AreaOfStudyAbbreviation: lib.StrToPtr("PSCI"),
	}
	createSurveyExpectError(assert, router, params, lib.ErrorSurveyAreaOfStudyNotFound)

	// Test 9: create survey via courseID
	params = SurveyCreateParams{ProfessorID: &p1.ID, Comment: generateSurveyTestComment(),
		CourseID:         &c1.ID,
		WouldTakeAnother: lib.BoolToPtr(false),
		CourseWorkload:   lib.IntToPtr(4),
	}
	resSurvey := createSurveyExpectSuccess(assert, router, params)

	// Assert these values:
	assert.Equal(params.Comment, resSurvey.Comment)
	assert.Equal(*params.ProfessorID, resSurvey.ProfessorID)
	assert.Equal(*params.CourseID, resSurvey.CourseID)
	assert.Equal(s1.ID, resSurvey.UserID)
	assert.False(*resSurvey.WouldTakeAnother)
	assert.Equal(*params.CourseWorkload, *resSurvey.CourseWorkload)
	assert.Nil(resSurvey.WouldRecommendCourse)
	assert.Nil(resSurvey.CourseStimulating)
	assert.Zero(resSurvey.TotalAgree)
	assert.Zero(resSurvey.TotalDisagree)
	assert.False(resSurvey.Flagged)
	// Assert that we preload class, area, and professor
	assert.Equal(c1.Number, resSurvey.Course.Number)
	assert.Equal(c1.AreaOfStudy.Abbreviation, resSurvey.Course.AreaOfStudy.Abbreviation)
	assert.Equal(p1.UnixID, resSurvey.Professor.UnixID)

	// Test 10: error on unique survey requirement
	params = SurveyCreateParams{ProfessorID: &p1.ID, Comment: generateSurveyTestComment(),
		CourseID: &c1.ID,
	}
	createSurveyExpectError(assert, router, params, lib.ErrorSurveyAlreadyExists)

	// Test 11: create survey via existing course number
	params = SurveyCreateParams{ProfessorID: &p1.ID, Comment: generateSurveyTestComment(),
		CourseNumber:            &c2.Number,
		AreaOfStudyAbbreviation: &c2.AreaOfStudy.Abbreviation,
		WouldRecommendCourse:    lib.BoolToPtr(true),
		Approachability:         lib.IntToPtr(0),
	}
	resSurvey = createSurveyExpectSuccess(assert, router, params)

	// Assert these values:
	assert.Equal(params.Comment, resSurvey.Comment)
	// Make sure course and area of study are correct
	assert.Equal(c2.ID, resSurvey.CourseID)
	assert.Equal(c2.Number, resSurvey.Course.Number)
	assert.Equal(c2.AreaOfStudy.ID, resSurvey.Course.AreaOfStudy.ID)
	assert.Equal(c2.AreaOfStudy.Abbreviation, resSurvey.Course.AreaOfStudy.Abbreviation)
	// Check passed values
	assert.True(*resSurvey.WouldRecommendCourse)
	assert.Equal(*params.Approachability, *resSurvey.Approachability)

	// Test 12: create survey via new course number
	params = SurveyCreateParams{ProfessorID: &p1.ID, Comment: generateSurveyTestComment(),
		CourseNumber:            lib.StrToPtr("c3"),
		AreaOfStudyAbbreviation: &a1.Abbreviation,
		WouldRecommendCourse:    lib.BoolToPtr(true),
		Approachability:         lib.IntToPtr(7),
	}

	resSurvey = createSurveyExpectSuccess(assert, router, params)

	// Assert these values:
	assert.Equal(params.Comment, resSurvey.Comment)
	// Make sure course and area of study are correct
	assert.Equal(*params.CourseNumber, resSurvey.Course.Number)
	assert.Equal(a1.ID, resSurvey.Course.AreaOfStudy.ID)
	assert.Equal(a1.Abbreviation, resSurvey.Course.AreaOfStudy.Abbreviation)
	// Check passed values
	assert.True(*resSurvey.WouldRecommendCourse)
	assert.Equal(*params.Approachability, *resSurvey.Approachability)
}

func generateSurveyTestComment() string {
	randBytes := make([]byte, 100)
	for i := 0; i < 100; i++ {
		randBytes[i] = byte(65 + rand.Intn(25)) //A=65 and Z = 65+25
	}
	return string(randBytes)
}

func createSurveyExpectError(assert *testify.Assertions, router *gin.Engine, params SurveyCreateParams, apiErr *lib.APIError) {
	paramsData, err := json.Marshal(&params)
	assert.NoError(err)

	// Get bad survey (expect failure)
	w, err := utils.DoHTTPReq(router, http.MethodPost, "/surveys", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	// Assert correct error
	respErrCode := utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode
	assert.Equal(apiErr.Code, respErrCode)
}

func createSurveyExpectSuccess(assert *testify.Assertions, router *gin.Engine, params SurveyCreateParams) models.FactrakSurvey {
	paramsData, err := json.Marshal(&params)
	assert.NoError(err)

	// Get bad survey (expect failure)
	w, err := utils.DoHTTPReq(router, http.MethodPost, "/surveys", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(http.StatusCreated, w.Code)

	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)

	var resp models.FactrakSurvey
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	return resp
}
