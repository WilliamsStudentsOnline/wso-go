package factrak_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/factrak"
	"github.com/gin-gonic/gin"
	testify "github.com/stretchr/testify/assert"
)

func TestRemoveUserIDFromSurveys(t *testing.T) {
	assert := testify.New(t)

	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Set the id and scopes
	var selfID uint = 3
	ctx.Set("id", selfID)
	ctx.Set("scopes", []string{auth.ScopeFactrakFull, auth.ScopeWriteSelf})

	// Create some demo users
	u1 := models.NewUserWithID(1)
	u1.UnixID = "u1"
	u2 := models.NewUserWithID(2)
	u1.UnixID = "u2"
	self := models.NewUserWithID(selfID)
	u1.UnixID = "self"

	// Test 1: remove everything that isn't self
	surveys := []*models.FactrakSurvey{
		{UserID: 1, User: &u1, Comment: "s1"},
		{UserID: 2, User: &u2, Comment: "s2"},
		{UserID: selfID, User: &self, Comment: "s3"},
		{UserID: 1, User: &u1, Comment: "s4"},
		{UserID: 1, User: &u1, Comment: "s5"},
		{UserID: selfID, User: &self, Comment: "s6"},
	}
	expected := []*models.FactrakSurvey{
		{UserID: 0, User: nil, Comment: "s1"},
		{UserID: 0, User: nil, Comment: "s2"},
		{UserID: selfID, User: &self, Comment: "s3"},
		{UserID: 0, User: nil, Comment: "s4"},
		{UserID: 0, User: nil, Comment: "s5"},
		{UserID: selfID, User: &self, Comment: "s6"},
	}

	RemoveUserIDFromSurveys(ctx, surveys)

	assert.EqualValues(expected, surveys)

	// Test 2: bypass if admin
	surveys = []*models.FactrakSurvey{
		{UserID: 1, User: &u1, Comment: "s1"},
		{UserID: 2, User: &u2, Comment: "s2"},
		{UserID: selfID, User: &self, Comment: "s3"},
		{UserID: 1, User: &u1, Comment: "s4"},
		{UserID: 1, User: &u1, Comment: "s5"},
		{UserID: selfID, User: &self, Comment: "s6"},
	}
	expected = []*models.FactrakSurvey{
		{UserID: 1, User: &u1, Comment: "s1"},
		{UserID: 2, User: &u2, Comment: "s2"},
		{UserID: selfID, User: &self, Comment: "s3"},
		{UserID: 1, User: &u1, Comment: "s4"},
		{UserID: 1, User: &u1, Comment: "s5"},
		{UserID: selfID, User: &self, Comment: "s6"},
	}

	ctx.Set("scopes", []string{auth.ScopeAdminAll})
	RemoveUserIDFromSurveys(ctx, surveys)
	assert.EqualValues(expected, surveys)

	// Test 2: bypass if factrak admin
	surveys = []*models.FactrakSurvey{
		{UserID: 1, User: &u1, Comment: "s1"},
		{UserID: 2, User: &u2, Comment: "s2"},
		{UserID: selfID, User: &self, Comment: "s3"},
		{UserID: 1, User: &u1, Comment: "s4"},
		{UserID: 1, User: &u1, Comment: "s5"},
		{UserID: selfID, User: &self, Comment: "s6"},
	}

	ctx.Set("scopes", []string{auth.ScopeFactrakAdmin})
	RemoveUserIDFromSurveys(ctx, surveys)
	assert.EqualValues(expected, surveys)
}

func TestIsScopeLimited(t *testing.T) {
	assert := testify.New(t)

	// Returns true when scope is limited
	c := &gin.Context{}
	c.Set("scopes", []string{auth.ScopeFactrakLimited})
	assert.True(IsScopeLimited(c))

	// Returns true when scope is empty
	c = &gin.Context{}
	c.Set("scopes", []string{})
	assert.True(IsScopeLimited(c))

	// Returns true when scope is limited and other
	c = &gin.Context{}
	c.Set("scopes", []string{auth.ScopeFactrakLimited, auth.ScopeUsers})
	assert.True(IsScopeLimited(c))

	// Returns false when scope is full
	c = &gin.Context{}
	c.Set("scopes", []string{auth.ScopeFactrakFull})
	assert.False(IsScopeLimited(c))

	// Returns false when scope is factrak admin
	c = &gin.Context{}
	c.Set("scopes", []string{auth.ScopeFactrakAdmin})
	assert.False(IsScopeLimited(c))

	// Returns false when scope is admin
	c = &gin.Context{}
	c.Set("scopes", []string{auth.ScopeAdminAll})
	assert.False(IsScopeLimited(c))

	// Returns false when scope is full and limited
	c = &gin.Context{}
	c.Set("scopes", []string{auth.ScopeFactrakLimited, auth.ScopeFactrakFull})
	assert.False(IsScopeLimited(c))
}

func TestIsScopeAdmin(t *testing.T) {
	assert := testify.New(t)

	// Returns true when scope is factrak admin
	c := &gin.Context{}
	c.Set("scopes", []string{auth.ScopeFactrakAdmin})
	assert.True(IsScopeAdmin(c))

	// Returns true when scope is global admin
	c = &gin.Context{}
	c.Set("scopes", []string{auth.ScopeAdminAll})
	assert.True(IsScopeAdmin(c))

	// Returns true when scope is factrak admin and other
	c = &gin.Context{}
	c.Set("scopes", []string{auth.ScopeUsers, auth.ScopeFactrakAdmin})
	assert.True(IsScopeAdmin(c))

	// Returns false when scope is full
	c = &gin.Context{}
	c.Set("scopes", []string{auth.ScopeFactrakFull})
	assert.False(IsScopeAdmin(c))

	// Returns false when scope is limited
	c = &gin.Context{}
	c.Set("scopes", []string{auth.ScopeFactrakLimited})
	assert.False(IsScopeAdmin(c))

	// Returns false when scope is empty
	c = &gin.Context{}
	c.Set("scopes", []string{})
	assert.False(IsScopeAdmin(c))

	// Returns false when scope is full and limited
	c = &gin.Context{}
	c.Set("scopes", []string{auth.ScopeFactrakLimited, auth.ScopeFactrakFull})
	assert.False(IsScopeAdmin(c))
}

func TestLimitedScopeAccess(t *testing.T) {
	assert := testify.New(t)

	db := utils.SetupServiceTest(assert)

	// Populate DB
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
		Name:   "Student 1",
		UnixID: "s1",
	}
	s2 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "Student 2",
		UnixID: "s2",
	}
	fs1 := models.FactrakSurvey{
		User:      &s1,
		Professor: &p1,
		Course:    &c1,
		Comment:   "Survey 1",
	}
	fs2 := models.FactrakSurvey{
		User:      &s2,
		Professor: &p1,
		Course:    &c1,
		Comment:   "Survey 2",
	}
	assert.NoError(db.Create(&c1).Create(&s1).Create(&s2).Create(&p1).Create(&fs1).Create(&fs2).Error)

	r := utils.SetupRouter(auth.ScopeFactrakLimited, auth.ScopeUsers, auth.ScopeWriteSelf)
	utils.AddUserContexts(r, s1.ID)
	SetupRouter(r, db)

	// Can list professors
	w, err := utils.DoHTTPReq(r, http.MethodGet, fmt.Sprintf("/professors"), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Can get professor, but not with surveys
	w, err = utils.DoHTTPReq(r, http.MethodGet, fmt.Sprintf("/professors/%d", p1.ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)
	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var profResp models.User
	assert.NoError(json.Unmarshal(respData.Data, &profResp))
	assert.Nil(profResp.FactrakSurveys)

	// Cannot get professor's surveys
	w, err = utils.DoHTTPReq(r, http.MethodGet, fmt.Sprintf("/professors/%d/surveys", p1.ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusForbidden, w.Code)

	// Cannot get professor's ratings
	w, err = utils.DoHTTPReq(r, http.MethodGet, fmt.Sprintf("/professors/%d/ratings", p1.ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusForbidden, w.Code)

	// Can get user's surveys
	w, err = utils.DoHTTPReq(r, http.MethodGet, fmt.Sprintf("/users/%d/surveys", s1.ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Cannot get course's surveys
	w, err = utils.DoHTTPReq(r, http.MethodGet, fmt.Sprintf("/courses/%d/surveys", c1.ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusForbidden, w.Code)

	// Cannot get course's ratings
	w, err = utils.DoHTTPReq(r, http.MethodGet, fmt.Sprintf("/courses/%d/ratings", c1.ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusForbidden, w.Code)

	// Cannot list surveys
	w, err = utils.DoHTTPReq(r, http.MethodGet, fmt.Sprintf("/surveys"), nil)
	assert.NoError(err)
	assert.Equal(http.StatusForbidden, w.Code)

	// Can get survey, but only owned survey (works as s1)
	w, err = utils.DoHTTPReq(r, http.MethodGet, fmt.Sprintf("/surveys/%d", fs1.ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Can get survey, but only owned survey (fails as s1)
	w, err = utils.DoHTTPReq(r, http.MethodGet, fmt.Sprintf("/surveys/%d", fs2.ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusForbidden, w.Code)

	// Cannot flag survey
	w, err = utils.DoHTTPReq(r, http.MethodPost, fmt.Sprintf("/surveys/%d/flag", fs1.ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusForbidden, w.Code)

	// Cannot ___ agreement
	methods := []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete}
	for _, method := range methods {
		w, err = utils.DoHTTPReq(r, method, fmt.Sprintf("/surveys/%d/agreement", fs1.ID), nil)
		assert.NoError(err)
		assert.Equal(http.StatusForbidden, w.Code)
	}
}
