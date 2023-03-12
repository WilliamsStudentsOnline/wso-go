package booktrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	sql_types "github.com/WilliamsStudentsOnline/wso-go/lib/sql_types"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/gin-gonic/gin"
)

type ListBooksParams struct {
	models.GetAllBooksOptions
}

func (t *Controller) ListBooks(c *gin.Context) {
	var books []*models.Book
	var err error

	opts := ListBooksParams{}
	err = c.ShouldBind(&opts)
	if err != nil {
		t.RespondBadBind(c, err)
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
	}

	volume := volumes.Items[0].VolumeInfo
	book := models.Book{
		Title:     volume.Title,
		Subtitle:  volume.Subtitle,
		Authors:   sql_types.CSV(volume.Authors),
		Publisher: volume.Publisher,
		ISBN_10:   isbn10,
		ISBN_13:   isbn13,
		InfoLink:  volume.InfoLink,
		ImageLink: volume.ImageLinks.Thumbnail,
	}

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

	err = t.bookModel.CreateBook(&book)
	if err != nil {
		t.RespondAPIError(c, lib.ErrorInternalServerError)
		return
	}

	err = t.bookModel.AddCoursesToBook(book.ID, &params.CourseIDs)
	if err != nil {
		t.RespondAPIError(c, lib.ErrorInternalServerError)
		return
	}

	t.RespondCreated(c, book)
}

type AddCoursesToBookParams struct {
	BookID    uint   `json:"bookID" binding:"required"`
	CourseIDs []uint `json:"coursesIDs" binding:"required,min=1"`
}

func (t *Controller) AddCoursesToBook(c *gin.Context) {
	params := AddCoursesToBookParams{}
	err := c.ShouldBind(&params)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	err = t.bookModel.AddCoursesToBook(params.BookID, &params.CourseIDs)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, nil)
}

func (t *Controller) GetBook(c *gin.Context) {}
