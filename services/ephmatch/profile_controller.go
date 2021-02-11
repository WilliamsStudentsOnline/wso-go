package ephmatch

import (
	"image"
	"net/http"
	"os"

	"unicode/utf8"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/sanitize"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
)

// GetSelfProfile godoc
// @Summary Get Ephmatch profile of self
// @Description gets self ephmatch profile even if deleted
// @ID ephmatch-get-self-profile
// @Tags ephmatch
// @Accept  json
// @Produce  json
// @Success 200 {object} models.EphmatchProfile
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /ephmatch/profile [get]
func (t *Controller) GetSelfProfile(c *gin.Context) {
	userID := services.GetUserID(c)

	var profile models.EphmatchProfile

	// Do database query
	err := t.profileModel.GetSelfProfileByID(userID, &profile)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Sanitize user preloaded
	sanitize.User(profile.User, c)

	t.RespondOK(c, profile)
}

type ProfileCreateParams struct {
	Description       *string `json:"description"`
	MatchMessage      *string `json:"matchMessage"`
	LocationVisible   *bool   `json:"locationVisible"`
	LocationTown      *string `json:"locationTown"`
	LocationState     *string `json:"locationState"`
	LocationCountry   *string `json:"LocationCountry"`
	MessagingPlatform *string `json:"messagingPlatform"`
	MessagingUsername *string `json:"messagingUsername"`
}

// CreateProfile godoc
// @Summary Create an Ephmatch profile
// @Description creates self ephmatch profile
// @ID ephmatch-create-profile
// @Tags ephmatch
// @Accept  json
// @Produce  json
// @Param createParams body ephmatch.ProfileCreateParams true "Create Profile Params"
// @Success 201 {object} models.EphmatchProfile
// @Failure 1101 {object} lib.APIError "request data validation failed"
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /ephmatch/profile [post]
func (t *Controller) CreateProfile(c *gin.Context) {
	userID := services.GetUserID(c)

	// Bind create params
	createData := ProfileCreateParams{}
	err := c.ShouldBind(&createData)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Can have no platform and no username
	if createData.MessagingPlatform != nil && (*createData.MessagingPlatform == "NONE" || *createData.MessagingPlatform == "") {
		createData.MessagingPlatform = lib.StrToPtr("")
		createData.MessagingUsername = lib.StrToPtr("")
	}

	if createData.MessagingPlatform != nil && *createData.MessagingPlatform != "" {
		// Must have valid platform or no platform (NONE)
		if !models.ValidateEphmatchMessagingPlatform(*createData.MessagingPlatform) {
			t.RespondError(c, lib.ErrorEphmatchInvalidMessagingPlatform)
			return
		}

		// Cannot have a valid platform and no username
		if createData.MessagingUsername == nil {
			t.RespondError(c, lib.ErrorEphmatchEmptyMessagingUsername)
			return
		}
	}

	if createData.Description != nil && utf8.RuneCountInString(*createData.Description) >= 255 {
		t.RespondError(c, lib.ErrorEphmatchDescriptionTooLong)
		return
	}

	newProfile := models.EphmatchProfile{
		Description:       createData.Description,
		MatchMessage:      createData.MatchMessage,
		UserID:            userID,
		LocationVisible:   createData.LocationVisible,
		LocationTown:      createData.LocationTown,
		LocationState:     createData.LocationState,
		LocationCountry:   createData.LocationCountry,
		MessagingPlatform: createData.MessagingPlatform,
		MessagingUsername: createData.MessagingUsername,
	}

	var profile models.EphmatchProfile

	err = t.profileModel.CreateOrUpdateProfileUnscoped(userID, newProfile, &profile)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Update when created
	t.SetUpdateToken(c)

	// Sanitize user preloaded
	sanitize.User(profile.User, c)

	t.RespondCreated(c, profile)
}

type ProfileUpdateParams struct {
	Description       *string `json:"description"`
	MatchMessage      *string `json:"matchMessage"`
	LocationVisible   *bool   `json:"locationVisible"`
	LocationTown      *string `json:"locationTown"`
	LocationState     *string `json:"locationState"`
	LocationCountry   *string `json:"locationCountry"`
	MessagingPlatform *string `json:"messagingPlatform"`
	MessagingUsername *string `json:"messagingUsername"`
}

// UpdateProfile godoc
// @Summary Update an Ephmatch profile
// @Description updates self's ephmatch profile. Profile must be created
// @ID ephmatch-update-profile
// @Tags ephmatch
// @Accept  json
// @Produce  json
// @Param updateParams body ephmatch.ProfileUpdateParams true "Update Profile Params"
// @Success 200 {object} models.EphmatchProfile
// @Failure 1101 {object} lib.APIError "request data validation failed"
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /ephmatch/profile [patch]
func (t *Controller) UpdateProfile(c *gin.Context) {
	userID := services.GetUserID(c)

	// Bind update params
	updateData := ProfileUpdateParams{}
	err := c.ShouldBind(&updateData)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Do database query
	var profile models.EphmatchProfile
	err = t.profileModel.GetSelfProfileByIDScopedNoDefault(userID, &profile)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	profile.Description = lib.StrPtrDefaults(updateData.Description, profile.Description)
	profile.MatchMessage = lib.StrPtrDefaults(updateData.MatchMessage, profile.MatchMessage)
	profile.LocationVisible = lib.BoolPtrDefaults(updateData.LocationVisible, profile.LocationVisible)
	profile.LocationTown = lib.StrPtrDefaults(updateData.LocationTown, profile.LocationTown)
	profile.LocationState = lib.StrPtrDefaults(updateData.LocationState, profile.LocationState)
	profile.LocationCountry = lib.StrPtrDefaults(updateData.LocationCountry, profile.LocationCountry)
	profile.MessagingPlatform = lib.StrPtrDefaults(updateData.MessagingPlatform, profile.MessagingPlatform)
	profile.MessagingUsername = lib.StrPtrDefaults(updateData.MessagingUsername, profile.MessagingUsername)

	// Can have no platform and no username
	if profile.MessagingPlatform != nil && (*profile.MessagingPlatform == "NONE" || *profile.MessagingPlatform == "") {
		profile.MessagingPlatform = lib.StrToPtr("")
		profile.MessagingUsername = lib.StrToPtr("")
	}

	if profile.MessagingPlatform != nil && *profile.MessagingPlatform != "" {
		// Must have valid platform or no platform (NONE)
		if !models.ValidateEphmatchMessagingPlatform(*profile.MessagingPlatform) {
			t.RespondError(c, lib.ErrorEphmatchInvalidMessagingPlatform)
			return
		}

		// Cannot have a valid platform and no username
		if profile.MessagingUsername == nil || *profile.MessagingUsername == "" {
			t.RespondError(c, lib.ErrorEphmatchEmptyMessagingUsername)
			return
		}
	}

	if updateData.Description != nil && utf8.RuneCountInString(*updateData.Description) >= 255 {
		t.RespondError(c, lib.ErrorEphmatchDescriptionTooLong)
		return
	}

	err = t.profileModel.UpdateProfile(&profile)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Sanitize user preloaded
	sanitize.User(profile.User, c)

	t.RespondOK(c, profile)
}

// DeleteProfile godoc
// @Summary Delete profile
// @Description delete self's ephmatch profile
// @ID ephmatch-delete-profile
// @Tags ephmatch
// @Accept  json
// @Produce  json
// @Success 200 {object} models.EphmatchProfile
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /ephmatch/profile [delete]
func (t *Controller) DeleteProfile(c *gin.Context) {
	userID := services.GetUserID(c)

	// Do database query
	var profile models.EphmatchProfile
	err := t.profileModel.GetSelfProfileByIDScopedNoDefault(userID, &profile)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Do db delete
	err = t.profileModel.DeleteProfile(&profile)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Sanitize user preloaded
	sanitize.User(profile.User, c)

	// Update when deleted
	t.SetUpdateToken(c)

	t.RespondOK(c, profile)
}

// UploadEphmatchProfilePhoto godoc
// @Summary Upload an ephmatch profile photo by user id
// @Description upload an ephmatch user's ephmatch profile photo by user id.
// @ID upload-ephmatch-profile-photo
// @Tags users
// @Accept  multipart/form-data
// @Produce  json
// @Param file formData file true "Profile Photo"
// @Success 200
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /ephmatch/profile/photo [put]
func (t *Controller) UploadEphmatchProfilePhoto(c *gin.Context) {
	userID := services.GetUserID(c)

	// Do database query
	var profile models.EphmatchProfile
	err := t.profileModel.GetSelfProfileByIDScopedNoDefault(userID, &profile)
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

	imgScaled := imaging.Fill(img, 300, 300, imaging.Center, imaging.Lanczos)
	err = t.pictureBackend.SaveEphmatchPhoto(profile.User.UnixID, imgScaled)
	if err != nil {
		c.Error(err)
		t.RespondError(c, lib.ErrorUnableToSavePicture)
		return
	}

	t.RespondOK(c, nil)
}

// DeleteEphmatchProfilePhoto godoc
// @Summary Delete an ephmatch profile photo by user id
// @Description delete an ephmatch user's ephmatch profile photo by user id.
// @ID delete-ephmatch-profile-photo
// @Tags users
// @Produce  json
// @Success 200
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /ephmatch/profile/photo [delete]
func (t *Controller) DeleteEphmatchProfilePhoto(c *gin.Context) {
	userID := services.GetUserID(c)

	// Do database query
	var profile models.EphmatchProfile
	err := t.profileModel.GetSelfProfileByIDScopedNoDefault(userID, &profile)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	err = t.pictureBackend.DeleteEphmatchPhoto(profile.User.UnixID)
	if err != nil && os.IsNotExist(err) {
		t.RespondError(c, lib.ErrorRecordNotFound)
		return
	} else if err != nil {
		c.Error(err)
		t.RespondError(c, lib.ErrorUnableToDeletePicture)
		return
	}

	t.RespondOK(c, nil)
}
