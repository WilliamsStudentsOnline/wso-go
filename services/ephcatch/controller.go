package ephcatch

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/jinzhu/gorm"
)

type Controller struct {
	services.BaseController
	// Put a model here, like:
	ephcatchModel   *models.EphcatchModel
	ephcatcherModel *models.EphcatcherModel
}

// Construct a new dormtrak controller
func NewController(db *gorm.DB, cfg *config.Config) *Controller {
	return &Controller{
		ephcatchModel:   models.NewEphcatchModel(db),
		ephcatcherModel: models.NewEphcatcherModel(db),
	}
}
