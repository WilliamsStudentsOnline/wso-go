package dormtrak_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/dormtrak"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

// Quick setup for factrak tests, configuring the router and database
func SetupDormtrakTest(t *testing.T) (*testify.Assertions, *gorm.DB, *gin.Engine) {
	// Create test environment using factrak scopes
	env := utils.SetupTest(t, auth.ScopeDormtrakFull, auth.ScopeWriteSelf)
	// Set up the dormtrak router
	SetupRouter(env.Router, env.DB, env.Cfg, zaptest.NewLogger(t).Sugar())

	return env.Assert, env.DB, env.Router
}

func TestRemoveUserIDFromReviews(t *testing.T) {
	assert := testify.New(t)

	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Set the id and scopes
	var selfID uint = 3
	ctx.Set("id", selfID)
	ctx.Set("scopes", []string{auth.ScopeDormtrakFull, auth.ScopeWriteSelf})

	// Create some demo users
	u1 := models.NewUserWithID(1)
	u1.UnixID = "u1"
	u2 := models.NewUserWithID(2)
	u1.UnixID = "u2"
	self := models.NewUserWithID(selfID)
	u1.UnixID = "self"

	// Test 1: remove everything that isn't self
	reviews := []*models.DormtrakReview{
		{UserID: 1, User: &u1, Comment: lib.StrToPtr("s1")},
		{UserID: 2, User: &u2, Comment: lib.StrToPtr("s2")},
		{UserID: selfID, User: &self, Comment: lib.StrToPtr("s3")},
		{UserID: 1, User: &u1, Comment: lib.StrToPtr("s4")},
		{UserID: 1, User: &u1, Comment: lib.StrToPtr("s5")},
		{UserID: selfID, User: &self, Comment: lib.StrToPtr("s6")},
	}
	expected := []*models.DormtrakReview{
		{UserID: 0, User: nil, Comment: lib.StrToPtr("s1")},
		{UserID: 0, User: nil, Comment: lib.StrToPtr("s2")},
		{UserID: selfID, User: &self, Comment: lib.StrToPtr("s3")},
		{UserID: 0, User: nil, Comment: lib.StrToPtr("s4")},
		{UserID: 0, User: nil, Comment: lib.StrToPtr("s5")},
		{UserID: selfID, User: &self, Comment: lib.StrToPtr("s6")},
	}

	RemoveUserIDFromReviews(ctx, reviews)

	assert.EqualValues(expected, reviews)

	// Test 2: bypass if admin
	reviews = []*models.DormtrakReview{
		{UserID: 1, User: &u1, Comment: lib.StrToPtr("s1")},
		{UserID: 2, User: &u2, Comment: lib.StrToPtr("s2")},
		{UserID: selfID, User: &self, Comment: lib.StrToPtr("s3")},
		{UserID: 1, User: &u1, Comment: lib.StrToPtr("s4")},
		{UserID: 1, User: &u1, Comment: lib.StrToPtr("s5")},
		{UserID: selfID, User: &self, Comment: lib.StrToPtr("s6")},
	}
	expected = []*models.DormtrakReview{
		{UserID: 1, User: &u1, Comment: lib.StrToPtr("s1")},
		{UserID: 2, User: &u2, Comment: lib.StrToPtr("s2")},
		{UserID: selfID, User: &self, Comment: lib.StrToPtr("s3")},
		{UserID: 1, User: &u1, Comment: lib.StrToPtr("s4")},
		{UserID: 1, User: &u1, Comment: lib.StrToPtr("s5")},
		{UserID: selfID, User: &self, Comment: lib.StrToPtr("s6")},
	}

	ctx.Set("scopes", []string{auth.ScopeAdminAll})
	RemoveUserIDFromReviews(ctx, reviews)
	assert.EqualValues(expected, reviews)

	// Test 2: bypass if dormtrak admin
	reviews = []*models.DormtrakReview{
		{UserID: 1, User: &u1, Comment: lib.StrToPtr("s1")},
		{UserID: 2, User: &u2, Comment: lib.StrToPtr("s2")},
		{UserID: selfID, User: &self, Comment: lib.StrToPtr("s3")},
		{UserID: 1, User: &u1, Comment: lib.StrToPtr("s4")},
		{UserID: 1, User: &u1, Comment: lib.StrToPtr("s5")},
		{UserID: selfID, User: &self, Comment: lib.StrToPtr("s6")},
	}

	ctx.Set("scopes", []string{auth.ScopeDormtrakAdmin})
	RemoveUserIDFromReviews(ctx, reviews)
	assert.EqualValues(expected, reviews)
}

func TestIsScopeLimited(t *testing.T) {
	assert := testify.New(t)

	// Returns true when scope is limited
	c := &gin.Context{}
	c.Set("scopes", []string{auth.ScopeDormtrakLimited})
	assert.True(IsScopeLimited(c))

	// Returns true when scope is empty
	c = &gin.Context{}
	c.Set("scopes", []string{})
	assert.True(IsScopeLimited(c))

	// Returns true when scope is limited and other
	c = &gin.Context{}
	c.Set("scopes", []string{auth.ScopeDormtrakLimited, auth.ScopeUsers})
	assert.True(IsScopeLimited(c))

	// Returns false when scope is full
	c = &gin.Context{}
	c.Set("scopes", []string{auth.ScopeDormtrakFull})
	assert.False(IsScopeLimited(c))

	// Returns false when scope is factrak admin
	c = &gin.Context{}
	c.Set("scopes", []string{auth.ScopeDormtrakAdmin})
	assert.False(IsScopeLimited(c))

	// Returns false when scope is admin
	c = &gin.Context{}
	c.Set("scopes", []string{auth.ScopeAdminAll})
	assert.False(IsScopeLimited(c))

	// Returns false when scope is full and limited
	c = &gin.Context{}
	c.Set("scopes", []string{auth.ScopeDormtrakLimited, auth.ScopeDormtrakFull})
	assert.False(IsScopeLimited(c))
}

func TestIsScopeAdmin(t *testing.T) {
	assert := testify.New(t)

	// Returns true when scope is factrak admin
	c := &gin.Context{}
	c.Set("scopes", []string{auth.ScopeDormtrakAdmin})
	assert.True(IsScopeAdmin(c))

	// Returns true when scope is global admin
	c = &gin.Context{}
	c.Set("scopes", []string{auth.ScopeAdminAll})
	assert.True(IsScopeAdmin(c))

	// Returns true when scope is factrak admin and other
	c = &gin.Context{}
	c.Set("scopes", []string{auth.ScopeUsers, auth.ScopeDormtrakAdmin})
	assert.True(IsScopeAdmin(c))

	// Returns false when scope is full
	c = &gin.Context{}
	c.Set("scopes", []string{auth.ScopeDormtrakFull})
	assert.False(IsScopeAdmin(c))

	// Returns false when scope is limited
	c = &gin.Context{}
	c.Set("scopes", []string{auth.ScopeDormtrakLimited})
	assert.False(IsScopeAdmin(c))

	// Returns false when scope is empty
	c = &gin.Context{}
	c.Set("scopes", []string{})
	assert.False(IsScopeAdmin(c))

	// Returns false when scope is full and limited
	c = &gin.Context{}
	c.Set("scopes", []string{auth.ScopeDormtrakLimited, auth.ScopeDormtrakFull})
	assert.False(IsScopeAdmin(c))
}

func TestLimitedScopeAccess(t *testing.T) {
	assert := testify.New(t)

	db := utils.SetupServiceTest(assert)

	n1 := models.Neighborhood{
		Name: "Currier",
	}
	do1 := models.Dorm{
		Neighborhood: &n1,
		Name:         "Hubble",
		DormRooms: []*models.DormRoom{
			{
				Number: "102",
			},
		},
	}
	d1 := models.DormRoom{
		Dorm:   &do1,
		DormID: 01,
		Number: "2.CONST",
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
	dr1 := models.DormtrakReview{
		User:     &s1,
		DormRoom: &d1,
		Comment:  lib.StrToPtr("Review 1"),
	}
	dr2 := models.DormtrakReview{
		User:     &s2,
		DormRoom: &d1,
		Comment:  lib.StrToPtr("Review 2"),
	}
	assert.NoError(db.Create(&n1).Create(&s1).Create(&s2).Create(&do1).Create(&d1).Create(&dr1).Create(&dr2).Error)

	r := utils.SetupRouter(auth.ScopeDormtrakLimited, auth.ScopeUsers, auth.ScopeWriteSelf)
	utils.AddUserContexts(r, s1.ID)
	cfg := utils.SetupConfig()
	SetupRouter(r, db, cfg, zaptest.NewLogger(t).Sugar())

	// // Can list professors
	// w, err := utils.DoHTTPReq(r, http.MethodGet, fmt.Sprintf("/professors"), nil)
	// assert.NoError(err)
	// assert.Equal(http.StatusOK, w.Code)

	// // Can get professor, but not with reviews
	// w, err = utils.DoHTTPReq(r, http.MethodGet, fmt.Sprintf("/professors/%d", p1.ID), nil)
	// assert.NoError(err)
	// profResp := GetUserFromResp(assert, w)
	// assert.Nil(profResp.DormtrakReviews)

	// Cannot get professor's reviews
	w, err := utils.DoHTTPReq(r, http.MethodGet, fmt.Sprintf("/dormtrak_reviews/%d/dorm_room_id", d1.ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusForbidden, w.Code)

	// // Cannot get professor's ratings
	// w, err = utils.DoHTTPReq(r, http.MethodGet, fmt.Sprintf("/professors/%d/ratings", p1.ID), nil)
	// assert.NoError(err)
	// assert.Equal(http.StatusForbidden, w.Code)

	// // Can get user's reviews
	// w, err = utils.DoHTTPReq(r, http.MethodGet, fmt.Sprintf("/users/%d/reviews", s1.ID), nil)
	// assert.NoError(err)
	// assert.Equal(http.StatusOK, w.Code)

	// // Cannot get course's reviews
	// w, err = utils.DoHTTPReq(r, http.MethodGet, fmt.Sprintf("/courses/%d/reviews", c1.ID), nil)
	// assert.NoError(err)
	// assert.Equal(http.StatusForbidden, w.Code)

	// // Cannot get course's ratings
	// w, err = utils.DoHTTPReq(r, http.MethodGet, fmt.Sprintf("/courses/%d/ratings", c1.ID), nil)
	// assert.NoError(err)
	// assert.Equal(http.StatusForbidden, w.Code)

	// // Cannot list reviews
	// w, err = utils.DoHTTPReq(r, http.MethodGet, fmt.Sprintf("/reviews"), nil)
	// assert.NoError(err)
	// assert.Equal(http.StatusForbidden, w.Code)

	// // Can get review, but only owned review (works as s1)
	w, err = utils.DoHTTPReq(r, http.MethodGet, fmt.Sprintf("/reviews/%d", dr1.ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// // Can get review, but only owned review (fails as s1)
	// w, err = utils.DoHTTPReq(r, http.MethodGet, fmt.Sprintf("/reviews/%d", fs2.ID), nil)
	// assert.NoError(err)
	// assert.Equal(http.StatusForbidden, w.Code)

	// // Cannot flag review
	// w, err = utils.DoHTTPReq(r, http.MethodPost, fmt.Sprintf("/reviews/%d/flag", fs1.ID), nil)
	// assert.NoError(err)
	// assert.Equal(http.StatusForbidden, w.Code)

	// // Cannot ___ agreement
	// methods := []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete}
	// for _, method := range methods {
	// 	w, err = utils.DoHTTPReq(r, method, fmt.Sprintf("/reviews/%d/agreement", fs1.ID), nil)
	// 	assert.NoError(err)
	// 	assert.Equal(http.StatusForbidden, w.Code)
	// }
}
