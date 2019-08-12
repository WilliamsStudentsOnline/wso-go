package bulletin

import (
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/jinzhu/gorm"
)

// Controller refers to the struct for the bulletinModel
type Controller struct {
	services.BaseController
	bulletinModel *models.BulletinModel
	userModel     *models.UserModel
}

// NewController constructs a new user controller
func NewController(db *gorm.DB) *Controller {
	return &Controller{
		bulletinModel: models.NewBulletinModel(db),
		userModel:     models.NewUserModel(db),
	}
}
