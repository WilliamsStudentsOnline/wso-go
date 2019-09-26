package ephmatch

import (
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

// ListEphmatchers godoc
// @Summary List ephmatchers
// @Description lists all Ephmatch-eligible students
// @ID ephmatch-list-ephmatchers
// @Tags ephmatch
// @Accept  json
// @Produce  json
// @Param offset query int false "Offset Pagination"
// @Param limit query int false "Limit Pagination"
// @Param preload query []string false "Preload List"
// @Param q query string false "Search Query"
// @Success 200 {array} models.Ephmatcher
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /ephmatch/ephmatchers [get]
func (t *Controller) ListEphmatchers(c *gin.Context) {
	userID := services.GetUserID(c)

	var ephmatchers []*models.Ephmatcher
	var err error

	opts := models.GetAllEphmatchersOptions{}
	if err = c.ShouldBindQuery(&opts); err != nil {
		t.RespondError(c, err)
		return
	}

	// We could implement search here as well...
	err = t.ephmatcherModel.GetAllEphmatchers(userID, &ephmatchers, &opts)

	if err != nil {
		t.RespondError(c, err)
		return
	}

	count, err := t.ephmatcherModel.CountEphmatchers()
	if err != nil {
		t.RespondError(c, err)
		return
	}
	t.SetPaginationTotal(c, count)

	t.RespondOK(c, ephmatchers)
}

// GetEphmatcher godoc
// @Summary Get ephmatcher
// @Description gets one ephmatch-eligible user
// @ID ephmatch-get-ephmatcher
// @Tags ephmatch
// @Accept  json
// @Produce  json
// @Param ephmatcherID path uint true "Ephmatcher ID"
// @Success 200 {object} models.Ephmatcher
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /ephmatch/ephmatchers/{ephmatcherID} [get]
func (t *Controller) GetEphmatcher(c *gin.Context) {
	userID := services.GetUserID(c)

	// Decode ephmatcherID.
	ephmatcherID, err := services.GetUIntParam(c, "ephmatcherID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Do database query
	var ephmatcher models.Ephmatcher
	err = t.ephmatcherModel.GetEphmatcherByID(ephmatcherID, userID, &ephmatcher)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, ephmatcher)
}

// LikeEphmatcher godoc
// @Summary Like ephmatcher
// @Description Likes one ephmatch-eligible user
// @ID ephmatch-like-ephmatcher
// @Tags ephmatch
// @Accept  json
// @Produce  json
// @Param ephmatcherID path uint true "Ephmatcher ID"
// @Success 201
// @Failure 1730 {object} lib.APIError "cannot ephmatch-like yourself"
// @Failure 1731 {object} lib.APIError "ephmatcher could not be found"
// @Failure 1732 {object} lib.APIError "ephmatch already exists with user ID and passed ephmatcher ID"
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /ephmatch/ephmatchers/{ephmatcherID}/like [post]
func (t *Controller) LikeEphmatcher(c *gin.Context) {
	userID := services.GetUserID(c)

	// Decode ephmatcherID.
	ephmatcherID, err := services.GetUIntParam(c, "ephmatcherID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	if ephmatcherID == userID {
		t.RespondError(c, lib.ErrorEphmatchLikeNoSelf)
		return
	}

	exists, err := t.ephmatcherModel.DoesEphmatcherExist(ephmatcherID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if !exists {
		t.RespondError(c, lib.ErrorEphmatcherNotFound)
		return
	}

	dupe, err := t.ephmatchModel.CheckDuplicateEphmatch(userID, ephmatcherID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if dupe {
		t.RespondError(c, lib.ErrorEphmatchAlreadyExists)
		return
	}

	// Do database query
	err = t.ephmatchModel.CreateEphmatchWithUserOther(userID, ephmatcherID)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondCreated(c, nil)
}

// UnlikeEphmatcher godoc
// @Summary Unlike ephmatcher
// @Description Removes a like from one ephmatch-eligible user
// @ID ephmatch-unlike-ephmatcher
// @Tags ephmatch
// @Accept  json
// @Produce  json
// @Param ephmatcherID path uint true "Ephmatcher ID"
// @Success 201
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /ephmatch/ephmatchers/{ephmatcherID}/unlike [post]
func (t *Controller) UnlikeEphmatcher(c *gin.Context) {
	userID := services.GetUserID(c)

	// Decode ephmatcherID.
	ephmatcherID, err := services.GetUIntParam(c, "ephmatcherID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	exists, err := t.ephmatcherModel.DoesEphmatcherExist(ephmatcherID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if !exists {
		t.RespondError(c, lib.ErrorEphmatcherNotFound)
		return
	}

	dupe, err := t.ephmatchModel.CheckDuplicateEphmatch(userID, ephmatcherID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if !dupe {
		t.RespondError(c, lib.ErrorEphmatchDoesNotExist)
		return
	}

	// Do database query
	err = t.ephmatchModel.DeleteEphmatchWithUserOther(userID, ephmatcherID)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, nil)
}
