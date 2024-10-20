package factrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
)

type Controller struct {
	services.BaseController
	// Put models here:
	clubModel *models.ClubModel
}
