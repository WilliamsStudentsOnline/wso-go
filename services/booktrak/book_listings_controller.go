package booktrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

type CreateBookListingParams struct {
	BookID       uint    `json:"bookID" binding:"required"`
	Condition    uint    `json:"condition"`
	Description  *string `json:"description,omitempty"`
	IsBuyListing bool    `json:"isBuyListing" binding:"required"`
}

// @Summary Create book listing
// @Description create a book listing
// @ID booktrak-create-book-listing
// @Tags booktrak
// @Accept  json
// @Produce  json
// @Param createParams body booktrak.CreateBookListingParams true "Create Book Listing Params"
// @Success 201 {object} models.BookListing
// @Failure 2232 {object} services.BaseErrorResponse "invalid book condition"
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /booktrak/listings [post]
func (t *Controller) CreateBookListing(c *gin.Context) {
	userID := services.GetUserID(c)

	// Bind create params
	createData := CreateBookListingParams{}
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

// List all book listings.
// ListBookListings godoc
// @Summary List book listings
// @Description lists all book listings
// @ID booktrak-list-book-listings
// @Tags booktrak
// @Accept  json
// @Produce  json
// @Param offset query int false "Offset Pagination"
// @Param limit query int false "Limit Pagination"
// @Param courseID query int false "Course ID"
// @Param userID query int false "User ID"
// @Param ISBN_10 query string false "Book ISBN-10 (must be in ISBN format"
// @Param ISBN_13 query string false "Book ISBN-13 (must be in ISBN format"
// @Param minCondition query int false "Minimum Book Condition"
// @Param maxCondition query int false "Maximum Book Condition"
// @Param isBuyListing query bool false "Whether the listing is a buy listing or not (a sell listing)"
// @Success 200 {array} models.BookListing
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /booktrak/listings [get]
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
		return
	}

	t.RespondOK(c, bookListings)
}

func (t *Controller) GetBookListing(c *gin.Context) {}

// Delete book listing. Can either do this to self if a user, or to everything if admin
// @Summary Delete book listing
// @Description delete a book listing
// @ID booktrak-delete-book-listing
// @Tags booktrak,booktrak-admin,admin
// @Accept  json
// @Produce  json
// @Param bookListingID path uint true "Book Listing ID"
// @Success 200 {object} models.BookListing
// @Failure 1331 {object} services.BaseErrorResponse "must be self"
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /booktrak/listings/{bookListingID} [delete]
func (t *Controller) DeleteBookListing(c *gin.Context) {
	userID := services.GetUserID(c)

	bookListingID, err := services.GetUIntParam(c, "bookListingID")
	if err != nil {
		t.RespondError(c, err)
		return
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
