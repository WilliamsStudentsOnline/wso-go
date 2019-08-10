package dormtrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	search "github.com/WilliamsStudentsOnline/wso-go/lib/search/dormtrak"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
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
func NewController(db *gorm.DB, cfg *config.Config) *Controller {
	return &Controller{
		neighborhoodModel: models.NewNeighborhoodModel(db),
		dormModel:         models.NewDormModel(db),
		dormRoomModel:     models.NewDormRoomModel(db),
		reviewModel:       models.NewDormtrakReviewModel(db),
		userModel:         models.NewUserModel(db),
		dormtrakSearch:    search.NewSearchDormtrak(db, cfg),
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
