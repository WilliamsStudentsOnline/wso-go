package clubtrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/gin-gonic/gin"
)

type ClubParams struct {
	Name               string `json:"name"`
	Subscribers        int    `json:"Subscriber"`
	MeetingDescription string `json:"meetingDescription"`

	// Belongs to user Some Club Leader
	ClubID      uint         `json:"clubID"`
	ClubAdminID *models.User `json:"club,omitempty"`
}

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
func (t *Controller) CreateClub(c *gin.Context) {
	// Bind create params
	createData := ClubParams{}
	err := c.ShouldBind(&createData)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}
	//TODO: Discuss User with Charlie
	p1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "Student 1",
		UnixID: "s1",
	}

	//Create a copy of the schema with relevent data for a given club
	club := models.Club{
		Subscribers:        createData.Subscribers,
		MeetingDescription: createData.MeetingDescription,
		ClubAdmin:          createData.ClubID,
		ClubAdminID:        &p1,
		Name:               createData.Name,
	}

	//Call create function to update database
	err = t.clubModel.CreateClub(&club)
	if err != nil {
		t.RespondError(c, err)
		return
	}
	t.RespondCreated(c, club)

}
