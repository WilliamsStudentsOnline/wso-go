package factrak_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"

	. "github.com/WilliamsStudentsOnline/wso-go/services/factrak"
)

func TestController_ListUserSurveys(t *testing.T) {
	// Setup
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	// Insert test user into db
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
	s2 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "Student",
		UnixID: "s2",
	}
	p2 := models.User{
		Type:   models.UserTypeProfessor,
		Name:   "Professor 2",
		UnixID: "p2",
	}
	assert.NoError(db.Create(&p1).Create(&s1).Create(&s2).Create(&p2).Error)

	fs1 := models.FactrakSurvey{
		User:      &s1,
		Professor: &p1,
		Comment:   "Survey 1",
	}
	fs2 := models.FactrakSurvey{
		User:      &s2,
		Professor: &p1,
		Comment:   "Survey 2",
	}
	fs3 := models.FactrakSurvey{
		User:      &s1,
		Professor: &p2,
		Comment:   "Survey 3",
	}

	assert.NoError(db.Create(&fs1).Create(&fs2).Create(&fs3).Error)

	router := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf)
	utils.AddUserContexts(router, s1.ID)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	/* Get test student 1's surveys when authorized (expect success) */
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/users/%d/surveys", s1.ID), nil)
	assert.NoError(err)

	// Check if response has user's surveys, in reverse chronological order
	resp := GetSurveysFromResp(assert, w)
	assert.NoError(EqualSurveyIDs([]models.FactrakSurvey{fs3, fs1}, resp))

	// Assert that userID is returned (as we are the owner)
	assert.NotZero(resp[0].UserID)

	/* Get surveys by prof 1 when authorized (expect empty success) */
	r2 := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf)
	utils.AddUserContexts(r2, p1.ID)
	SetupRouter(r2, db, cfg, zaptest.NewLogger(t).Sugar())

	w, err = utils.DoHTTPReq(r2, http.MethodGet, fmt.Sprintf("/users/%d/surveys", p1.ID), nil)
	assert.NoError(err)

	// Check if response is valid but contains no surveys
	resp = GetSurveysFromResp(assert, w)
	assert.Len(resp, 0)

	/* Get surveys for random fake user (expect failure) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/users/%d/surveys", 42), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusNotFound, w.Code)

	/* Get test student 2 when not authorized (expect forbidden) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/users/%d/surveys", s2.ID), nil)
	assert.NoError(err)

	// Status is forbidden
	assert.Equal(http.StatusForbidden, w.Code)
}
