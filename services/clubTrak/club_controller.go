package clubtrak

import (
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/gin-gonic/gin"
)

type ClubCreateParams struct {
	Subscribers  int    `json:"Subscriber"`
	ClubLeaders  string `json:"clubLeaders"`
	Description  string `json:"description"`
	MeetingTimes string `json:"meetingTimes"`
	Events       string `json:"events"`

	// Belongs to user Some Club Leader
	ClubID      uint   `json:"clubID"`
	ClubAdminID *User  `json:"club,omitempty"`
	Name        string `json:"name"`
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
	createData := ClubCreateParams{}
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
		MeetingDescription: createData.Description,
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

func (t *Controller) TestingClub(c *gin.Context) {
	p1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "Student 1",
		UnixID: "s1",
	}
	c.JSON(http.StatusOK, p1)
	club := models.Club{
		Subscribers:        0,
		MeetingDescription: "Test",
		ClubAdmin:          700,
		ClubAdminID:        &p1,
		Name:               "Test",
	}
	c.JSON(http.StatusOK, club)
	t.RespondCreated(c, club)
	t.Log.Info("Testing Endpoint Reached")
}
