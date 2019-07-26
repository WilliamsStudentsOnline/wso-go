package factrak_test

import (
	"net/http/httptest"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
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
