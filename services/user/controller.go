package user

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	search "github.com/WilliamsStudentsOnline/wso-go/lib/search/users"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type Controller struct {
	services.BaseController
	userModel  *models.UserModel
	userSearch search.SearchUsers
}

// Construct a new user controller
func NewController(db *gorm.DB, cfg *config.Config) *Controller {
	return &Controller{
		userModel:  models.NewUserModel(db),
		userSearch: search.NewSearchUsers(db, cfg),
	}
}

// ListUsers godoc
// @Summary List users
// @Description Get all users that are visible and at williams.
// @Description If you pass a search query (?q="blah"), you will get all users matching that search query
// @ID list-users
// @Tags users
// @Accept  json
// @Produce  json
// @Param offset query int false "Offset Pagination"
// @Param limit query int false "Limit Pagination"
// @Param preload query []string false "Preload List"
// @Param q query string false "Search Query"
// @Success 200 {array} models.User
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /users [get]
func (t *Controller) ListUsers(c *gin.Context) {
	var users []*models.User
	var err error

	opts := models.GetAllUsersOptions{}
	if err = c.ShouldBindQuery(&opts); err != nil {
		t.RespondError(c, err)
		return
	}

	if query, ok := c.GetQuery("q"); ok {
		err = t.userSearch.Search(query, &users, &opts)
	} else {
		err = t.userModel.GetAllUsers(&users, &opts)
	}

	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, users)
}

// Get user by id. Pass "me" if you want to get self
// GetUser godoc
// @Summary Get user by user id
// @Description get a user by user id that is visible and at williams. Also loads tags. Pass "me" if you want to get self
// @ID get-user
// @Tags users
// @Accept  json
// @Produce  json
// @Param userID path uint true "User ID"
// @Success 200 {object} models.User
// @Failure 1403 {object} lib.APIError "user not visible"
// @Failure 1404 {object} lib.APIError "user not at williams"
// @Failure 1405 {object} lib.APIError "user id could not be parsed"
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /users/{userID} [get]
func (t *Controller) GetUser(c *gin.Context) {
	// Decode userID or self.
	userID, err := getUserIDParamOrSelf(c)
	if err != nil {
		t.RespondError(c, err)
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

// UpdateUser godoc
// @Summary Update user by user id
// @Description updates a user by user id. You may only update yourself. You may pass "me" to get self as well.
// @ID update-user
// @Tags users
// @Accept  json
// @Produce  json
// @Param userID path uint true "User ID"
// @Param updateParams body user.UpdateUserParams true "Update User Parameters"
// @Success 200 {object} models.User
// @Failure 1405 {object} lib.APIError "user id could not be parsed"
// @Failure 1331 {object} lib.APIError "must be self"
// @Failure 1100 {object} lib.APIError "could not parse malformed request data"
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /users/{userID} [patch]
func (t *Controller) UpdateUser(c *gin.Context) {
	// Decode userID or self.
	userID, err := getUserIDParamOrSelf(c)
	if err != nil {
		t.RespondError(c, err)
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
		t.RespondError(c, lib.ErrorMalformedRequestData)
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

type UpdateUserTagsParams struct {
	Tags []string `json:"tags"`
}

// UpdateUserTags godoc
// @Summary Update user tags by user id
// @Description updates a user's tags by user id. You may only update yourself. You may pass "me" to get self as well.
// @ID update-user-tags
// @Tags users
// @Accept  json
// @Produce  json
// @Param userID path uint true "User ID"
// @Param updateTagsParams body user.UpdateUserTagsParams true "Update Tags Params"
// @Success 200 {object} models.User
// @Failure 1405 {object} lib.APIError "user id could not be parsed"
// @Failure 1331 {object} lib.APIError "must be self"
// @Failure 1100 {object} lib.APIError "could not parse malformed request data"
// @Failure 1406 {object} lib.APIError "invalid user tag"
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /users/{userID}/tags [put]
func (t *Controller) UpdateUserTags(c *gin.Context) {
	// Decode userID or self.
	userID, err := getUserIDParamOrSelf(c)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Must only be able to update self
	if userID != services.GetUserID(c) {
		t.RespondError(c, lib.ErrorMustBeSelf)
		return
	}

	// Bind update params
	var update UpdateUserTagsParams
	err = c.ShouldBind(&update)
	if err != nil {
		t.RespondError(c, lib.ErrorMalformedRequestData)
		return
	}

	// Update the user in the db
	err = t.userModel.UpdateUserTags(userID, update.Tags)
	if err != nil {
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
			return 0, lib.ErrorUserIDNoParse
		}
	}

	return userID, nil
}
