package models_test

import (
	"testing"

	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	. "github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestDiscussionModel_GetDiscussionWithDeletedUser(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewDiscussionModel(db, zaptest.NewLogger(t).Sugar())

	u1 := User{
		Type:   UserTypeStudent,
		Name:   "User 1",
		UnixID: "u1",
	}
	u2 := User{
		Type:   UserTypeStudent,
		Name:   "User 2",
		UnixID: "u2",
	}
	assert.NoError(db.Create(&u1).Create(&u2).Error)

	d1 := Discussion{
		User:  &u1,
		Title: "Discussion 1",
	}
	d2 := Discussion{
		User:  &u2,
		Title: "Discussion 2",
	}
	d3 := Discussion{
		UserID: 43,
		Title:  "Discussion 2",
	}

	assert.NoError(db.Create(&d1).Create(&d2).Create(&d3).Error)

	// Delete user 2
	assert.NoError(db.Delete(&u2).Error)

	/* Test getting the normal discussion */
	res := Discussion{}
	assert.NoError(m.GetDiscussionByID(d1.ID, &res, &GetDiscussionByIDOptions{Preload: []string{"user"}}))
	assert.NotNil(res.User)

	/* Test getting the soft delete discussion */
	res = Discussion{}
	assert.NoError(m.GetDiscussionByID(d2.ID, &res, &GetDiscussionByIDOptions{Preload: []string{"user"}}))
	assert.Nil(res.User)

	/* Test getting the hard delete discussion */
	res = Discussion{}
	assert.NoError(m.GetDiscussionByID(d3.ID, &res, &GetDiscussionByIDOptions{Preload: []string{"user"}}))
	assert.Nil(res.User)
}
