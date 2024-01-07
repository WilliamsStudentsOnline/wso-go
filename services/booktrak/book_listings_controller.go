package booktrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

type CreateBookListingParams struct {
	BookID      uint               `json:"bookID" binding:"required"`
	Condition   models.Condition   `json:"condition"`
	Description *string            `json:"description,omitempty"`
	ListingType models.ListingType `json:"listingType" binding:"required"`
}

type BookListingParamsBody struct {
	BookID       uint    `json:"bookID"`
	Condition    int     `json:"condition"` // Assuming the condition is an integer in JSON
	Description  *string `json:"description,omitempty"`
	IsBuyListing bool    `json:"isBuyListing"`
}

// CreateBookListing @Summary Create book listing
// @Description create a book listing
// @ID booktrak-create-book-listing
// @Tags booktrak
// @Accept  json
// @Produce  json
// @Param createParams body booktrak.BookListingParamsBody true "Create Book Listing Params"
// @Success 201 {object} models.BookListing
// @Failure 2232 {object} services.BaseErrorResponse "invalid book condition"
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /booktrak/listings [post]
func (t *Controller) CreateBookListing(c *gin.Context) {
	userID := services.GetUserID(c)

	reqBody := BookListingParamsBody{}
	reqErr := c.ShouldBind(&reqBody)
	if reqErr != nil {
		t.RespondBadBind(c, reqErr)
		return
	}

	// Bind create params
	createData := CreateBookListingParams{}
	createData.BookID = reqBody.BookID
	createData.Description = reqBody.Description

	// set condition based on condition # passed in
	switch reqBody.Condition {
	case 0:
		createData.Condition = models.ConditionPoor
	case 1:
		createData.Condition = models.ConditionFair
	case 2:
		createData.Condition = models.ConditionGood
	case 3:
		createData.Condition = models.ConditionVeryGood
	case 4:
		createData.Condition = models.ConditionLikeNew
	case 5:
		createData.Condition = models.ConditionNew
	default:
		createData.Condition = models.ConditionUndefined
		t.RespondAPIError(c, lib.ErrorMalformedRequestData)
	}

	// Set ListingType based on isBuyListing
	if reqBody.IsBuyListing {
		createData.ListingType = models.ListingTypeBuy
	} else {
		createData.ListingType = models.ListingTypeSell
	}

	exists, err := t.bookModel.DoesBookExist(models.BookIdentifier{Id: lib.UIntToPtr(createData.BookID)})
	if err != nil {
		t.RespondAPIError(c, lib.ErrorInternalServerError)
		return
	}
	if !exists {
		t.RespondAPIError(c, lib.ErrorBookNotFound)
		return
	}

	if createData.Condition == models.ConditionUndefined {
		t.RespondAPIError(c, lib.ErrorBookListingInvalidCondition)
		return
	}

	bookListing := models.BookListing{
		BookID:      createData.BookID,
		UserID:      userID,
		Condition:   createData.Condition,
		Description: createData.Description,
		ListingType: createData.ListingType,
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

// ListBookListings List all book listings.
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
// @Param isbn query string false "Book ISBN-13"
// @Param minCondition query models.Condition false "Minimum Book Condition"
// @Param maxCondition query models.Condition false "Maximum Book Condition"
// @Param listingType query models.ListingType false "Type of the listing"
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

// GetBookListing Get book listing by id
// GetBookListing godoc
// @Summary Get book listing by book listing id
// @Description get a book listing by book listing id
// @ID get-book-listing
// @Tags booktrak
// @Accept  json
// @Produce  json
// @Param bookListingID path uint true "Book Listing ID"
// @Success 200 {object} models.BookListing
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /listings/{bookListingID} [get]
func (t *Controller) GetBookListing(c *gin.Context) {
	bookListingID, err := services.GetUIntParam(c, "bookListingID")
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	var book models.BookListing
	err = t.bookListingModel.GetBookListingByID(bookListingID, &book)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, book)
}

// DeleteBookListing Delete book listing. Can either do this to self if a user, or to everything if admin
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
