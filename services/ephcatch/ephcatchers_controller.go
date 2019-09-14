package ephcatch

import (
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

// ListEphcatchers godoc
// @Summary List ephcatchers
// @Description lists all Ephcatch-eligible students
// @ID ephcatch-list-ephcatchers
// @Tags ephcatch
// @Accept  json
// @Produce  json
// @Param offset query int false "Offset Pagination"
// @Param limit query int false "Limit Pagination"
// @Param preload query []string false "Preload List"
// @Param q query string false "Search Query"
// @Success 200 {array} models.Ephcatcher
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /ephcatch/ephcatchers [get]
func (t *Controller) ListEphcatchers(c *gin.Context) {
	userID := services.GetUserID(c)

	var ephcatchers []*models.Ephcatcher
	var err error

	opts := models.GetAllEphcatchersOptions{}
	if err = c.ShouldBindQuery(&opts); err != nil {
		t.RespondError(c, err)
		return
	}

	// We could implement search here as well...
	err = t.ephcatcherModel.GetAllEphcatchers(userID, &ephcatchers, &opts)

	if err != nil {
		t.RespondError(c, err)
		return
	}

	count, err := t.ephcatcherModel.CountEphcatchers()
	if err != nil {
		t.RespondError(c, err)
		return
	}
	t.SetPaginationTotal(c, count)

	t.RespondOK(c, ephcatchers)
}

// GetEphcatcher godoc
// @Summary Get ephcatcher
// @Description gets one ephcatch-eligible user
// @ID ephcatch-get-ephcatcher
// @Tags ephcatch
// @Accept  json
// @Produce  json
// @Param ephcatcherID path uint true "Ephcatcher ID"
// @Success 200 {object} models.Ephcatcher
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /ephcatch/ephcatchers/{ephcatcherID} [get]
func (t *Controller) GetEphcatcher(c *gin.Context) {
	userID := services.GetUserID(c)

	// Decode ephcatcherID.
	ephcatcherID, err := services.GetUIntParam(c, "ephcatcherID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Do database query
	var ephcatcher models.Ephcatcher
	err = t.ephcatcherModel.GetEphcatcherByID(ephcatcherID, userID, &ephcatcher)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, ephcatcher)
}

// LikeEphcatcher godoc
// @Summary Like ephcatcher
// @Description Likes one ephcatch-eligible user
// @ID ephcatch-like-ephcatcher
// @Tags ephcatch
// @Accept  json
// @Produce  json
// @Param ephcatcherID path uint true "Ephcatcher ID"
// @Success 201
// @Failure 1730 {object} lib.APIError "cannot ephcatch-like yourself"
// @Failure 1731 {object} lib.APIError "ephcatcher could not be found"
// @Failure 1732 {object} lib.APIError "ephcatch already exists with user ID and passed ephcatcher ID"
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /ephcatch/ephcatchers/{ephcatcherID}/like [post]
func (t *Controller) LikeEphcatcher(c *gin.Context) {
	userID := services.GetUserID(c)

	// Decode ephcatcherID.
	ephcatcherID, err := services.GetUIntParam(c, "ephcatcherID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	if ephcatcherID == userID {
		t.RespondError(c, lib.ErrorEphcatchLikeNoSelf)
		return
	}

	exists, err := t.ephcatcherModel.DoesEphcatcherExist(ephcatcherID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if !exists {
		t.RespondError(c, lib.ErrorEphcatcherNotFound)
		return
	}

	dupe, err := t.ephcatchModel.CheckDuplicateEphcatch(userID, ephcatcherID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if dupe {
		t.RespondError(c, lib.ErrorEphcatchAlreadyExists)
		return
	}

	// Do database query
	err = t.ephcatchModel.CreateEphcatchWithUserOther(userID, ephcatcherID)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondCreated(c, nil)
}

// UnlikeEphcatcher godoc
// @Summary Unlike ephcatcher
// @Description Removes a like from one ephcatch-eligible user
// @ID ephcatch-unlike-ephcatcher
// @Tags ephcatch
// @Accept  json
// @Produce  json
// @Param ephcatcherID path uint true "Ephcatcher ID"
// @Success 201
// @Failure 400 {object} lib.APIError
// @Failure 404 {object} lib.APIError
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /ephcatch/ephcatchers/{ephcatcherID}/unlike [post]
func (t *Controller) UnlikeEphcatcher(c *gin.Context) {
	userID := services.GetUserID(c)

	// Decode ephcatcherID.
	ephcatcherID, err := services.GetUIntParam(c, "ephcatcherID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	exists, err := t.ephcatcherModel.DoesEphcatcherExist(ephcatcherID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if !exists {
		t.RespondError(c, lib.ErrorEphcatcherNotFound)
		return
	}

	dupe, err := t.ephcatchModel.CheckDuplicateEphcatch(userID, ephcatcherID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if !dupe {
		t.RespondError(c, lib.ErrorEphcatchDoesNotExist)
		return
	}

	// Do database query
	err = t.ephcatchModel.DeleteEphcatchWithUserOther(userID, ephcatcherID)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, nil)
}
