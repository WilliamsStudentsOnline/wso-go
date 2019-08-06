package dormtrak

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
	// Put a model here, like:
	dormtrakModel *models.Dormtrak
}

// Construct a new user controller
func NewController(db *gorm.DB) *Controller {
	return &Controller{
		dormtrakModel: &models.Dormtrak{
			BaseModel: models.BaseModel{
				DB: db,
			},
		},
	}
}

// Endpoint example
/*
func (t *Controller) FetchAllUsers(c *gin.Context) {
	var users []models.User
	err := t.userModel.GetAllUsers(&users)
	
	if err != nil {
		t.RespondErrorCode(http.StatusInternalServerError, err, c)
		return
	}

	t.RespondOK(users, c)
}
*/
