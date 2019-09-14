package user

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	search "github.com/WilliamsStudentsOnline/wso-go/lib/search/users"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/WilliamsStudentsOnline/wso-go/services/user/responses"
	"github.com/disintegration/imaging"
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

//go:generate go run github.com/WilliamsStudentsOnline/wso-go/lib/generate/service_responses/cmd -in responses/list_users.json -out responses/list_users.go

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
// @Success 200 {array} responses.ListUsersResponseUser
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

		if err != nil {
			t.RespondError(c, err)
			return
		}
	} else {
		err = t.userModel.GetAllUsers(&users, &opts)

		if err != nil {
			t.RespondError(c, err)
			return
		}

		count, err := t.userModel.CountAllUsers()
		if err != nil {
			t.RespondError(c, err)
			return
		}
		t.SetPaginationTotal(c, count)
	}

	t.RespondOK(c, responses.ConvertListUsersResponse(users))
}

//go:generate go run github.com/WilliamsStudentsOnline/wso-go/lib/generate/service_responses/cmd -in responses/get_user.json -out responses/get_user.go

// Get user by id. Pass "me" if you want to get self
// GetUser godoc
// @Summary Get user by user id
// @Description get a user by user id that is visible and at williams. Also loads tags. Pass "me" if you want to get self
// @ID get-user
// @Tags users
// @Accept  json
// @Produce  json
// @Param userID path uint true "User ID"
// @Success 200 {object} responses.GetUserResponseUser
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

	selfUserID := services.GetUserID(c)

	// If we are not self or admin, return iff user is not visible and is at Williams. Otherwise,
	// if we are admin or we are looking at self, return no matter what.
	if selfUserID != user.ID && !auth.HasScope(c, auth.ScopeAdminAll) {
		if !*user.Visible {
			t.RespondError(c, lib.ErrorUserNotVisible)
			return
		}

		if !*user.AtWilliams {
			t.RespondError(c, lib.ErrorUserNotAtWilliams)
			return
		}
	}

	t.RespondOK(c, responses.ConvertGetUserResponse(&user))
}

type UpdateUserParams struct {
	Visible                   *bool   `json:"visible"`
	DormVisible               *bool   `json:"dormVisible"`
	HomeVisible               *bool   `json:"homeVisible"`
	Pronoun                   *string `json:"pronoun"`
	OffCycle                  *bool   `json:"offCycle"`
	HasAcceptedFactrakPolicy  *bool   `json:"hasAcceptedFactrakPolicy"`
	HasAcceptedDormtrakPolicy *bool   `json:"hasAcceptedDormtrakPolicy"`
	Nickname                  *string `json:"nickname"`
	OptOutEphcatch            *bool   `json:"optOutEphcatch"`
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
	user.Nickname = lib.StrPtrDefaults(updateData.Nickname, user.Nickname)
	user.OptOutEphcatch = lib.BoolPtrDefaults(updateData.OptOutEphcatch, user.OptOutEphcatch)

	// Update the user in the db
	err = t.userModel.UpdateUser(&user)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Update token if we changed these
	if updateData.HasAcceptedFactrakPolicy != nil || updateData.HasAcceptedDormtrakPolicy != nil {
		t.SetUpdateToken(c)
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

// UploadProfilePhoto godoc
// @Summary Upload a profile photo by user id
// @Description upload a user's profile photo by user id. You may only update yourself. You may pass "me" to get self as well.
// @ID upload-profile-photo
// @Tags users
// @Accept  multipart/form-data
// @Produce  json
// @Param userID path uint true "User ID"
// @Param file formData file true "Profile Photo"
// @Success 200 {object} models.User
// @Failure 1405 {object} lib.APIError "user id could not be parsed"
// @Failure 1331 {object} lib.APIError "must be self"
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /users/{userID}/photo [put]
func (t *Controller) UploadProfilePhoto(c *gin.Context) {
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

	formFile, err := c.FormFile("file")
	if err != nil {
		t.RespondError(c, err)
		return
	}

	file, err := formFile.Open()
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	imgScaled := imaging.Fill(img, 300, 300, imaging.Center, imaging.Lanczos)
	imgThumb := imaging.Fill(img, 50, 50, imaging.Center, imaging.Lanczos)

	_, _ = imgScaled, imgThumb
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
			return 0, lib.ErrorUserIDNoParse
		}
	}

	return userID, nil
}
