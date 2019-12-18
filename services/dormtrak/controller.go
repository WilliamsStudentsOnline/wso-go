package dormtrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	search "github.com/WilliamsStudentsOnline/wso-go/lib/search/dormtrak"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type Controller struct {
	services.BaseController
	// Put a model here, like:
	neighborhoodModel *models.NeighborhoodModel
	dormModel         *models.DormModel
	dormRoomModel     *models.DormRoomModel
	reviewModel       *models.DormtrakReviewModel
	userModel         *models.UserModel
	dormtrakSearch    search.SearchDormtrak
}

// Construct a new dormtrak controller
func NewController(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) *Controller {
	return &Controller{
		BaseController:    services.BaseController{Log: log},
		neighborhoodModel: models.NewNeighborhoodModel(db, log),
		dormModel:         models.NewDormModel(db, log),
		dormRoomModel:     models.NewDormRoomModel(db, log),
		reviewModel:       models.NewDormtrakReviewModel(db, log),
		userModel:         models.NewUserModel(db, log),
		dormtrakSearch:    search.NewSearchDormtrak(db, cfg, log),
	}
}

func RemoveUserIDFromReviews(c *gin.Context, r []*models.DormtrakReview) {
	if auth.HasScope(c, auth.ScopeAdminAll, auth.ScopeFactrakAdmin) {
		return
	}

	userID := services.GetUserID(c)

	for _, review := range r {
		if review.UserID == userID {
			continue
		}
		review.UserID = 0
		review.User = nil
	}
}
