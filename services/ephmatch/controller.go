package ephmatch

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type Controller struct {
	services.BaseController
	// Put a model here, like:
	ephmatchModel *models.EphmatchModel
	profileModel  *models.EphmatchProfileModel
}

// Construct a new dormtrak controller
func NewController(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) *Controller {
	return &Controller{
		BaseController: services.BaseController{Log: log},
		ephmatchModel:  models.NewEphmatchModel(db, log),
		profileModel:   models.NewEphmatchProfileModel(db, log),
	}
}
