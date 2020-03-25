package ephmatch

import (
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
// @Failure 500 {object} lib.APIError
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
		sanitize.User(match.MatchedUser, c)
	}

	t.RespondOK(c, responses.ConvertListMatchesResponse(matches))
}
