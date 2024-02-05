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

func TestController_ListBooks(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeBooktrak, auth.ScopeBooktrakWrite, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()

	books := []*models.Book{
		{
			Title: "first",
		},
		{
			Publisher: lib.StrToPtr("second"),
		},
		{
			Isbn: "9780131101630",
		},
	}

	for _, val := range books {
		assert.NoError(db.Create(val).Error)
	}

	testCases := []struct {
		name        string
		query       string
		expectedIds []uint
	}{
		{
			"default",
			"",
			[]uint{3, 2, 1},
		},
		{
			"filter by title",
			"title=first",
			[]uint{1},
		},
		{
			"filter by publisher",
			"publisher=second",
			[]uint{2},
		},
		{
			"filter by isbn",
			"isbn=9780131101630",
			[]uint{3},
		},
		{
			"limit",
			"limit=2",
			[]uint{3, 2},
		},
		{
			"limit and offset",
			"limit=1&offset=1",
			[]uint{2},
		},
	}

	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := testify.New(t)
			resp, err := utils.DoHTTPReq(router, http.MethodGet, "/books?"+tc.query, nil)
			a.NoError(err)

			// Status is okay
			a.Equal(http.StatusOK, resp.Code)

			// Decode response
			respData := utils.GetHTTPDataResp(a, resp.Body.Bytes())
			a.Nil(respData.Error)
			var respBooks []*models.Book
			a.NoError(json.Unmarshal(respData.Data, &respBooks))

			// Check if correct result
			a.Len(respBooks, len(tc.expectedIds))
			for i, book := range respBooks {
				a.Equal(tc.expectedIds[i], book.ID)
			}
		})
	}
}

func TestController_GetBook(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeBooktrak, auth.ScopeBooktrakWrite, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()

	book := &models.Book{
		Title: "The C Programming Language",
		Isbn:  "9780131101630",
		Authors: sql_types.CSV([]string{
			"Brian W. Kernighan",
			"Dennis M. Ritchie",
		}),
		Publisher: lib.StrToPtr("Prentice Hall"),
		InfoLink:  lib.StrToPtr("info link"),
		ImageLink: lib.StrToPtr("image link"),
	}
	assert.NoError(db.Create(&book).Error)
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	resp, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/books/%d", book.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, resp.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, resp.Body.Bytes())
	assert.Nil(respData.Error)
	var respBook models.Book
	assert.NoError(json.Unmarshal(respData.Data, &respBook))

	// Check if correct book
	assert.Equal(book.ID, respBook.ID)
	assert.Equal(book.Title, respBook.Title)
	assert.Equal(book.Isbn, respBook.Isbn)
	assert.Equal(book.Authors, respBook.Authors)
	assert.Equal(book.Publisher, respBook.Publisher)
	assert.Equal(book.InfoLink, respBook.InfoLink)
	assert.Equal(book.ImageLink, respBook.ImageLink)

	// Get book with invalid id
	resp, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/books/%d", 42), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusNotFound, resp.Code)
}

func TestController_CreateBoook(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeBooktrak, auth.ScopeBooktrakWrite, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()

	book := &models.Book{
		Title: "C Programming Language",
		Isbn:  "9780133086218",
		Authors: sql_types.CSV([]string{
			"Brian W. Kernighan",
			"Dennis Ritchie",
		}),
		Publisher: lib.StrToPtr("Prentice Hall"),
	}

	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	// Create book with invalid data (expect failure)
	apiError := lib.ErrorRequestDataValidationFailed
	resp, err := utils.DoHTTPReq(router, http.MethodPost, "/books", bytes.NewBufferString(`{"title": "The C Programming Language"}`))
	assert.NoError(err)
	assert.Equal(apiError.HTTPCode, resp.Code)
	assert.Equal(apiError.Code, utils.GetHTTPDataResp(assert, resp.Body.Bytes()).Error.ErrorCode)

	// Create book with malformed isbn (expect failure)
	apiError = lib.ErrorRequestDataValidationFailed
	resp, err = utils.DoHTTPReq(router, http.MethodPost, "/books", bytes.NewBufferString(`{"isbn": "978013110370"}`))
	assert.NoError(err)
	assert.Equal(apiError.HTTPCode, resp.Code)
	assert.Equal(apiError.Code, utils.GetHTTPDataResp(assert, resp.Body.Bytes()).Error.ErrorCode)

	params := CreateBookParams{
		ISBN: "9780133086218",
	}
	paramsData, err := json.Marshal(params)
	assert.NoError(err)

	// Create book with valid data (expect success)
	resp, err = utils.DoHTTPReq(router, http.MethodPost, "/books", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(http.StatusCreated, resp.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, resp.Body.Bytes())
	assert.Nil(respData.Error)
	var respBook models.Book
	assert.NoError(json.Unmarshal(respData.Data, &respBook))

	// Check if correct book
	assert.Equal(book.Title, respBook.Title)
	assert.Equal(book.Isbn, respBook.Isbn)
	assert.Equal(book.Authors, respBook.Authors)
	assert.Equal(book.Publisher, respBook.Publisher)

	// Assert found in db
	var count int
	assert.NoError(db.Model(&models.Book{}).Where("id = ?", respBook.ID).Count(&count).Error)
	assert.Equal(1, count)
}

func TestController_UpdateBookCourses(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeBooktrak, auth.ScopeBooktrakWrite, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()

	book := &models.Book{
		Title: "C Programming Language",
		Isbn:  "9780133086218",
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

	assert.NoError(db.Create(book).Error)
	assert.NoError(db.Create(course).Error)

	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	// Update book with invalid data (expect failure)
	apiError := lib.ErrorRequestDataValidationFailed
	resp, err := utils.DoHTTPReq(router, http.MethodPatch, fmt.Sprintf("/books/%d", book.ID), bytes.NewBufferString(`{"courseIDs": []}`))
	assert.NoError(err)
	assert.Equal(apiError.HTTPCode, resp.Code)
	assert.Equal(apiError.Code, utils.GetHTTPDataResp(assert, resp.Body.Bytes()).Error.ErrorCode)

	// Update book with invalid course id (expect failure)
	apiError = lib.ErrorBookCourseNotFound
	resp, err = utils.DoHTTPReq(router, http.MethodPatch, fmt.Sprintf("/books/%d", book.ID), bytes.NewBufferString(`{"courseIDs": [0]}`))
	assert.NoError(err)
	assert.Equal(apiError.HTTPCode, resp.Code)
	assert.Equal(apiError.Code, utils.GetHTTPDataResp(assert, resp.Body.Bytes()).Error.ErrorCode)

	params := UpdateBookCoursesParams{
		CourseIDs: []uint{course.ID},
	}
	paramsData, err := json.Marshal(params)
	assert.NoError(err)

	// Update book with valid data (expect success)
	resp, err = utils.DoHTTPReq(router, http.MethodPatch, fmt.Sprintf("/books/%d", book.ID), bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(http.StatusOK, resp.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, resp.Body.Bytes())
	assert.Nil(respData.Error)
	var respBook models.Book
	assert.NoError(json.Unmarshal(respData.Data, &respBook))

	book.Courses = []*models.Course{
		{
			BaseSchema: models.BaseSchema{
				ID: 1,
			},
			Number:        "136",
			AreaOfStudyID: lib.UIntToPtr(1),
		},
	}

	// Check if correct book
	assert.Equal(book.Title, respBook.Title)
	assert.Equal(book.Isbn, respBook.Isbn)
	assert.Equal(book.Authors, respBook.Authors)
	assert.Equal(book.Publisher, respBook.Publisher)
	assert.Equal(book.Courses, respBook.Courses)
}
