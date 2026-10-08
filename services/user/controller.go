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
	"github.com/WilliamsStudentsOnline/wso-go/lib/pictures"
	searchLib "github.com/WilliamsStudentsOnline/wso-go/lib/search"
	search "github.com/WilliamsStudentsOnline/wso-go/lib/search/users"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/sanitize"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/WilliamsStudentsOnline/wso-go/services/user/responses"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type Controller struct {
	services.BaseController
	userModel      *models.UserModel
	userSearch     search.SearchUsers
	pictureBackend pictures.PictureBackend
}

// Construct a new user controller
func NewController(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) *Controller {
	pb, err := pictures.NewPictureBackend(cfg, log)
	if err != nil {
		log.Error(err)
		log.Warn("Using picture backend none")
		pb = pictures.NewPictureBackendDummy()
	}

	return &Controller{
		BaseController: services.BaseController{Log: log},
		userModel:      models.NewUserModel(db, log),
		userSearch:     search.NewSearchUsers(db, cfg, log),
		pictureBackend: pb,
	}
}

//go:generate go run github.com/WilliamsStudentsOnline/wso-go/lib/generate/service_responses/cmd -in responses/list_users.json -out responses/list_users.go

// ListUsers godoc
// @Summary List users
// @Description Get all users that are visible and at williams.
// @Description If you pass a search query (?q="blah"), you will get all users matching that search query
// @ID listUsers
// @Tags users
// @Accept  json
// @Produce  json
// @Param offset query int false "Offset Pagination"
// @Param limit query int false "Limit Pagination"
// @Param preload query []string false "Preload List"
// @Param q query string false "Search Query"
// @Success 200 {array} responses.ListUsersResponseUser
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /users [get]
func (t *Controller) ListUsers(c *gin.Context) {
	var users []*models.User
	var totalResults int
	var err error

	opts := models.GetAllUsersOptions{}
	if err = c.ShouldBindQuery(&opts); err != nil {
		t.RespondBadBind(c, err)
		return
	}

	if query, ok := c.GetQuery("q"); ok {
		users, totalResults, err = t.userSearch.Search(query, &search.SearchUsersOptionsMysql{GetAllUsersOptions: &opts})

		if err != nil {
			if searchLib.IsInvalidTokenError(err) {
				err = lib.NewErrorInvalidSearchToken(err)
			} else if searchLib.IsQueryError(err) {
				err = lib.NewErrorInvalidSearchQuery(err)
			}
			t.RespondError(c, err)
			return
		}
	} else {
		err = t.userModel.GetAllUsers(&users, &opts)

		if err != nil {
			t.RespondError(c, err)
			return
		}

		totalResults, err = t.userModel.CountAllUsers()
		if err != nil {
			t.RespondError(c, err)
			return
		}
	}

	// Secure home and dorm if not admin
	if !auth.HasScope(c, auth.ScopeAdminAll) {
		for i := range users {
			if !*users[i].HomeVisible {
				users[i].HomeTown = nil
				users[i].HomeState = nil
				users[i].HomeZip = nil
				users[i].HomeCountry = nil
			}

			if !*users[i].DormVisible {
				users[i].DormRoom = nil
				users[i].DormRoomID = nil
			}
		}
	}

	t.SetPaginationTotal(c, totalResults)

	sanitize.Users(users, c)

	t.RespondOK(c, responses.ConvertListUsersResponse(users))
}

//go:generate go run github.com/WilliamsStudentsOnline/wso-go/lib/generate/service_responses/cmd -in responses/get_user.json -out responses/get_user.go

// Get user by id. Pass "me" if you want to get self
// GetUser godoc
// @Summary Get user by user id
// @Description get a user by user id that is visible and at williams. Also loads tags. Pass "me" if you want to get self
// @ID getUser
// @Tags users
// @Accept  json
// @Produce  json
// @Param userID path uint true "User ID"
// @Success 200 {object} responses.GetUserResponseUser
// @Failure 1403 {object} services.BaseErrorResponse "user not visible"
// @Failure 1404 {object} services.BaseErrorResponse "user not at williams"
// @Failure 1405 {object} services.BaseErrorResponse "user id could not be parsed"
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
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

		// Secure home and dorm
		if !*user.HomeVisible {
			user.HomeTown = nil
			user.HomeState = nil
			user.HomeZip = nil
			user.HomeCountry = nil
		}

		if !*user.DormVisible {
			user.DormRoom = nil
			user.DormRoomID = nil
		}
	}

	sanitize.User(&user, c)

	t.RespondOK(c, responses.ConvertGetUserResponse(&user))
}

type UpdateUserParams struct {
	Visible                   *bool   `json:"visible"`
	DormVisible               *bool   `json:"dormVisible"`
	HomeVisible               *bool   `json:"homeVisible"`
	OffCycle                  *bool   `json:"offCycle"`
	HasAcceptedFactrakPolicy  *bool   `json:"hasAcceptedFactrakPolicy"`
	HasAcceptedDormtrakPolicy *bool   `json:"hasAcceptedDormtrakPolicy"`
	Nickname                  *string `json:"nickname"`
	OptOutEphcatch            *bool   `json:"optOutEphcatch"`
	CampusStatus              *string `json:"campusStatus"`
}

// UpdateUser godoc
// @Summary Update user by user id
// @Description updates a user by user id. You may only update yourself. You may pass "me" to get self as well.
// @ID updateUser
// @Tags users
// @Accept  json
// @Produce  json
// @Param userID path uint true "User ID"
// @Param updateParams body user.UpdateUserParams true "Update User Parameters"
// @Success 200 {object} models.User
// @Failure 1405 {object} services.BaseErrorResponse "user id could not be parsed"
// @Failure 1331 {object} services.BaseErrorResponse "must be self"
// @Failure 1100 {object} services.BaseErrorResponse "could not parse malformed request data"
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
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
		t.RespondBadBind(c, lib.ErrorMalformedRequestData)
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
	user.OffCycle = lib.BoolPtrDefaults(updateData.OffCycle, user.OffCycle)
	user.HasAcceptedFactrakPolicy = lib.BoolPtrDefaults(updateData.HasAcceptedFactrakPolicy, user.HasAcceptedFactrakPolicy)
	user.HasAcceptedDormtrakPolicy = lib.BoolPtrDefaults(updateData.HasAcceptedDormtrakPolicy, user.HasAcceptedDormtrakPolicy)
	user.Nickname = lib.StrPtrDefaults(updateData.Nickname, user.Nickname)
	user.OptOutEphcatch = lib.BoolPtrDefaults(updateData.OptOutEphcatch, user.OptOutEphcatch)
	user.CampusStatus = lib.StrPtrDefaults(updateData.CampusStatus, user.CampusStatus)

	// Error if bad campus status
	if user.CampusStatus != nil && *user.CampusStatus != "" && !models.ValidateCampusStatus(*user.CampusStatus) {
		t.RespondError(c, lib.ErrorUserInvalidCampusStatus)
		return
	}

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
// @ID updateUserTags
// @Tags users
// @Accept  json
// @Produce  json
// @Param userID path uint true "User ID"
// @Param updateTagsParams body user.UpdateUserTagsParams true "Update Tags Params"
// @Success 200
// @Failure 1405 {object} services.BaseErrorResponse "user id could not be parsed"
// @Failure 1331 {object} services.BaseErrorResponse "must be self"
// @Failure 1100 {object} services.BaseErrorResponse "could not parse malformed request data"
// @Failure 1406 {object} services.BaseErrorResponse "invalid user tag"
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
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
		t.RespondBadBind(c, lib.ErrorMalformedRequestData)
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
// @ID updateUserPhoto
// @Tags users
// @Accept  multipart/form-data
// @Produce  json
// @Param userID path uint true "User ID"
// @Param file formData file true "Profile Photo"
// @Success 200
// @Failure 1405 {object} services.BaseErrorResponse "user id could not be parsed"
// @Failure 1331 {object} services.BaseErrorResponse "must be self"
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
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

	// Do database query to get the user
	var user models.User
	err = t.userModel.GetUserByID(userID, &user)
	if err != nil {
		t.RespondError(c, err)
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

	err = t.pictureBackend.SaveUserPhotoBoth(user.UnixID, img)
	if err != nil {
		c.Error(err)
		t.RespondAPIError(c, lib.ErrorUnableToSavePicture)
		return
	}

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
			return 0, lib.ErrorUserIDNoParse
		}
	}

	return userID, nil
}
