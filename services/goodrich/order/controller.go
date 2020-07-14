package order

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type Controller struct {
	services.BaseController                    // Inherit the base controller
	orderModel              *models.OrderModel // DB communication to get order info
}

// NewController constructs a new onboarding controller
func NewController(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) *Controller {
	return &Controller{
		BaseController: services.BaseController{Log: log},
		orderModel:     models.NewOrderModel(db, log),
	}
}
