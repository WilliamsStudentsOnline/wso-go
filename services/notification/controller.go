package notification

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type Controller struct {
	services.BaseController
	cfg                *config.Config
	notifSettingsModel *models.NotificationSettingsModel
	notifTokenModel    *models.NotificationTokenModel
}

// Construct a new user controller
func NewController(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) *Controller {
	return &Controller{
		BaseController:     services.BaseController{Log: log},
		cfg:                cfg,
		notifSettingsModel: models.NewNotificationSettingsModel(db, log),
		notifTokenModel:    models.NewNotificationTokenModel(db, log),
	}
}
