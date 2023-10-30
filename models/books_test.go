package models_test

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	. "github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestBookModel_GetAllBooks(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewBookModel(db, zaptest.NewLogger(t).Sugar())

	books := []Book{
		{
			Title: "The C Programming Language",
			Authors: []string{
				"Brian W. Kernighan",
				"Dennis M. Ritchie",
			},
			Isbn:     "9780131103627",
			InfoLink: lib.StrToPtr("https://books.google.com/books/about/The_C_Programming_Language.html?id=HHhGAAAAYAAJ"),
		},
		{
			Title: "The Go Programming Language",
			Authors: []string{
				"Alan A. A. Donovan",
				"Brian W. Kernighan",
			},
			Isbn: "9780134190440",
		},
	}

	for i := range books {
		assert.NoError(db.Create(&books[i]).Error)
	}

	var res []*Book
	assert.NoError(m.GetAllBooks(&res, nil))
	for i := range books {
		assert.Equal(books[i].ID, res[i].ID)
		assert.Equal(books[i].Title, res[i].Title)
		assert.Equal(books[i].Authors, res[i].Authors)
		assert.Equal(books[i].Isbn, res[i].Isbn)
		assert.Equal(books[i].InfoLink, res[i].InfoLink)
	}
}

func TestBookModel_GetBookByID(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewBookModel(db, zaptest.NewLogger(t).Sugar())

	book := Book{
		Title: "The C Programming Language",
		Authors: []string{
			"Brian W. Kernighan",
			"Dennis M. Ritchie",
		},
		Isbn:     "9780131103627",
		InfoLink: lib.StrToPtr("https://books.google.com/books/about/The_C_Programming_Language.html?id=HHhGAAAAYAAJ"),
	}

	assert.NoError(db.Create(&book).Error)

	var res Book
	assert.NoError(m.GetBookByID(1, &res))
	assert.Equal(book.ID, res.ID)
	assert.Equal(book.Title, res.Title)
	assert.Equal(book.Authors, res.Authors)
	assert.Equal(book.Isbn, res.Isbn)
	assert.Equal(book.InfoLink, res.InfoLink)
}

func TestBookModel_DoesBookExit(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewBookModel(db, zaptest.NewLogger(t).Sugar())

	book := Book{
		BaseSchema: BaseSchema{
			ID: 1,
		},
		Title: "The C Programming Language",
		Authors: []string{
			"Brian W. Kernighan",
			"Dennis M. Ritchie",
		},
		Isbn:     "9780131103627",
		InfoLink: lib.StrToPtr("https://books.google.com/books/about/The_C_Programming_Language.html?id=HHhGAAAAYAAJ"),
	}

	assert.NoError(db.Create(&book).Error)

	exists, err := m.DoesBookExist(BookIdentifier{Id: lib.UIntToPtr(1)})
	assert.NoError(err)
	assert.True(exists)

	exists, err = m.DoesBookExist(BookIdentifier{Isbn: lib.StrToPtr("9780131103627")})
	assert.NoError(err)
	assert.True(exists)

	exists, err = m.DoesBookExist(BookIdentifier{Id: lib.UIntToPtr(2)})
	assert.NoError(err)
	assert.False(exists)

	exists, err = m.DoesBookExist(BookIdentifier{})
	assert.Error(err)
	assert.False(exists)
}

func TestBookModel_CreateBook(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewBookModel(db, zaptest.NewLogger(t).Sugar())

	book := Book{
		Title: "The C Programming Language",
		Authors: []string{
			"Brian W. Kernighan",
			"Dennis M. Ritchie",
		},
		Isbn:     "9780131103627",
		InfoLink: lib.StrToPtr("https://books.google.com/books/about/The_C_Programming_Language.html?id=HHhGAAAAYAAJ"),
	}

	assert.NoError(m.CreateBook(&book))

	var res Book
	assert.NoError(db.Find(&res).First(&res).Error)
	assert.Equal(book.ID, res.ID)
	assert.Equal(book.Title, res.Title)
	assert.Equal(book.Authors, res.Authors)
	assert.Equal(book.Isbn, res.Isbn)
	assert.Equal(book.InfoLink, res.InfoLink)

	var count int
	assert.NoError(m.CreateBook(&book))
	assert.NoError(db.Model(&Book{}).Where("books.isbn == 9780131103627").Count(&count).Error)
	assert.Equal(1, count)

}

func TestBookModel_AddCoursesToBook(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewBookModel(db, zaptest.NewLogger(t).Sugar())

	books := []Book{
		{
			BaseSchema: BaseSchema{
				ID: 1,
			},
			Title: "The C Programming Language",
			Authors: []string{
				"Brian W. Kernighan",
				"Dennis M. Ritchie",
			},
			Isbn:     "9780131103627",
			InfoLink: lib.StrToPtr("https://books.google.com/books/about/The_C_Programming_Language.html?id=HHhGAAAAYAAJ"),
		},
		{
			BaseSchema: BaseSchema{
				ID: 2,
			},
			Title: "The Go Programming Language",
			Authors: []string{
				"Alan A. A. Donovan",
				"Brian W. Kernighan",
			},
			Isbn: "9780134190440",
		},
	}

	for i := range books {
		assert.NoError(db.Create(&books[i]).Error)
	}

	courses := []Course{
		{
			BaseSchema: BaseSchema{
				ID: 1,
			},
			Number:        "256",
			AreaOfStudyID: lib.UIntToPtr(1),
		},
		{
			BaseSchema: BaseSchema{
				ID: 2,
			},
			Number:        "120",
			AreaOfStudyID: lib.UIntToPtr(2),
		},
	}
	for i := range courses {
		assert.NoError(db.Create(&courses[i]).Error)
	}

	courseIDs := []uint{1, 2}

	_, err := m.AddCoursesToBook(1, &courseIDs)
	assert.NoError(err)
	_, err = m.AddCoursesToBook(2, &courseIDs)
	assert.NoError(err)

	// Check Many2Many relationship
	for i := range books {
		var book Book
		assert.NoError(db.Preload("Courses").First(&book, books[i].ID).Error)
		assert.Equal(2, len(book.Courses))
		for j := range courses {
			assert.Equal(courses[j].ID, book.Courses[j].ID)
			assert.Equal(courses[j].Number, book.Courses[j].Number)
			assert.Equal(courses[j].AreaOfStudyID, book.Courses[j].AreaOfStudyID)
		}
	}

	for i := range courses {
		var course Course
		assert.NoError(db.Preload("Books").First(&course, courses[i].ID).Error)
		assert.Equal(2, len(course.Books))
		for j := range books {
			assert.Equal(books[j].ID, course.Books[j].ID)
			assert.Equal(books[j].Title, course.Books[j].Title)
			assert.Equal(books[j].Authors, course.Books[j].Authors)
			assert.Equal(books[j].Isbn, course.Books[j].Isbn)
			assert.Equal(books[j].InfoLink, course.Books[j].InfoLink)
		}
	}
}

func TestBookModel_DeleteBook(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewBookModel(db, zaptest.NewLogger(t).Sugar())

	book := Book{
		Title: "The C Programming Language",
		Authors: []string{
			"Brian W. Kernighan",
			"Dennis M. Ritchie",
		},
		Isbn:     "9780131103627",
		InfoLink: lib.StrToPtr("https://books.google.com/books/about/The_C_Programming_Language.html?id=HHhGAAAAYAAJ"),
	}

	assert.NoError(db.Create(&book).Error)

	assert.NoError(m.DeleteBook(&book))

	var count int
	assert.NoError(db.Model(&Book{}).Where("isbn = 9780131103627").Count(&count).Error)
	assert.Equal(0, count)
}
