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
	Title     string   `json:"title" binding:"required"`
	Subtitle  string   `json:"subtitle"`
	Authors   []string `json:"authors"`
	Publisher string   `json:"publisher,omitempty"`
	ISBN_10   string   `json:"ISBN_10" binding:"required,len=10"`
	ISBN_13   string   `json:"ISBN_13" binding:"required,len=13"`
	InfoLink  string   `json:"infoLink"`
	ImageLink string   `json:"imageLink"`

	CourseIDs []uint `json:"courseIDs" binding:"required,min=1"`
}

func (t *Controller) CreateOrUpdateBook(c *gin.Context) {
	createData := CreateOrUpdateBookParams{}
	err := c.ShouldBind(&createData)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	book := models.Book{
		Title:     createData.Title,
		Subtitle:  createData.Subtitle,
		Authors:   sql_types.CSV(createData.Authors),
		Publisher: createData.Publisher,
		ISBN_10:   createData.ISBN_10,
		ISBN_13:   createData.ISBN_13,
		InfoLink:  createData.InfoLink,
		ImageLink: createData.ImageLink,
	}

	volumes, err := t.searchVolumes(book.ISBN_13, lib.IntToPtr(1))
	if err != nil {
		t.RespondAPIError(c, lib.ErrorBookNotFoundByISBN)
		return
	}

	if !bookMatchesOnlineData(book, volumes.Items[0].VolumeInfo) {
		t.RespondAPIError(c, lib.ErrorBookDoesNotMatchOnlineData)
		return
	}

	for _, course := range createData.CourseIDs {
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

	err = t.bookModel.AddCoursesToBook(book.ID, &createData.CourseIDs)
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
