package goodrich

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
	orderModel *models.GoodrichOrderModel
	menuModel  *models.GoodrichMenuItemModel
	cfg        *config.Config
}

// Construct a new user controller
func NewController(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) *Controller {
	return &Controller{
		BaseController: services.BaseController{Log: log},
		orderModel:     models.NewGoodrichOrderModel(db, log),
		menuModel:      models.NewGoodrichMenuItemModel(db, log),
		cfg:            cfg,
	}
}
