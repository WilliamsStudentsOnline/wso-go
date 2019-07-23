package user

import (
	"errors"
	"net/http"

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
		userModel: &models.UserModel{
			BaseModel: models.BaseModel{
				DB: db,
			},
		},
	}
}

// Fetch all users
func (t *Controller) FetchAllUsers(c *gin.Context) {
	var users []models.User
	err := t.userModel.GetAllUsers(&users)

	if err != nil {
		t.RespondError(http.StatusInternalServerError, err, c)
		return
	}

	t.RespondOK(users, c)
}

// Get user by id. Pass "me" if you want to get self
func (t *Controller) GetUser(c *gin.Context) {
	userIDStr := c.Param("user_id")

	var userID uint
	var err error

	// Decode userID or self.
	if userIDStr == "me" {
		userID = (c.MustGet("user_id")).(uint)
	} else {
		userID, err = services.GetUIntParam("user_id", c)
		if err != nil {
			t.RespondError(http.StatusBadRequest, errors.New("could not parse user id"), c)
			return
		}
	}

	// Do database query
	var user models.User
	err = t.userModel.GetUserByID(uint(userID), &user)
	if err != nil {
		t.RespondError(http.StatusInternalServerError, err, c)
		return
	}

	t.RespondOK(user, c)
}

func (t *Controller) UpdateUser(c *gin.Context) {
	// Decode parameter
	userID, err := services.GetUIntParam("user_id", c)
	if err != nil {
		t.RespondError(http.StatusBadRequest, errors.New("could not parse user id"), c)
		return
	}

	// Must only be able to update self
	if userID != services.GetUserID(c) {
		t.RespondError(http.StatusForbidden, errors.New("can only update self"), c)
		return
	}

	// Bind update params
	var update map[string]interface{}
	err = c.ShouldBind(&update)
	if err != nil {
		t.RespondError(http.StatusBadRequest, errors.New("could not bind update params"), c)
		return
	}

	// Update the user in the db
	err = t.userModel.UpdateUser(userID, update)
	if err != nil {
		t.RespondError(http.StatusInternalServerError, err, c)
		return
	}

	// Return nothing
	t.RespondOK(nil, c)
}
