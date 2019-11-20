package ephmatch

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/sanitize"
	"github.com/WilliamsStudentsOnline/wso-go/services"
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
	Description *string `json:"description" binding:"required"`
	Gender      *string `json:"gender" binding:"required"`
	OtherGender bool    `json:"otherGender"`
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
// @Failure 1940 {object} lib.APIError "unknown gender type specified; use other gender flag for custom gender"
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

	// If the gender is unknown (or other), we check if the other gender flag is set. If so, we use that. Otherwise, we error
	// This is mostly to ensure that the majority of people explicitly use our set gender strings, rather than custom
	// ones (hard to distinguish between "he/him" and "he/him/his" etc.)
	if models.EphmatchProfileGenderType(*createData.Gender) == models.EphmatchProfileGenderOther && !createData.OtherGender {
		t.RespondError(c, lib.ErrorEphmatchGenderUnknown)
		return
	}

	newProfile := models.EphmatchProfile{
		Gender:      *createData.Gender,
		Description: *createData.Description,
		UserID:      userID,
	}

	var profile models.EphmatchProfile

	err = t.profileModel.CreateOrUpdateProfileUnscoped(userID, newProfile, &profile)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Sanitize user preloaded
	sanitize.User(profile.User, c)

	t.RespondCreated(c, profile)
}

type ProfileUpdateParams struct {
	Description *string `json:"description"`
	Gender      *string `json:"gender"`
	OtherGender bool    `json:"otherGender"`
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
// @Failure 1940 {object} lib.APIError "unknown gender type specified; use other gender flag for custom gender"
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

	// If we set gender, and...
	// If the gender is unknown (or other), we check if the other gender flag is set. If so, we use that. Otherwise, we error
	// This is mostly to ensure that the majority of people explicitly use our set gender strings, rather than custom
	// ones (hard to distinguish between "he/him" and "he/him/his" etc.)
	if updateData.Gender != nil &&
		models.EphmatchProfileGenderType(*updateData.Gender) == models.EphmatchProfileGenderOther &&
		!updateData.OtherGender {

		t.RespondError(c, lib.ErrorEphmatchGenderUnknown)
		return
	}

	// Do database query
	var profile models.EphmatchProfile
	err = t.profileModel.GetSelfProfileByIDScopedNoDefault(userID, &profile)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	profile.Gender = *lib.StrPtrDefaults(updateData.Gender, &profile.Gender)
	profile.Description = *lib.StrPtrDefaults(updateData.Description, &profile.Description)

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

	t.RespondOK(c, profile)
}
