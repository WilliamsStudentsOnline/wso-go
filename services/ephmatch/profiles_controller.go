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
// @Param preload query []string false "Preload List [tags, liked, matched]"
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

	/* Get total count */

	count, err := t.profileModel.CountProfiles()
	if err != nil {
		t.RespondError(c, err)
		return
	}
	t.SetPaginationTotal(c, count)

	/* Sanitize Users */
	for _, profile := range profiles {
		// Remove match message if did not match (or did not preload for matched)
		if profile.Matched == nil || !*profile.Matched {
			profile.MatchMessage = nil
		}
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
	likeExists := false
	matchExists := false

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

		// This is liked, not matching
		likeExists, err = t.likeModel.DoesLikeExist(userID, profileUserID)
		if err != nil {
			t.RespondError(c, err)
			return
		}

		// This is liked, not matching
		matchExists, err = t.matchModel.IsMatching(userID, profileUserID)
		if err != nil {
			t.RespondError(c, err)
			return
		}
	}

	profile.Liked = &likeExists
	profile.Matched = &matchExists

	// Remove match message if not matched
	if !matchExists {
		profile.MatchMessage = nil
	}

	// Sanitize user preloaded
	sanitize.User(profile.User, c)

	t.RespondOK(c, profile)
}

type LikeProfileResp struct {
	Matched bool `json:"matched"`
}

// LikeProfile godoc
// @Summary Like Ephmatch profile
// @Description Likes one ephmatch-eligible user profile
// @ID ephmatch-like-profile
// @Tags ephmatch
// @Accept  json
// @Produce  json
// @Param profileUserID path uint true "Profile User ID"
// @Success 201 {object} LikeProfileResp
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

	dupe, err := t.likeModel.DoesLikeExist(userID, profileUserID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if dupe {
		t.RespondError(c, lib.ErrorEphmatchAlreadyExists)
		return
	}

	// Do database query
	matched, err := t.ephmatchModel.CreateLikeAndMatch(userID, profileUserID)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondCreated(c, LikeProfileResp{
		Matched: matched,
	})
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

	dupe, err := t.likeModel.DoesLikeExist(userID, profileUserID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if !dupe {
		t.RespondError(c, lib.ErrorEphmatchDoesNotExist)
		return
	}

	// Do database query (delete like, match if exists)
	err = t.ephmatchModel.DeleteLikeAndMatch(userID, profileUserID)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, nil)
}
