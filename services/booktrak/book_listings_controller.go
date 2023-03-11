package booktrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

type CreateBookListing struct {
	BookID       uint    `json:"bookID" binding:"required"`
	Condition    uint    `json:"condition" binding:"required"`
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

	exists, err := t.bookModel.DoesBookExistByID(createData.BookID)
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

type GetAllBookListings struct {
	CourseID uint `form:"courseID"`
}

func (t *Controller) ListBookListings(c *gin.Context) {}

func (t *Controller) GetBookListing(c *gin.Context) {}
