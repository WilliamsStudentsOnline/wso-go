package models_test

import (
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	. "github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestBookListingModel_GetAllBookListings(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewBookListingModel(db, zaptest.NewLogger(t).Sugar())

	bookListings := []BookListing{
		{
			BookID:      1,
			UserID:      1,
			Condition:   ConditionNew,
			Description: lib.StrToPtr("This is a great book"),
			ListingType: ListingTypeBuy,
		},
		{
			BookID:      2,
			UserID:      2,
			Condition:   ConditionGood,
			Description: lib.StrToPtr("This is a good book"),
			ListingType: ListingTypeSell,
		},
	}

	for i := range bookListings {
		assert.NoError(db.Create(&bookListings[i]).Error)
	}

	var res []*BookListing
	assert.NoError(m.GetAllBookListings(&res, nil))
	for i := range bookListings {
		assert.Equal(bookListings[i].ID, res[i].ID)
		assert.Equal(bookListings[i].BookID, res[i].BookID)
		assert.Equal(bookListings[i].UserID, res[i].UserID)
		assert.Equal(bookListings[i].Condition, res[i].Condition)
		assert.Equal(bookListings[i].Description, res[i].Description)
		assert.Equal(bookListings[i].ListingType, res[i].ListingType)
	}
}

func TestBookListingModel_GetBookListingByID(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewBookListingModel(db, zaptest.NewLogger(t).Sugar())

	bookListing := BookListing{
		BaseSchema: BaseSchema{
			ID: 1,
		},
		BookID:      1,
		UserID:      1,
		Condition:   ConditionNew,
		Description: lib.StrToPtr("This is a great book"),
		ListingType: ListingTypeBuy,
	}

	assert.NoError(db.Create(&bookListing).Error)

	var res BookListing
	assert.NoError(m.GetBookListingByID(bookListing.ID, &res))
	assert.Equal(bookListing.ID, res.ID)
	assert.Equal(bookListing.BookID, res.BookID)
	assert.Equal(bookListing.UserID, res.UserID)
	assert.Equal(bookListing.Condition, res.Condition)
	assert.Equal(bookListing.Description, res.Description)
	assert.Equal(bookListing.ListingType, res.ListingType)
}

func TestBookListingModel_CreateBookListing(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewBookListingModel(db, zaptest.NewLogger(t).Sugar())

	bookListing := BookListing{
		BookID:      1,
		UserID:      1,
		Condition:   ConditionNew,
		Description: lib.StrToPtr("This is a great book"),
		ListingType: ListingTypeBuy,
	}

	assert.NoError(m.CreateBookListing(&bookListing))

	var res BookListing
	assert.NoError(db.First(&res).Error)
	assert.Equal(bookListing.ID, res.ID)
	assert.Equal(bookListing.BookID, res.BookID)
	assert.Equal(bookListing.UserID, res.UserID)
	assert.Equal(bookListing.Condition, res.Condition)
	assert.Equal(bookListing.Description, res.Description)
	assert.Equal(bookListing.ListingType, res.ListingType)
}

func TestBookListingModel_UpdateBookListing(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewBookListingModel(db, zaptest.NewLogger(t).Sugar())

	bookListing := BookListing{
		BookID:      1,
		UserID:      1,
		Condition:   ConditionNew,
		Description: lib.StrToPtr("This is a great book"),
		ListingType: ListingTypeBuy,
	}

	assert.NoError(db.Create(&bookListing).Error)

	bookListing.Condition = ConditionGood
	bookListing.Description = lib.StrToPtr("This is a good book")
	bookListing.ListingType = ListingTypeSell

	assert.NoError(m.UpdateBookListing(&bookListing))

	var res BookListing
	assert.NoError(db.First(&res).Error)
	assert.Equal(bookListing.ID, res.ID)
	assert.Equal(bookListing.BookID, res.BookID)
	assert.Equal(bookListing.UserID, res.UserID)
	assert.Equal(bookListing.Condition, res.Condition)
	assert.Equal(bookListing.Description, res.Description)
	assert.Equal(bookListing.ListingType, res.ListingType)
}

func TestBookListingModel_DeleteBookListing(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewBookListingModel(db, zaptest.NewLogger(t).Sugar())

	bookListing := BookListing{
		BookID:      1,
		UserID:      1,
		Condition:   ConditionNew,
		Description: lib.StrToPtr("This is a great book"),
		ListingType: ListingTypeBuy,
	}

	assert.NoError(db.Create(&bookListing).Error)

	assert.NoError(m.DeleteBookListing(&bookListing))

	var count int
	assert.NoError(db.Model(&BookListing{}).Where("id = ?", bookListing.ID).Count(&count).Error)
	assert.Equal(0, count)
}
