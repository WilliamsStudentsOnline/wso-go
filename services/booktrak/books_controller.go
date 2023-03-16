package booktrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	sql_types "github.com/WilliamsStudentsOnline/wso-go/lib/sql_types"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

type ListBooksParams struct {
	models.GetAllBooksOptions
}

// List all books
// @Summary List books
// @Description lists all books . Order by creation date.
// @ID booktrak-list-books
// @Tags booktrak
// @Accept  json
// @Produce  json
// @Param offset query int false "Offset Pagination (timestamp)"
// @Param limit query int false "Limit Pagination"
// @Param title query string false "Book Title"
// @Param publisher query string false "Book Publisher"
// @Param ISBN_10 query string false "Book ISBN-10 (must be in ISBN format)"
// @Param ISBN_13 query string false "Book ISBN-13 (must be in ISBN format)"
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

type CreateOrUpdateBookParams struct {
	ISBN      string `json:"ISBN" binding:"required,isbn"`
	CourseIDs []uint `json:"courseIDs" binding:"required,min=1"`
}

// @Summary Create or Update book
// @Description create or update a book
// @ID booktrak-create-or-update-book
// @Tags booktrak
// @Accept  json
// @Produce  json
// @Param createParams body booktrak.CreateOrUpdateBookParams true "Create Or Update Book Params"
// @Success 201 {object} models.Book
// @Failure 2251 {object} services.BaseErrorResponse "failed to find book online by isbn"
// @Failure 2253 {object} services.BaseErrorResponse "some courses could not be found"
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /booktrak/books [post]
func (t *Controller) CreateOrUpdateBook(c *gin.Context) {
	params := CreateOrUpdateBookParams{}
	err := c.ShouldBind(&params)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	params.ISBN = models.CleanISBN(params.ISBN)

	volumes, err := t.searchVolumes(params.ISBN, lib.IntToPtr(1))
	if err != nil {
		t.RespondAPIError(c, lib.ErrorInternalServerError)
		return
	}
	if len(volumes.Items) != 1 {
		t.RespondAPIError(c, lib.ErrorBookNotFoundByISBN)
		return
	}

	isbn10, isbn13 := "", ""
	for _, v := range volumes.Items[0].VolumeInfo.IndustryIdentifiers {
		if v.Type == "ISBN_10" {
			isbn10 = v.Identifier
		}
		if v.Type == "ISBN_13" {
			isbn13 = v.Identifier
		}
	}
	if params.ISBN != isbn10 && params.ISBN != isbn13 {
		t.RespondAPIError(c, lib.ErrorBookNotFoundByISBN)
		return
	}

	volume := volumes.Items[0].VolumeInfo
	book := &models.Book{
		Title:     volume.Title,
		Subtitle:  volume.Subtitle,
		Authors:   sql_types.CSV(volume.Authors),
		Publisher: volume.Publisher,
		ISBN_10:   isbn10,
		ISBN_13:   isbn13,
		InfoLink:  volume.InfoLink,
		ImageLink: volume.ImageLinks.Thumbnail,
	}

	// If the courses aren't valid, we don't create a book
	for _, course := range params.CourseIDs {
		exists, err := t.courseModel.DoesCourseExist(course)
		if err != nil {
			t.RespondAPIError(c, lib.ErrorInternalServerError)
			return
		}
		if !exists {
			t.RespondAPIError(c, lib.ErrorBookCourseNotFound)
			return
		}
	}

	err = t.bookModel.CreateBook(book)
	if err != nil {
		t.RespondAPIError(c, lib.ErrorInternalServerError)
		return
	}

	book, err = t.bookModel.AddCoursesToBook(book.ID, &params.CourseIDs)
	if err != nil {
		t.RespondAPIError(c, lib.ErrorInternalServerError)
		return
	}

	t.RespondCreated(c, book)
}

type UpdateBookCoursesParams struct {
	CourseIDs []uint `json:"courseIDs" binding:"required,min=1"`
}

// Update book courses
// @Summary Update book
// @Description update a book's courses
// @ID booktrak-update-book
// @Tags booktrak
// @Accept  json
// @Produce  json
// @Param updateParams body booktrak.UpdateBookCourses true "Update Book Params"
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

func (t *Controller) GetBook(c *gin.Context) {}
