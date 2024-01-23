package models_test

import (
	"testing"

	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	. "github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

// This tests GetAllKeywords
func TestDiningKeywordModel_GetAllKeywords(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	m := NewDiningKeywordModel(db, zaptest.NewLogger(t).Sugar())

	keywords := []DiningKeyword{
		{
			Keyword: "chicken",
		},
		{
			Keyword: "potsticker",
		},
	}

	for i := range keywords {
		assert.NoError(db.Create(&keywords[i]).Error)
	}

	var res []*DiningKeyword
	assert.NoError(m.GetAllKeywords(&res))

	for i := range keywords {
		assert.Equal(keywords[i].ID, res[i].ID)
		assert.Equal(keywords[i].Keyword, res[i].Keyword)
	}
}

func TestDiningKeywordModel_CreateKeyword(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	m := NewDiningKeywordModel(db, zaptest.NewLogger(t).Sugar())

	//create two users

	// Create a new keyword
	new_keyword := &DiningKeyword{
		Keyword: "yellow chicken",
	}

	// Test CreateKeyword
	assert.NoError(m.CreateKeyword(new_keyword))

	// Verify the keyword is created
	var createdKeyword DiningKeyword
	db.Where("keyword = ?", "yellow chicken").First(&createdKeyword)
	assert.NotZero(createdKeyword.ID)
	assert.Equal("yellow chicken", createdKeyword.Keyword)
}

func TestDiningKeywordModel_KeywordAssociation(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	m := NewDiningKeywordModel(db, zaptest.NewLogger(t).Sugar())
	new_keyword := &DiningKeyword{
		Keyword: "yellow chicken",
	}
	new_user := &User{
		Type:   UserTypeStudent,
		Name:   "Donald J Trump",
		UnixID: "djt12",
	}

	assert.NoError(
		db.Create(&new_user).Error)

	assert.NoError(m.CreateKeyword(new_keyword))

	assert.NoError(m.NewKeywordAssociation(new_keyword, new_user))

	var associated_users []User
	assert.NoError(m.DB.Model(new_keyword).Association("Users").Find(&associated_users).Error)
	assert.Equal(associated_users[0].Name, "Donald J Trump")

	assert.NoError(m.DeleteKeywordAssociation(new_keyword, new_user))
	return

}
