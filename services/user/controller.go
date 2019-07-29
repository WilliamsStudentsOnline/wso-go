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

// List users
func (t *Controller) ListUsers(c *gin.Context) {
	var users []models.User
	err := t.userModel.GetAllUsers(&users)

	if err != nil {
		t.RespondError(c, err)
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
		return
	}

	// Do database query
	var user models.User
	err = t.userModel.GetUserByID(userID, &user)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	if !*user.Visible {
		t.RespondError(c, lib.ErrorUserNotVisible)
		return
	}

	if !*user.AtWilliams {
		t.RespondError(c, lib.ErrorUserNotAtWilliams)
		return
	}

	t.RespondOK(c, user)
}

type UpdateUserParams struct {
	Visible                   *bool   `json:"visible"`
	DormVisible               *bool   `json:"dormVisible"`
	HomeVisible               *bool   `json:"homeVisible"`
	Pronoun                   *string `json:"pronoun"`
	OffCycle                  *bool   `json:"offCycle"`
	HasAcceptedFactrakPolicy  *bool   `json:"hasAcceptedFactrakPolicy"`
	HasAcceptedDormtrakPolicy *bool   `json:"hasAcceptedDormtrakPolicy"`
}

func (t *Controller) UpdateUser(c *gin.Context) {
	// Decode userID or self.
	userID, err := getUserIDParamOrSelf(c)
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Must only be able to update self
	if userID != services.GetUserID(c) {
		t.RespondError(c, lib.ErrorMustBeSelf)
		return
	}

	// Bind update params
	updateData := UpdateUserParams{}
	err = c.ShouldBind(&updateData)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Do database query to get the user
	var user models.User
	err = t.userModel.GetUserByID(userID, &user)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Update fields: this is a bit long and verbose, but I don't want to mess with reflect
	user.Visible = lib.BoolPtrDefaults(updateData.Visible, user.Visible)
	user.DormVisible = lib.BoolPtrDefaults(updateData.DormVisible, user.DormVisible)
	user.HomeVisible = lib.BoolPtrDefaults(updateData.HomeVisible, user.HomeVisible)
	user.Pronoun = lib.StrPtrDefaults(updateData.Pronoun, user.Pronoun)
	user.OffCycle = lib.BoolPtrDefaults(updateData.OffCycle, user.OffCycle)
	user.HasAcceptedFactrakPolicy = lib.BoolPtrDefaults(updateData.HasAcceptedFactrakPolicy, user.HasAcceptedFactrakPolicy)
	user.HasAcceptedDormtrakPolicy = lib.BoolPtrDefaults(updateData.HasAcceptedDormtrakPolicy, user.HasAcceptedDormtrakPolicy)

	// Update the user in the db
	err = t.userModel.UpdateUser(&user)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Return updated user
	t.RespondOK(c, user)
}

func (t *Controller) UpdateUserTags(c *gin.Context) {
	// Decode userID or self.
	userID, err := getUserIDParamOrSelf(c)
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Must only be able to update self
	if userID != services.GetUserID(c) {
		t.RespondError(c, lib.ErrorMustBeSelf)
		return
	}

	// Bind update params
	var update []string
	err = c.ShouldBind(&update)
	if err != nil {
		t.RespondError(c, lib.ErrorMalformedRequestData)
		return
	}

	// Update the user in the db
	err = t.userModel.UpdateUserTags(userID, update)
	if err != nil {
		// TODO: make this error an API error
		if err.Error() == "invalid user tag" {
			t.RespondErrorCode(c, http.StatusBadRequest, err)
			return
		}

		t.RespondError(c, err)
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
		userID, err = services.GetUIntParam(c, "userID")
		if err != nil {
			// TODO: make this an API error
			return 0, errors.New("could not parse user id")
		}
	}

	return userID, nil
}
