package clubTrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/gin-gonic/gin"
)

// ListBulletins godoc
// @Summary Adds clubs to database
// @Description Adds club to database
// @ID ?
// @Tags clubtrack
// @Accept  json
// @Produce  json
// @Param all query bool false "Get All Bulletins (no restriction on startDate, endDate)"
// @Success 200 {array} models.Clubtrack
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /clubtrak [post]

// Adds club to database
func (t *Controller) AddClub(c *gin.Context) {
	// Bind create params
	var createData models.ClubTrak
	err := c.ShouldBind(&createData)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	//Create a copy of the schema with relevent data for a given club
	club := models.ClubTrak{
		//Need to figure out how to set club name in struct!!
		Description: "Test Test",
	}

	//Call create function to update database
	err = t.clubModel.CreateClub(&club)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	t.RespondCreated(c, club)

}
