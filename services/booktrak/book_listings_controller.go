package booktrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

type CreateBookListing struct {
	BookID       uint    `json:"bookID" binding:"required"`
	Condition    uint    `json:"condition"`
	Description  *string `json:"description,omitempty"`
	IsBuyListing bool    `json:"isBuyListing" binding:"required"`
}

func (t *Controller) CreateBookListing(c *gin.Context) {
	userID := services.GetUserID(c)

	// Bind create params
	createData := CreateBookListing{}
	err := c.ShouldBind(&createData)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	exists, err := t.bookModel.DoesBookExist(createData.BookID)
	if err != nil {
		t.RespondAPIError(c, lib.ErrorInternalServerError)
		return
	}
	if !exists {
		t.RespondAPIError(c, lib.ErrorBookNotFound)
		return
	}

	if !(createData.Condition < models.ConditionMAX) {
		t.RespondAPIError(c, lib.ErrorBookListingInvalidCondition)
		return
	}

	bookListing := models.BookListing{
		BookID:       createData.BookID,
		UserID:       userID,
		Condition:    createData.Condition,
		Description:  createData.Description,
		IsBuyListing: createData.IsBuyListing,
	}

	err = t.bookListingModel.CreateBookListing(&bookListing)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondCreated(c, bookListing)
}

type ListBookListingsParams struct {
	models.GetAllBookListingsOptions
}

func (t *Controller) ListBookListings(c *gin.Context) {
	var bookListings []*models.BookListing
	var err error

	opts := ListBookListingsParams{}
	err = c.ShouldBindQuery(&opts)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	err = t.bookListingModel.GetAllBookListings(&bookListings, &opts.GetAllBookListingsOptions)
	if err != nil {
		t.RespondError(c, err)
	}

	t.RespondOK(c, bookListings)
}

func (t *Controller) GetBookListing(c *gin.Context) {}

func (t *Controller) DeleteBookListing(c *gin.Context) {
	userID := services.GetUserID(c)

	bookListingID, err := services.GetUIntParam(c, "bookListingID")
	if err != nil {
		t.RespondError(c, err)
	}

	var bookListing models.BookListing
	err = t.bookListingModel.GetBookListingByID(bookListingID, &bookListing)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	if bookListing.UserID != userID {
		t.RespondError(c, lib.ErrorMustBeSelf)
		return
	}

	err = t.bookListingModel.DeleteBookListing(&bookListing)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	c.Set(services.UpdateTokenKey, true)

	t.RespondOK(c, bookListing)
}
