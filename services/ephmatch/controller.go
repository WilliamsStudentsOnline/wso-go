package ephmatch

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/jinzhu/gorm"
)

type Controller struct {
	services.BaseController
	// Put a model here, like:
	ephmatchModel   *models.EphmatchModel
	ephmatcherModel *models.EphmatcherModel
}

// Construct a new dormtrak controller
func NewController(db *gorm.DB, cfg *config.Config) *Controller {
	return &Controller{
		ephmatchModel:   models.NewEphmatchModel(db),
		ephmatcherModel: models.NewEphmatcherModel(db),
	}
}
