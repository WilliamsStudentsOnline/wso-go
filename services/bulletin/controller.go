package bulletin

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// Controller refers to the struct for the bulletinModel
type Controller struct {
	services.BaseController
	bulletinModel *models.BulletinModel
	rideModel     *models.BulletinRideModel
}

// NewController constructs a new user controller
func NewController(db *gorm.DB) *Controller {
	return &Controller{
		bulletinModel: models.NewBulletinModel(db),
		rideModel:     models.NewBulletinRideModel(db),
	}
}

func hasUserAuth(c *gin.Context) bool {
	return auth.HasScope(c, auth.ScopeUsers, auth.ScopeAdminAll)
}

func removeUserInfoFromBulletin(c *gin.Context, bulletins []*models.Bulletin) {
	if hasUserAuth(c) {
		return
	}

	for _, b := range bulletins {
		b.User = nil
	}
}

func removeUserInfoFromRides(c *gin.Context, rides []*models.BulletinRide) {
	if hasUserAuth(c) {
		return
	}

	for _, r := range rides {
		r.User = nil
	}
}
