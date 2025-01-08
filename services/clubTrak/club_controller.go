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
	createData := ClubCreateParams{}
	err := c.ShouldBind(&createData)
	if err != nil {
		//t.RespondBadBind(c, err)
		return
	}

	p1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "Student 1",
		UnixID: "s1",
	}

	//Create a copy of the schema with relevent data for a given club
	club := models.ClubTrak{
		NumMembers:   createData.NumMembers,
		ClubLeaders:  createData.ClubLeaders,
		Description:  createData.Description,
		MeetingTimes: createData.MeetingTimes,
		Events:       createData.Events,
		ClubID:       createData.ClubID,
		Club:         &p1,
		Name:         createData.Name,
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
	club := models.ClubTrak{
		NumMembers:   "Test",
		ClubLeaders:  "Test",
		Description:  "Test",
		MeetingTimes: "Test",
		Events:       "Test",
		ClubID:       700,
		Club:         &p1,
		Name:         "Test",
	}
	t.RespondCreated(c, club)
}

type ClubCreateParams struct {
	NumMembers   string `gorm:"size:65535" json:"numMembers"`
	ClubLeaders  string `json:"clubLeaders"`
	Description  string `json:"description"`
	MeetingTimes string `json:"meetingTimes"`
	Events       string `json:"events"`

	// Belongs to user Some Club Leader
	ClubID uint   `json:"clubID"`
	Name   string `json:"name"`
}
