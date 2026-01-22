package clubtrak

import (
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

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
// @Router /clubtrak/admin/clubs [post]
func (t *Controller) CreateClub(c *gin.Context) {

	//Check that user is admin
	isAdmin := auth.HasScope(c, auth.ScopeAdminAll)

	if !isAdmin {
		t.RespondError(c, lib.ErrorNoScopeAuthorization)
		return
	}

	// Bind create params
	createData := ClubParams{}
	err := c.ShouldBind(&createData)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	userID := createData.ClubAdminID
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
		ClubAdminID:        userID,
		Name:               createData.Name,
		Category:           models.Category(createData.Category),
		ClubDescription:    createData.ClubDescription,
		ClubPhotoFilePath:  createData.ClubPhotoFilePath,
		ContactEmail:       createData.ContactEmail,
		ContactPhoneNumber: createData.ContactPhoneNumber,
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
// @Router /clubtrak/admin/clubs [get]
func (t *Controller) AdminGetAllClubs(c *gin.Context) {
	var clubs []*models.Club
	var err error

	opts := models.GetAllClubsOptions{}
	if err = c.ShouldBindQuery(&opts); err != nil {
		t.RespondBadBind(c, err)
		return
	}

	if query, ok := c.GetQuery("q"); ok {
		err = t.clubModel.SearchClubs(t.clubModel.DB, query, &clubs, &opts)
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
// @Router /clubtrak/admin/clubs/:clubID [delete]
func (t *Controller) AdminDeleteClub(c *gin.Context) {
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
	err = t.clubModel.DeleteClub(club.ID)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, club)

}

// UpdateClub godoc
// @Summary Updates Club column
// @Description Updates any of the data shown on a given club's webpage
// @ID clubtrack-update-club
// @Tags clubtrak
// @Accept  json
// @Produce  json
// @Param clubID path uint true "Club ID"
// @Param updateParams body clubtrak.ClubUpdateParams true "Update Club Params"
// @Success 200 {object} models.Club
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /clubtrak/admin/clubs/:clubID [patch]
func (t *Controller) AdminUpdateClub(c *gin.Context) {
	clubID, err := services.GetUIntParam(c, "clubID")
	if err != nil {
		t.RespondErrorCode(c, http.StatusBadRequest, err)
		return
	}

	// Bind update params
	updateData := ClubUpdateParams{}
	err = c.ShouldBind(&updateData)
	if err != nil {
		t.RespondBadBind(c, err)
		return
	}

	// Get club (also serves as valid club check)
	var club models.Club
	var opts models.Options
	err = t.clubModel.GetClubByID(clubID, &club, opts)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	//Update fields
	club.Category = models.Category(lib.StrDefaults(string(updateData.Category), string(club.Category)))
	//Validate Category
	if t.clubModel.ValidateCategory(club.Category) != nil {
		t.RespondError(c, err)
	}

	club.MeetingDescription = lib.StrDefaults(updateData.MeetingDescription, club.MeetingDescription)
	club.ClubDescription = lib.StrDefaults(updateData.ClubDescription, club.ClubDescription)
	club.ClubPhotoFilePath = lib.StrDefaults(updateData.ClubPhotoFilePath, club.ClubPhotoFilePath)
	club.ContactEmail = lib.StrDefaults(updateData.ContactEmail, club.ContactEmail)
	club.ContactPhoneNumber = lib.StrDefaults(updateData.ContactPhoneNumber, club.ContactPhoneNumber)
	club.Website = lib.StrDefaults(updateData.Website, club.Website)

	// Do db update
	err = t.clubModel.UpdateClub(&club)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, club)

}
