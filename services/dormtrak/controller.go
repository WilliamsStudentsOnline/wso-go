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
	dormModel         *models.DormModel
	dormRoomModel     *models.DormRoomModel
}

// Construct a new user controller
func NewController(db *gorm.DB) *Controller {
	return &Controller{
		neighborhoodModel: models.NewNeighborhoodModel(db),
		dormModel:         models.NewDormModel(db),
		dormRoomModel:     models.NewDormRoomModel(db),
	}
}
