package ephmatch

import (
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/sanitize"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/WilliamsStudentsOnline/wso-go/services/ephmatch/responses"
	"github.com/gin-gonic/gin"
)

//go:generate go run github.com/WilliamsStudentsOnline/wso-go/lib/generate/service_responses/cmd -in responses/list_matches.json -out responses/list_matches.go

// ListMatches godoc
// @Summary List matches
// @Description lists all Ephmatch-eligible students that user has matched with
// @ID ephmatch-list-matches
// @Tags ephmatch
// @Accept  json
// @Produce  json
// @Param preload query []string false "Preload List [tags]"
// @Success 200 {array} responses.ListMatchesResponseEphmatchMatch
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /ephmatch/matches [get]
func (t *Controller) ListMatches(c *gin.Context) {
	userID := services.GetUserID(c)

	var matches []*models.EphmatchMatch
	var err error

	opts := models.GetMatchesOptions{}
	if err = c.ShouldBindQuery(&opts); err != nil {
		t.RespondError(c, err)
		return
	}

	// We could implement search here as well...
	err = t.matchModel.GetMatches(userID, &opts, &matches)

	if err != nil {
		t.RespondError(c, err)
		return
	}

	for _, match := range matches {
		sanitize.EphmatchProfile(match.MatchedUser.EphmatchProfile, c)
		sanitize.User(match.MatchedUser, c)
	}

	t.RespondOK(c, responses.ConvertListMatchesResponse(matches))

	err = t.matchModel.SetMatchesAsSeen(userID)
	if err != nil {
		lg := t.Log.With(
			"userID", userID,
			"err", err)
		lg.Warn("error when setting matches as seen")
	}
}

type CountMatchesResponse struct {
	Total  int `json:"total"`
	Unseen int `json:"unseen"`
}

// CountMatches godoc
// @Summary Count matches
// @Description counts all Ephmatch-eligible students that user has matched with by unseen and total
// @ID ephmatch-count-matches
// @Tags ephmatch
// @Accept  json
// @Produce  json
// @Success 200 {object} ephmatch.CountMatchesResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /ephmatch/matches-count [get]
func (t *Controller) CountMatches(c *gin.Context) {
	userID := services.GetUserID(c)
	var err error

	// We could implement search here as well...
	unseen, total, err := t.matchModel.CountMatchesAndUnseen(userID)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Log these numbers
	ephmatchMatchesMatchSummaryTotal.Observe(float64(total))
	ephmatchMatchesMatchSummaryUnseen.Observe(float64(total))

	t.RespondOK(c, CountMatchesResponse{
		Unseen: unseen,
		Total:  total,
	})
}

// Unmatch godoc
// @Summary Unmatches Ephmatch matched users
// @Description Removes the match (and sets the to-relation to none) of a specific matched pair
// @ID ephmatch-unmatch
// @Tags ephmatch
// @Accept  json
// @Produce  json
// @Param matchUserID path uint true "Match User ID"
// @Success 200
// @Failure 1730 {object} services.BaseErrorResponse "cannot ephmatch-relate yourself"
// @Failure 1733 {object} services.BaseErrorResponse "ephmatch match could not be found"
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /ephmatch/matches/{matchUserID} [delete]
func (t *Controller) Unmatch(c *gin.Context) {
	userID := services.GetUserID(c)

	// Decode matchUserID.
	matchUserID, err := services.GetUIntParam(c, "matchUserID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	if matchUserID == userID {
		t.RespondError(c, lib.ErrorEphmatchLikeNoSelf)
		return
	}

	matching, err := t.matchModel.IsMatching(userID, matchUserID)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	if !matching {
		t.RespondError(c, lib.ErrorEphmatchDoesNotExist)
		return
	}

	err = t.ephmatchModel.DeleteRelationWithMatchHooks(userID, matchUserID)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Another one to the count
	ephmatchMatchesUnmatchCounter.Inc()

	t.RespondOK(c, nil)
}
