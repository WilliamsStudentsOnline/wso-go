package booktrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/isbn"
	sql_types "github.com/WilliamsStudentsOnline/wso-go/lib/sql_types"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	books "google.golang.org/api/books/v1"
)

type ListBooksParams struct {
	models.GetAllBooksOptions
}

// ListBooks List all books
// @Summary List books
// @Description lists all books. Order by creation date.
// @ID booktrak-list-books
// @Tags booktrak
// @Accept  json
// @Produce  json
// @Param offset query int false "Offset Pagination"
// @Param limit query int false "Limit Pagination"
// @Param title query string false "Book Title"
// @Param publisher query string false "Book Publisher"
// @Param isbn query string false "Book ISBN-13 (must be in ISBN format)"
// @Success 200 {array} models.Book
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /booktrak/books [get]
func (t *Controller) ListBooks(c *gin.Context) {
	var books []*models.Book
	var err error

	opts := ListBooksParams{}
	err = c.ShouldBind(&opts)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	err = t.bookModel.GetAllBooks(&books, &opts.GetAllBooksOptions)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, books)
}

type CreateBookParams struct {
	ISBN string `json:"isbn" binding:"required,isbn"`
}

// CreateBook
// @Summary Create a book if it doesn't exist already
// @Description create a book
// @ID booktrak-create-book
// @Tags booktrak
// @Accept  json
// @Produce  json
// @Param createParams body booktrak.CreateBookParams true "Create Book Params"
// @Success 201 {object} models.Book
// @Failure 2251 {object} services.BaseErrorResponse "failed to find book online by isbn"
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /booktrak/books [put]
func (t *Controller) CreateBook(c *gin.Context) {
	params := CreateBookParams{}
	err := c.ShouldBind(&params)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	params.ISBN = isbn.CleanISBN(params.ISBN)
	volumes, err := t.searchVolumes("isbn:"+params.ISBN, lib.IntToPtr(20))
	if err != nil {
		t.RespondAPIError(c, lib.ErrorInternalServerError)
		return
	}

	isbn10, isbn13 := "", ""
	var correctVolume *books.Volume
	for _, volume := range volumes.Items {
		for _, v := range volume.VolumeInfo.IndustryIdentifiers {
			if params.ISBN == v.Identifier {
				if v.Type == "ISBN_10" {
					isbn10 = v.Identifier
				}
				if v.Type == "ISBN_13" {
					isbn13 = v.Identifier
				}
				correctVolume = volume
				break
			}
		}
	}

	if correctVolume == nil {
		t.RespondAPIError(c, lib.ErrorBookNotFoundByISBN)
		return
	}

	if params.ISBN != isbn10 && params.ISBN != isbn13 {
		t.RespondAPIError(c, lib.ErrorBookNotFoundByISBN)
		return
	}

	if isbn13 == "" {
		isbn13 = isbn.ConvertIsbn10to13(params.ISBN)
	}

	volume := correctVolume.VolumeInfo
	book := &models.Book{
		Title:     volume.Title,
		Subtitle:  lib.StrToPtr(volume.Subtitle),
		Authors:   sql_types.CSV(volume.Authors),
		Publisher: lib.StrToPtr(volume.Publisher),
		Isbn:      isbn13,
		InfoLink:  lib.StrToPtr(volume.InfoLink),
		ImageLink: lib.StrToPtr(volume.ImageLinks.Thumbnail),
	}

	err = t.bookModel.CreateBook(book)
	if err != nil {
		t.RespondAPIError(c, lib.ErrorInternalServerError)
		return
	}

	t.RespondCreated(c, book)
}

type UpdateBookCoursesParams struct {
	CourseIDs []uint `json:"courseIDs" binding:"required,min=1"`
}

// UpdateBookCourses Update book courses
// @Summary Update book
// @Description update a book's courses
// @ID booktrak-update-book
// @Tags booktrak
// @Accept  json
// @Produce  json
// @Param updateParams body booktrak.UpdateBookCoursesParams true "Update Book Params"
// @Param bookID path uint true "Book ID"
// @Success 200 {object} models.Book
// @Failure 2253 {object} services.BaseErrorResponse "some courses could not be found"
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /booktrak/books/{bookID} [patch]
func (t *Controller) UpdateBookCourses(c *gin.Context) {
	bookID, err := services.GetUIntParam(c, "bookID")
	if err != nil {
		t.RespondError(c, err)
		return
	}

	params := UpdateBookCoursesParams{}
	err = c.ShouldBind(&params)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	book, err := t.bookModel.AddCoursesToBook(bookID, &params.CourseIDs)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, book)
}

// GetBook Get book by id
// @Summary Get book by book id
// @Description get a book by book id
// @ID get-book
// @Tags books
// @Accept  json
// @Produce  json
// @Param bookID path uint true "Book ID"
// @Success 200 {object} models.Book
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /books/{bookID} [get]
func (t *Controller) GetBook(c *gin.Context) {
	bookID, err := services.GetUIntParam(c, "bookID")
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	var book models.Book
	err = t.bookModel.GetBookByID(bookID, &book)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, book)
}
