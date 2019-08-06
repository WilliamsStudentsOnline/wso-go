package dormtrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/jinzhu/gorm"
)

type Controller struct {
	services.BaseController
	// Put a model here, like:
	neighborhoodModel *models.NeighborhoodModel
}

// Construct a new user controller
func NewController(db *gorm.DB) *Controller {
	return &Controller{
		neighborhoodModel: models.NewNeighborhoodModel(db),
	}
}
