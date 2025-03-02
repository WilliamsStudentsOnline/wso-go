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
		ClubAdminID:        userID,
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
	err = t.clubModel.DeleteClub(club.ID)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, club)

}

type Category string

const (
	//Using the same categories as defined in interal spreadsheet of RSOs
	CategoryClubSports                      Category = "club sport"
	CategoryDance                           Category = "dance performance"
	CategoryAcademicAndHonors               Category = "academic and honors"
	CategoryAdvocacyDebatePolitical         Category = "advocacy, debate, and political"
	CategoryAffinityCulturallyBasedMinco    Category = "affinity, culterally based, and MiNCO"
	CategoryArtsEntertainment               Category = "arts and entertainment"
	CategoryCommunitySupportServiceLearning Category = "community Support and/or Service Learning"
	CategoryEnvironmentSustainability       Category = "environmental and sustainability"
	CategoryHealthWellness                  Category = "health and wellness"
	CategoryProfessionalCareer              Category = "professional and career"
	CategoryRecreationSports                Category = "recreation and sports"
	CategoryReligiousSpiritual              Category = "religious and spiritual"
)

type ClubUpdateParams struct {
	Name               string   `json:"name"`
	Category           Category `gnorm:"type=ENUM('club sport', 'dance performance', 'academic and honors', 'advocacy, debate, and political', 'affinity, culterally based, and MiNCO', 'arts and entertainment', 'community Support and/or Service Learning', 'environmental and sustainability', 'health and wellness', 'professional and career', 'recreation and sports', 'religious and spiritual');not null" json:"category"`
	MeetingDescription string   `json:"meetingDescription"`
	ClubDescription    string   `json:"clubDescription"`
	ClubPhotoFilePath  string   `json:"clubPhoto"`
	ContactEmail       string   `json:"contactEmail"`
	ContactPhoneNumber string   `json:"contactPhoneNumber"`
	Website            string   `json:"website"`

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
// @Param updateParams body clubtrak.ClubUpdateParams true "Update Club Params"
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
	switch string(club.Category) {
	case string(CategoryClubSports),
		string(CategoryDance),
		string(CategoryAcademicAndHonors),
		string(CategoryAdvocacyDebatePolitical),
		string(CategoryAffinityCulturallyBasedMinco),
		string(CategoryArtsEntertainment),
		string(CategoryCommunitySupportServiceLearning),
		string(CategoryEnvironmentSustainability),
		string(CategoryHealthWellness),
		string(CategoryProfessionalCareer),
		string(CategoryRecreationSports),
		string(CategoryReligiousSpiritual):
	default:
		//Case where category is not valid
		t.RespondError(c, err)
		return
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
