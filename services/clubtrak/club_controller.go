package clubtrak

import (
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
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
// @Tags clubtrak
// @Accept  json
// @Produce  json
// @Param createParams body clubtrak.ClubParams true "Create Club Params"
// @Success 200 {array} models.Club
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /clubtrak/clubs [post]
func (t *Controller) CreateClub(c *gin.Context) {
	userID := services.GetUserID(c)

	// Bind create params
	createData := ClubParams{}
	err := c.ShouldBind(&createData)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}
	//Check that User exists
	user := new(models.User)
	if err = t.userModel.GetUserByID(userID, user); err != nil {
		// Don't return 404; instead, return authed user not found
		if gorm.IsRecordNotFoundError(err) {
			c.Set(services.UpdateTokenKey, true)
			err = lib.ErrorAuthedUserNotFound
		}

		t.RespondError(c, err)
		return
	}

	//Create a copy of the schema with relevent data for a given club
	club := models.Club{
		Subscribers:        createData.Subscribers,
		MeetingDescription: createData.MeetingDescription,
		ClubAdmin:          createData.ClubID,
		ClubAdminID:        user,
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

// GetAllClubs godoc
// @Summary Lists all clubs
// @Description Lists all clubs in the database
// @ID clubtrack-get-all-clubs
// @Tags clubtrak
// @Accept  json
// @Produce  json
// @Param offset query int false "Offset Pagination"
// @Param limit query int false "Limit Pagination"
// @Param q query string false "Search Query"
// @Success 200 {array} models.Club
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /clubtrak/clubs [get]
func (t *Controller) GetAllClubs(c *gin.Context) {
	var clubs []*models.Club
	var err error

	opts := models.GetAllClubsOptions{}
	if err = c.ShouldBindQuery(&opts); err != nil {
		t.RespondBadBind(c, err)
		return
	}

	if query, ok := c.GetQuery("q"); ok {
		err = t.clubtrakSearch.SearchClubs(query, &clubs, &opts)
	} else {
		err = t.clubModel.GetAllClubs(&clubs, &opts)
	}

	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, clubs)

}

// DeleteClub godoc
// @Summary Deletes a Club
// @Description Deletes a Club given a clubID
// @ID clubtrack-delete-club
// @Tags clubtrak
// @Accept  json
// @Produce  json
// @Param clubID path uint true "Club ID"
// @Success 200 {object} models.Club
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /clubtrak/clubs/:clubID [delete]
func (t *Controller) DeleteClub(c *gin.Context) {
	// Get clubID
	clubID, err := services.GetUIntParam(c, "clubID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	var club models.Club
	var opts models.Options
	err = t.clubModel.GetClubByID(clubID, &club, opts)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	// Delete club
	err = t.clubModel.DeleteClub(&club)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, club)

}

type ReviewUpdateParams struct {
	Name               string `json:"name"`
	Category           string `json:"category"`
	Subscribers        int    `json:"subscribers"`
	MeetingDescription string `json:"meetingDescription"`
	ClubDescription    string `json:"clubDescription"`
	ClubPhoto          string `json:"clubPhoto"`

	// Belongs to some club leader
	ClubAdmin   uint         `json:"clubAdmin"`
	ClubAdminID *models.User `json:"clubAdminID,omitempty"`
}

// UpdateClub godoc
// @Summary Updates Club column
// @Description Updates any of the data shown on a given club's webpage
// @ID clubtrack-update-club
// @Tags clubtrak
// @Accept  json
// @Produce  json
// @Param clubID path uint true "Club ID"
// @Param updateParams body clubtrak.ReviewUpdateParams true "Update Club Params"
// @Success 200 {object} models.Club
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /clubtrak/clubs/:clubID [patch]
func (t *Controller) UpdateClub(c *gin.Context) {
	clubID, err := services.GetUIntParam(c, "clubID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	var club models.Club
	var opts models.Options
	err = t.clubModel.GetClubByID(clubID, &club, opts)
	if err != nil {
		t.RespondError(c, err)
		return
	}

}
