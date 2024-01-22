package booktrak_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	sql_types "github.com/WilliamsStudentsOnline/wso-go/lib/sql_types"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/booktrak"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestController_CreateBookListing(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeBulletin, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()

	u1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "User 1",
		UnixID: "u1",
	}
	assert.NoError(db.Create(&u1).Error)

	utils.AddUserContexts(router, u1.ID)
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	/* Create listing with book that hasn't been created (expect failure) */
	params := CreateBookListingParams{
		BookID:      1,
		Condition:   models.ConditionPoor,
		Description: lib.StrToPtr("test"),
		ListingType: models.ListingTypeBuy,
	}
	apiErr := lib.ErrorBookNotFound
	paramsData, err := json.Marshal(&params)
	assert.NoError(err)
	w, err := utils.DoHTTPReq(router, http.MethodPost, "/listings", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Create book */
	book := &models.Book{
		Title: "C Programming Language",
		Isbn:  "9780133086218",
		Authors: sql_types.CSV([]string{
			"Brian W. Kernighan",
			"Dennis Ritchie",
		}),
		Publisher: lib.StrToPtr("Prentice Hall"),
	}
	assert.NoError(db.Create(book).Error)

	/* Create listing with undefined condition (expect failure) */
	params = CreateBookListingParams{
		BookID:      book.ID,
		Condition:   models.ConditionUndefined,
		Description: lib.StrToPtr("test"),
		ListingType: models.ListingTypeBuy,
	}
	apiErr = lib.ErrorBookListingInvalidCondition
	paramsData, err = json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(router, http.MethodPost, "/listings", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Create listing (expect success) */
	params = CreateBookListingParams{
		BookID:      book.ID,
		Condition:   models.ConditionFair,
		Description: lib.StrToPtr("test"),
		ListingType: models.ListingTypeBuy,
	}
	paramsData, err = json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(router, http.MethodPost, "/listings", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(http.StatusCreated, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.BookListing
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Assert good response
	assert.NotZero(resp.ID)
	assert.Equal(params.BookID, resp.BookID)
	assert.Equal(params.Condition, resp.Condition)
	assert.Equal(params.Description, resp.Description)
	assert.Equal(params.ListingType, resp.ListingType)
}

func TestController_UpdateBookListing(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeBulletin, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()

	u1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "User 1",
		UnixID: "u1",
	}
	u2 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "User 2",
		UnixID: "u2",
	}
	assert.NoError(db.Create(&u1).Create(&u2).Error)

	utils.AddUserContexts(router, u1.ID)
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	/* Create book */
	book := &models.Book{
		Title: "C Programming Language",
		Isbn:  "9780133086218",
		Authors: sql_types.CSV([]string{
			"Brian W. Kernighan",
			"Dennis Ritchie",
		}),
		Publisher: lib.StrToPtr("Prentice Hall"),
	}
	assert.NoError(db.Create(book).Error)

	/* Create listing */
	bookListings := []*models.BookListing{
		// 1
		{
			BookID:      book.ID,
			Book:        book,
			UserID:      1,
			Condition:   models.ConditionLikeNew,
			Description: lib.StrToPtr("This is a new book to buy"),
			ListingType: models.ListingTypeBuy,
			User:        &u1,
		},
		// 2
		{
			BookID:      100,
			UserID:      2,
			Condition:   models.ConditionFair,
			Description: lib.StrToPtr("This is a fair book to buy"),
			ListingType: models.ListingTypeBuy,
			User:        &u2,
		},
		// 3
		{
			BookID:      100,
			UserID:      1,
			Condition:   models.ConditionFair,
			Description: lib.StrToPtr("This is a fair book to buy"),
			ListingType: models.ListingTypeBuy,
			User:        &u1,
		},
	}

	// add book listings to db
	for _, val := range bookListings {
		assert.NoError(db.Create(val).Error)
	}

	/* Update listing with bad owner (expect failure) */
	params := CreateBookListingParams{
		BookID:      book.ID,
		Condition:   models.ConditionFair,
		Description: lib.StrToPtr("test"),
		ListingType: models.ListingTypeBuy,
	}
	apiErr := lib.ErrorMustBeSelf
	paramsData, err := json.Marshal(&params)
	assert.NoError(err)
	w, err := utils.DoHTTPReq(router, http.MethodPut, fmt.Sprintf("/listings/2"), bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Update listing with nonexistent book (expect failure) */
	params = CreateBookListingParams{
		BookID:      100,
		Condition:   models.ConditionFair,
		Description: lib.StrToPtr("test"),
		ListingType: models.ListingTypeBuy,
	}
	apiErr = lib.ErrorBookNotFound
	paramsData, err = json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(router, http.MethodPut, fmt.Sprintf("/listings/3"), bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Update listing with undefined condition (expect failure) */
	params = CreateBookListingParams{
		BookID:      book.ID,
		Condition:   models.ConditionUndefined,
		Description: lib.StrToPtr("test"),
		ListingType: models.ListingTypeBuy,
	}
	apiErr = lib.ErrorBookListingInvalidCondition
	paramsData, err = json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(router, http.MethodPut, fmt.Sprintf("/listings/1"), bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Update listing (expect success) */
	params = CreateBookListingParams{
		BookID:      book.ID,
		Condition:   models.ConditionGood,
		Description: lib.StrToPtr("test"),
		ListingType: models.ListingTypeSell,
	}
	paramsData, err = json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(router, http.MethodPut, fmt.Sprintf("/listings/1"), bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.BookListing
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Assert good response
	assert.Equal(params.BookID, resp.BookID)
	assert.Equal(params.Condition, resp.Condition)
	assert.Equal(params.Description, resp.Description)
	assert.Equal(params.ListingType, resp.ListingType)
}

func TestController_ListBookListings(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeBooktrak, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()

	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	// Create book + course pairing for later course filtering
	book := &models.Book{
		Title: "C Programming Language",
		Isbn:  "9780133086218",
		Authors: sql_types.CSV([]string{
			"Brian W. Kernighan",
			"Dennis Ritchie",
		}),
		Publisher: lib.StrToPtr("Prentice Hall"),
	}
	book2 := &models.Book{
		Title: "test",
		Isbn:  "978013110370",
		Authors: sql_types.CSV([]string{
			"Brian W. Kernighan",
			"Dennis Ritchie",
		}),
		Publisher: lib.StrToPtr("Prentice Hall"),
	}
	course := &models.Course{
		Number:        "136",
		AreaOfStudyID: lib.UIntToPtr(1),
	}
	assert.NoError(db.Create(book).Create(book2).Create(course).Error)
	params := UpdateBookCoursesParams{
		CourseIDs: []uint{course.ID},
	}
	paramsData, err := json.Marshal(params)
	assert.NoError(err)
	resp, err := utils.DoHTTPReq(router, http.MethodPatch, fmt.Sprintf("/books/%d", book.ID), bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(http.StatusOK, resp.Code)

	// create book listings
	bookListings := []*models.BookListing{
		// 1
		{
			BookID:      book2.ID,
			Book:        book2,
			UserID:      1,
			Condition:   models.ConditionLikeNew,
			Description: lib.StrToPtr("This is a new book to buy"),
			ListingType: models.ListingTypeBuy,
		},
		// 2
		{
			BookID:      book.ID,
			Book:        book,
			UserID:      2,
			Condition:   models.ConditionFair,
			Description: lib.StrToPtr("This is a fair book to buy"),
			ListingType: models.ListingTypeBuy,
		},
		// 3
		{
			BookID:      book2.ID,
			Book:        book2,
			UserID:      3,
			Condition:   models.ConditionNew,
			Description: lib.StrToPtr("This is a new book to sell"),
			ListingType: models.ListingTypeSell,
		},
		// 4
		{
			BookID:      book.ID,
			Book:        book,
			UserID:      4,
			Condition:   models.ConditionPoor,
			Description: lib.StrToPtr("This is a poor book to sell"),
			ListingType: models.ListingTypeSell,
		},
		// 5
		{
			BookID:      book2.ID,
			Book:        book2,
			UserID:      3,
			Condition:   models.ConditionNew,
			Description: lib.StrToPtr("This is a new book to sell"),
			ListingType: models.ListingTypeSell,
		},
	}

	// add book listings to db
	for _, val := range bookListings {
		assert.NoError(db.Create(val).Error)
	}

	testCases := []struct {
		name     string
		query    string
		expected []uint
	}{
		{
			"default",
			"",
			[]uint{5, 4, 3, 2, 1},
		},
		{
			"filter by book",
			fmt.Sprintf("bookID=%d", book.ID),
			[]uint{4, 2},
		},
		{
			"filter by course",
			fmt.Sprintf("courseID=%d", course.ID),
			[]uint{4, 2},
		},
		{
			"filter by user ID",
			"userID=3",
			[]uint{5, 3},
		},
		{
			"filter by min condition",
			"minCondition=LIKE_NEW",
			[]uint{5, 3, 1},
		},
		{
			"filter by max condition",
			"maxCondition=FAIR",
			[]uint{4, 2},
		},
		{
			"limit and offset",
			"limit=1&offset=1",
			[]uint{4},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := testify.New(t)
			// Get test user
			resp, err := utils.DoHTTPReq(router, http.MethodGet, "/listings?"+tc.query, nil)
			a.NoError(err)

			// Status is okay
			a.Equal(http.StatusOK, resp.Code)

			// Decode response
			respData := utils.GetHTTPDataResp(a, resp.Body.Bytes())
			a.Nil(respData.Error)
			var respBookListings []*models.BookListing
			a.NoError(json.Unmarshal(respData.Data, &respBookListings))

			// Check if correct result
			a.Len(respBookListings, len(tc.expected))
			for i, bookListing := range respBookListings {
				a.Equal(tc.expected[i], bookListing.ID)
			}
		})
	}
}

func TestController_GetBookListing(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeBooktrak, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()

	u1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "User 1",
		UnixID: "u1",
	}
	assert.NoError(db.Create(&u1).Error)

	utils.AddUserContexts(router, u1.ID)
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	// Create book
	book := &models.Book{
		Title: "C Programming Language",
		Isbn:  "9780133086218",
		Authors: sql_types.CSV([]string{
			"Brian W. Kernighan",
			"Dennis Ritchie",
		}),
		Publisher: lib.StrToPtr("Prentice Hall"),
	}
	assert.NoError(db.Create(book).Error)

	// add book listing to db
	bookListing := &models.BookListing{
		BookID:      book.ID,
		Book:        book,
		UserID:      1,
		Condition:   models.ConditionLikeNew,
		Description: lib.StrToPtr("This is a new book to buy"),
		ListingType: models.ListingTypeBuy,
	}
	assert.NoError(db.Create(bookListing).Error)

	// get nonexistent listing (expect failure)
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/listings/10"), nil)
	assert.NoError(err)
	assert.Equal(404, w.Code) // not found

	// get listing (expect success)
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/listings/1"), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.BookListing
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct bulletin
	assert.Equal(bookListing.ID, resp.ID)
	assert.Equal(bookListing.BookID, resp.BookID)
	assert.Equal(bookListing.UserID, resp.UserID)
	assert.Equal(bookListing.Condition, resp.Condition)
	assert.Equal(bookListing.Description, resp.Description)
	assert.Equal(bookListing.ListingType, resp.ListingType)
}

func TestController_DeleteBookListing(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeBooktrak, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()

	u1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "User 1",
		UnixID: "u1",
	}
	u2 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "User 2",
		UnixID: "u2",
	}
	assert.NoError(db.Create(&u1).Create(&u2).Error)

	utils.AddUserContexts(router, u1.ID)
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	// Create book
	book := &models.Book{
		Title: "C Programming Language",
		Isbn:  "9780133086218",
		Authors: sql_types.CSV([]string{
			"Brian W. Kernighan",
			"Dennis Ritchie",
		}),
		Publisher: lib.StrToPtr("Prentice Hall"),
	}
	assert.NoError(db.Create(book).Error)

	// add book listings to db
	bookListing1 := &models.BookListing{
		BookID:      book.ID,
		Book:        book,
		UserID:      1,
		User:        &u1,
		Condition:   models.ConditionLikeNew,
		Description: lib.StrToPtr("This is a new book to buy"),
		ListingType: models.ListingTypeBuy,
	}
	bookListing2 := &models.BookListing{
		BookID:      book.ID,
		Book:        book,
		UserID:      2,
		User:        &u2,
		Condition:   models.ConditionLikeNew,
		Description: lib.StrToPtr("This is a new book to buy"),
		ListingType: models.ListingTypeBuy,
	}
	assert.NoError(db.Create(bookListing1).Create(bookListing2).Error)

	// delete nonexistent listing (expect failure)
	w, err := utils.DoHTTPReq(router, http.MethodDelete, fmt.Sprintf("/listings/10"), nil)
	assert.NoError(err)
	assert.Equal(404, w.Code) // not found

	// delete listing with different owner (expect failure)
	apiErr := lib.ErrorMustBeSelf
	w, err = utils.DoHTTPReq(router, http.MethodDelete, fmt.Sprintf("/listings/2"), nil)
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	// delete listing (expect success)
	w, err = utils.DoHTTPReq(router, http.MethodDelete, fmt.Sprintf("/listings/1"), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.BookListing
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Assert good response
	assert.Equal(bookListing1.ID, resp.ID)

	// Assert not in db
	var count int
	assert.NoError(db.Model(&models.BookListing{}).Where("id = ?", resp.ID).Count(&count).Error)
	assert.Equal(0, count)
}
