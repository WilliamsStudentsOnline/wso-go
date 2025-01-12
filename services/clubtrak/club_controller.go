package clubtrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/gin-gonic/gin"
)

type ClubParams struct {
	Name               string `json:"name"`
	Subscribers        int    `json:"Subscriber"`
	MeetingDescription string `json:"meetingDescription"`

	ClubID uint `json:"clubID"`
	//Club leader's user struct assigned to ClubAdminID
	ClubAdminID *models.User `json:"club,omitempty"`
}

// CreateClub godoc
// @Summary Creates a new club
// @Description Adds club to database
// @ID clubtrack-create-club
// @Tags clubtrack
// @Accept  json
// @Produce  json
// @Param createParams body clubtrak.ClubParams true "Create Club Params"
// @Success 200 {array} models.Clubtrack
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router clubtrak/clubs [post]
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
