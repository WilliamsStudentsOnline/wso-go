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
// @Param preload query []string false "Preload List [tags, relation, matched]"
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
			profile.MessagingPlatform = nil
			profile.MessagingUsername = nil
		}
		sanitize.EphmatchProfile(profile, c)
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
	relExists := false
	relVal := ""
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
		relExists, relVal, err = t.relationModel.GetOutRelation(userID, profileUserID)
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

	if relExists {
		profile.Relation = &relVal
	}
	profile.Matched = &matchExists

	// Remove match message if not matched
	if !matchExists {
		profile.MatchMessage = nil
		profile.MessagingPlatform = nil
		profile.MessagingUsername = nil
	}

	// Sanitize user preloaded
	sanitize.User(profile.User, c)
	sanitize.EphmatchProfile(&profile, c)

	t.RespondOK(c, profile)
}

type SetProfileRelationResp struct {
	Matched *bool `json:"matched,omitempty"`
}

type SetProfileRelationParams struct {
	Relation string `json:"relation"`
}

const (
	ProfileRelationNone = "none"
)

// SetProfileRelation godoc
// @Summary Set Ephmatch profile relation
// @Description Sets the relation (like, dislike, nothing) between self and one ephmatch-eligible user profile
// @ID ephmatch-set-profile-relation
// @Tags ephmatch
// @Accept  json
// @Produce  json
// @Param profileUserID path uint true "Profile User ID"
// @Param relationParams body ephmatch.SetProfileRelationParams true "Set Profile Relation Params"
// @Success 201 {object} SetProfileRelationResp
// @Failure 1730 {object} lib.APIError "cannot ephmatch-relate yourself"
// @Failure 1731 {object} lib.APIError "ephmatch profile could not be found"
// @Failure 1732 {object} lib.APIError "ephmatch relation already exists with user ID and passed ephmatch profile user ID"
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /ephmatch/profiles/{profileUserID}/relation [put]
func (t *Controller) SetProfileRelation(c *gin.Context) {
	userID := services.GetUserID(c)

	// Bind create params
	relParams := SetProfileRelationParams{}
	err := c.ShouldBind(&relParams)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	if !models.ValidateEphmatchRelation(relParams.Relation) && relParams.Relation != ProfileRelationNone {
		t.RespondError(c, lib.ErrorEphmatchInvalidRelation)
		return
	}

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

	if relParams.Relation == ProfileRelationNone {
		err = t.ephmatchModel.DeleteRelationWithMatchHooks(userID, profileUserID)
		if err != nil {
			t.RespondError(c, err)
			return
		}
		t.RespondOK(c, SetProfileRelationResp{})
	} else {
		// Do database query
		matched, err := t.ephmatchModel.SetRelationWithMatchHooks(userID, profileUserID, relParams.Relation)
		if err != nil {
			if err == models.EphmatchModelErrorRelationAlreadyExists {
				t.RespondError(c, lib.ErrorEphmatchRelationAlreadyExists)
				return
			} else {
				t.RespondError(c, err)
				return
			}
		}
		t.RespondOK(c, SetProfileRelationResp{
			Matched: &matched,
		})
	}
}
