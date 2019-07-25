package bulletin

import (
	"errors"
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type Controller struct {
	services.BaseController
	bulletinModel *models.BulletinModel
}

// Construct a new user controller
func NewController(db *gorm.DB) *Controller {
	return &Controller{
		bulletinModel: &models.BulletinModel{
			BaseModel: models.BaseModel{
				DB: db,
			},
		},
	}
}

// Fetch all bulletins
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

// Get bulletin by id
func (t *Controller) GetBulletin(c *gin.Context) {
	var bulletinID uint
	var err error

	// Decode bulletinID.
	bulletinID, err = services.GetUIntParam(c, "bulletin_id")
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
	bulletinID, err := services.GetUIntParam(c, "bulletin_id")
	if err != nil {
		t.RespondError(c, http.StatusBadRequest, errors.New("could not parse bulletin id"))
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

// Delete bulletin by id
// TODO: adding scoping to only allow if admin
func (t *Controller) DeleteBulletin(c *gin.Context) {
	var bulletinID uint
	var err error

	// Decode bulletinID.
	bulletinID, err = services.GetUIntParam(c, "bulletin_id")
	if err != nil {
		t.RespondError(c, http.StatusBadRequest, errors.New("could not parse bulletin id"))
		return
	}

	// Do database query
	var bulletin models.Bulletin
	err = t.bulletinModel.DeleteBulletinByID(bulletinID, &bulletin)
	if err != nil {
		t.RespondError(c, http.StatusInternalServerError, err)
		return
	}

	t.RespondOK(c, bulletinID)
}
