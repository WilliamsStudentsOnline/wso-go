package ephcatch

import (
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/sanitize"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

// ListMatches godoc
// @Summary List matches
// @Description lists all Ephcatch-eligible students that user has matched with
// @ID listEphcatchMatches
// @Tags ephcatch
// @Accept  json
// @Produce  json
// @Success 200 {array} models.User
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /ephcatch/matches [get]
func (t *Controller) ListMatches(c *gin.Context) {
	userID := services.GetUserID(c)

	var matches []*models.User
	var err error

	// We could implement search here as well...
	err = t.ephcatcherModel.GetMatches(userID, &matches)

	if err != nil {
		t.RespondError(c, err)
		return
	}

	sanitize.Users(matches, c)

	t.RespondOK(c, matches)
}
