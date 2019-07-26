package bulletin

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// Controller refers to the struct for the bulletinModel
type Controller struct {
	services.BaseController
	bulletinModel *models.BulletinModel
	userModel     *models.UserModel
}

// NewController constructs a new user controller
func NewController(db *gorm.DB) *Controller {
	return &Controller{
		bulletinModel: &models.BulletinModel{
			BaseModel: models.BaseModel{
				DB: db,
			},
		},
	}
}

type BulletinCreateParams struct {
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Body      string    `gorm:"size:65535" json:"body"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
}

// CreateBulletin creates a bulletin
func (t *Controller) CreateBulletin(c *gin.Context) {
	userID := services.GetUserID(c)

	// Bind update params
	createData := BulletinCreateParams{}
	err := c.ShouldBind(&createData)
	if err != nil {
		// TODO: Wait for errors to be implmented, use 1100
		// t.RespondError(c, lib.ErrorMalformedRequestData)
		return
	}

	// Title and Body cannot be blank
	if createData.Title == "" || createData.Body == "" {
		t.RespondAPIError(c, ErrorBulletinMissingParams)
		return
	}

	// Start and End dates cannot be blank
	if (createData.StartDate == time.Time{}) || (createData.EndDate == time.Time{}) {
		t.RespondAPIError(c, ErrorBulletinMissingDates)
		return
	}

	// Start Date has to be before end Date
	if createData.StartDate.After(createData.EndDate) {
		t.RespondAPIError(c, ErrorBulletinInvalidDates)
		return
	}

	user := new(models.User)
	if err = t.userModel.GetUserByID(userID, user); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			t.RespondAPIError(c, ErrorBulletinUserNotFound)
			return
		}

		t.RespondError(c, http.StatusInternalServerError, err)
		return
	}

	// Construct new bulletin
	bulletin := models.Bulletin{
		Type:      createData.Type,
		Title:     createData.Title,
		Body:      createData.Body,
		StartDate: createData.StartDate,
		EndDate:   createData.EndDate,

		// Author information
		UserID: user.ID,
		User:   user,
	}

	err = t.bulletinModel.CreateBulletin(&bulletin)
	if err != nil {
		t.RespondError(c, http.StatusInternalServerError, err)
		return
	}

	// Todo: replace with respondcreated
	t.RespondOK(c, bulletin)
}

// FetchAllBulletins Fetches all bulletins
func (t *Controller) FetchAllBulletins(c *gin.Context) {
	bulletinType := c.Query("type")

	var bulletins []models.Bulletin
	var err error

	if bulletinType == "" {
		err = t.bulletinModel.GetAllBulletins(&bulletins)
	} else {
		err = t.bulletinModel.GetAllBulletinsByType(&bulletins, bulletinType)
	}

	if err != nil {
		t.RespondError(c, http.StatusInternalServerError, err)
		return
	}

	t.RespondOK(c, bulletins)
}

// GetBulletin Gets bulletin by id
func (t *Controller) GetBulletin(c *gin.Context) {
	var bulletinID uint
	var err error

	// Decode bulletinID.
	bulletinID, err = services.GetUIntParam(c, "bulletinID")
	if err != nil {
		t.RespondError(c, http.StatusBadRequest, errors.New("could not parse bulletin id"))
		return
	}

	// Do database query
	var bulletin models.Bulletin
	err = t.bulletinModel.GetBulletinByID(bulletinID, &bulletin)
	if err != nil {
		t.RespondError(c, http.StatusInternalServerError, err)
		return
	}

	t.RespondOK(c, bulletin)
}

// UpdateBulletin Updates bulletin by id
func (t *Controller) UpdateBulletin(c *gin.Context) {
	// Decode parameter
	bulletinID, err := services.GetUIntParam(c, "bulletinID")
	if err != nil {
		t.RespondError(c, http.StatusBadRequest, errors.New("could not parse bulletin id"))
		return
	}

	// Do database query to get bulletin
	var bulletin models.Bulletin
	err = t.bulletinModel.GetBulletinByID(bulletinID, &bulletin)
	if err != nil {
		t.RespondError(c, http.StatusInternalServerError, err)
		return
	}

	// Must only be able to update own bulletin
	if userID := bulletin.UserID; userID != services.GetUserID(c) {
		fmt.Println(userID)
		t.RespondError(c, http.StatusForbidden, errors.New("can only update own bulletin"))
		return
	}

	// Bind update params
	var update map[string]interface{}

	err = c.ShouldBind(&update)
	if err != nil {
		t.RespondError(c, http.StatusBadRequest, errors.New("could not parse malformed request data"))
		return
	}

	// Update the bulletin in the db
	err = t.bulletinModel.UpdateBulletin(bulletinID, update)
	if err != nil {
		t.RespondError(c, http.StatusInternalServerError, err)
		return
	}

	// Return nothing
	t.RespondOK(c, nil)

}

// DeleteBulletin Deletes bulletin by id
// TODO: adding scoping to only allow if admin
func (t *Controller) DeleteBulletin(c *gin.Context) {
	var bulletinID uint
	var err error

	// Decode bulletinID.
	bulletinID, err = services.GetUIntParam(c, "bulletinID")
	if err != nil {
		t.RespondError(c, http.StatusBadRequest, errors.New("could not parse bulletin id"))
		return
	}

	// Do database query to get bulletin
	var bulletin models.Bulletin
	err = t.bulletinModel.GetBulletinByID(bulletinID, &bulletin)
	if err != nil {
		t.RespondError(c, http.StatusInternalServerError, err)
		return
	}

	// Must only be able to delete own bulletin
	if userID := bulletin.UserID; userID != services.GetUserID(c) {
		fmt.Println(userID)
		t.RespondError(c, http.StatusForbidden, errors.New("can only update own bulletin"))
		return
	}

	// Delete Bulletin
	err = t.bulletinModel.DeleteBulletinByID(bulletinID, &bulletin)
	if err != nil {
		t.RespondError(c, http.StatusInternalServerError, err)
		return
	}

	t.RespondOK(c, bulletinID)
}
