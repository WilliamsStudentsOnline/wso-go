package user

import (
	"errors"
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type Controller struct {
	services.BaseController
	userModel *models.UserModel
}

// Construct a new user controller
func NewController(db *gorm.DB) *Controller {
	return &Controller{
		userModel: models.NewUserModel(db),
	}
}

// Fetch all users
func (t *Controller) FetchAllUsers(c *gin.Context) {
	var users []models.User
	err := t.userModel.GetAllUsers(&users)

	if err != nil {
		t.RespondErrorCode(c, http.StatusInternalServerError, err)
		return
	}

	t.RespondOK(c, users)
}

// Get user by id. Pass "me" if you want to get self
func (t *Controller) GetUser(c *gin.Context) {
	// Decode userID or self.
	userID, err := getUserIDParamOrSelf(c)
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
	}

	// Do database query
	var user models.User
	err = t.userModel.GetUserByID(userID, &user)
	if err != nil {
		t.RespondErrorCode(c, http.StatusInternalServerError, err)
		return
	}

	if !user.Visible {
		t.RespondAPIError(c, lib.ErrorUserNotVisible)
		return
	}

	if !user.AtWilliams {
		t.RespondAPIError(c, lib.ErrorUserNotAtWilliams)
		return
	}

	t.RespondOK(c, user)
}

func (t *Controller) UpdateUser(c *gin.Context) {
	// Decode userID or self.
	userID, err := getUserIDParamOrSelf(c)
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
	}

	// Must only be able to update self
	if userID != services.GetUserID(c) {
		t.RespondErrorCode(c, http.StatusForbidden, lib.ErrorMustBeSelf)
		return
	}

	// Bind update params
	var update map[string]interface{}
	err = c.ShouldBind(&update)
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, errors.New("could not parse malformed request data"))
		return
	}

	// Update the user in the db
	err = t.userModel.UpdateUser(userID, update)
	if err != nil {
		t.RespondErrorCode(c, http.StatusInternalServerError, err)
		return
	}

	// Return nothing
	t.RespondOK(c, nil)
}

func (t *Controller) UpdateUserTags(c *gin.Context) {
	// Decode userID or self.
	userID, err := getUserIDParamOrSelf(c)
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
	}

	// Must only be able to update self
	if userID != services.GetUserID(c) {
		t.RespondErrorCode(c, http.StatusForbidden, lib.ErrorMustBeSelf)
		return
	}

	// Bind update params
	var update []string
	err = c.ShouldBind(&update)
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, errors.New("could not parse malformed request data"))
		return
	}

	// Update the user in the db
	err = t.userModel.UpdateUserTags(userID, update)
	if err != nil {
		if err.Error() == "invalid user tag" {
			t.RespondErrorCode(c, http.StatusBadRequest, err)
			return
		}

		t.RespondErrorCode(c, http.StatusInternalServerError, err)
		return
	}

	// Return nothing
	t.RespondOK(c, nil)
}

// Decode userID from passed param or get self's userID if param="me".
func getUserIDParamOrSelf(c *gin.Context) (uint, error) {
	userIDStr := c.Param("userID")

	var userID uint
	var err error

	// Decode userID or self.
	if userIDStr == "me" {
		userID = services.GetUserID(c)
	} else {
		userID, err = services.GetUIntParam(c,"userID")
		if err != nil {
			return 0, errors.New("could not parse user id")
		}
	}

	return userID, nil
}