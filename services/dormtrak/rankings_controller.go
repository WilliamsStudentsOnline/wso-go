package dormtrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/models"
	_ "github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

// GetRankings godoc
// @Summary Get rankings
// @Description gets rankings of dorms by specific metrics.
// @ID dormtrak-get-rankings
// @Tags dormtrak
// @Accept  json
// @Produce  json
// @Success 200 {array} models.DormtrakRanking
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /dormtrak/rankings [get]
func (t *Controller) GetRankings(c *gin.Context) {
	// Do database query
	rankings := models.NewDormtrakRanking()
	err := t.dormModel.GetDormtrakRankings(3, rankings)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, rankings)
}
