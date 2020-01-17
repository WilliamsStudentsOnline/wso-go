package ephmatch

import (
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/sanitize"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

// ListProfiles godoc
// @Summary List Ephmatch profiles
// @Description lists all Ephmatch-eligible student profiles
// @ID ephmatch-list-profiles
// @Tags ephmatch
// @Accept  json
// @Produce  json
// @Param sort query string false "Sort (new, updated, alphabetical)"
// @Param offset query int false "Offset Pagination"
// @Param limit query int false "Limit Pagination"
// @Param preload query []string false "Preload List"
// @Success 200 {array} models.EphmatchProfile
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /ephmatch/profiles [get]
func (t *Controller) ListProfiles(c *gin.Context) {
	userID := services.GetUserID(c)

	var profiles []*models.EphmatchProfile
	var err error

	opts := models.GetAllProfilesOptions{}
	if err = c.ShouldBindQuery(&opts); err != nil {
		t.RespondError(c, err)
		return
	}

	// We could implement search here as well...
	err = t.profileModel.GetAllProfilesNoSelf(&profiles, userID, &opts)

	if err != nil {
		t.RespondError(c, err)
		return
	}

	/* Populate liked field */

	// Get ephmatches made by user self (userID) to populate "liked" field. This is significantly faster
	// than a SQL query by several magnitudes.
	var ephmatches []*models.Ephmatch
	err = t.ephmatchModel.GetUserLikes(userID, &ephmatches)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Create a map of users we liked.
	likedUserMap := make(map[uint]bool)
	for _, match := range ephmatches {
		likedUserMap[match.OtherID] = true
	}

	// If a profile is in the likedUserMap, set it to liked
	for _, profile := range profiles {
		if _, ok := likedUserMap[profile.UserID]; ok {
			profile.Liked = true
		}
	}

	/* Get total count */

	count, err := t.profileModel.CountProfiles()
	if err != nil {
		t.RespondError(c, err)
		return
	}
	t.SetPaginationTotal(c, count)

	/* Sanitize Users */
	for _, profile := range profiles {
		sanitize.User(profile.User, c)
	}

	t.RespondOK(c, profiles)
}

// GetProfile godoc
// @Summary Get Ephmatch profile
// @Description gets one ephmatch-eligible user profile
// @ID ephmatch-get-profile
// @Tags ephmatch
// @Accept  json
// @Produce  json
// @Param profileUserID path uint true "Profile User ID"
// @Success 200 {object} models.EphmatchProfile
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /ephmatch/profiles/{profileUserID} [get]
func (t *Controller) GetProfile(c *gin.Context) {
	userID := services.GetUserID(c)

	// Decode profileUserID.
	profileUserID, err := services.GetUIntParam(c, "profileUserID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	var profile models.EphmatchProfile
	isMatching := false

	// If get self, get it regardless. Otherwise, user must be valid
	if profileUserID == userID {
		// Do database query
		err = t.profileModel.GetSelfProfileByID(userID, &profile)
		if err != nil {
			t.RespondError(c, err)
			return
		}
	} else {
		// Do database query
		err = t.profileModel.GetProfileByID(profileUserID, &profile)
		if err != nil {
			t.RespondError(c, err)
			return
		}

		isMatching, err = t.ephmatchModel.IsMatching(userID, profileUserID)
		if err != nil {
			t.RespondError(c, err)
			return
		}
	}

	profile.Liked = isMatching

	// Sanitize user preloaded
	sanitize.User(profile.User, c)

	t.RespondOK(c, profile)
}

// LikeProfile godoc
// @Summary Like Ephmatch profile
// @Description Likes one ephmatch-eligible user profile
// @ID ephmatch-like-profile
// @Tags ephmatch
// @Accept  json
// @Produce  json
// @Param profileUserID path uint true "Profile User ID"
// @Success 201
// @Failure 1730 {object} lib.APIError "cannot ephmatch-like yourself"
// @Failure 1731 {object} lib.APIError "ephmatch profile could not be found"
// @Failure 1732 {object} lib.APIError "ephmatch already exists with user ID and passed ephmatch profile user ID"
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /ephmatch/profiles/{profileUserID}/like [post]
func (t *Controller) LikeProfile(c *gin.Context) {
	userID := services.GetUserID(c)

	// Decode profileUserID.
	profileUserID, err := services.GetUIntParam(c, "profileUserID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	if profileUserID == userID {
		t.RespondError(c, lib.ErrorEphmatchLikeNoSelf)
		return
	}

	exists, err := t.profileModel.DoesProfileExist(profileUserID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if !exists {
		t.RespondError(c, lib.ErrorEphmatchProfileNotFound)
		return
	}

	dupe, err := t.ephmatchModel.CheckDuplicateEphmatch(userID, profileUserID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if dupe {
		t.RespondError(c, lib.ErrorEphmatchAlreadyExists)
		return
	}

	// Do database query
	err = t.ephmatchModel.CreateEphmatchWithUserOther(userID, profileUserID)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondCreated(c, nil)
}

// UnlikeProfile godoc
// @Summary Unlike ephmatch profile
// @Description Removes a like from one ephmatch-eligible user profile
// @ID ephmatch-unlike-profile
// @Tags ephmatch
// @Accept  json
// @Produce  json
// @Param profileUserID path uint true "Profile User ID"
// @Success 201
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /ephmatch/profiles/{profileUserID}/unlike [post]
func (t *Controller) UnlikeProfile(c *gin.Context) {
	userID := services.GetUserID(c)

	// Decode profileUserID.
	profileUserID, err := services.GetUIntParam(c, "profileUserID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	exists, err := t.profileModel.DoesProfileExist(profileUserID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if !exists {
		t.RespondError(c, lib.ErrorEphmatchProfileNotFound)
		return
	}

	dupe, err := t.ephmatchModel.CheckDuplicateEphmatch(userID, profileUserID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if !dupe {
		t.RespondError(c, lib.ErrorEphmatchDoesNotExist)
		return
	}

	// Do database query
	err = t.ephmatchModel.DeleteEphmatchWithUserOther(userID, profileUserID)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, nil)
}
