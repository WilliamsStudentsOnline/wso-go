package dormtrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/lib/pictures"
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
	pictureBackend    pictures.PictureBackend
}

// Construct a new dormtrak controller
func NewController(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) *Controller {
	pb, err := pictures.NewPictureBackend(cfg, log)
	if err != nil {
		log.Error(err)
		log.Warn("Using picture backend none")
		pb = pictures.NewPictureBackendDummy()
	}

	return &Controller{
		BaseController:    services.BaseController{Log: log},
		neighborhoodModel: models.NewNeighborhoodModel(db, log),
		dormModel:         models.NewDormModel(db, log),
		dormRoomModel:     models.NewDormRoomModel(db, log),
		reviewModel:       models.NewDormtrakReviewModel(db, log),
		userModel:         models.NewUserModel(db, log),
		dormtrakSearch:    search.NewSearchDormtrak(db, cfg, log),
		pictureBackend:    pb,
	}
}

func RemoveUserIDFromReviews(c *gin.Context, r []*models.DormtrakReview) {
	if auth.HasScope(c, auth.ScopeAdminAll, auth.ScopeDormtrakAdmin) {
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

func IsScopeLimited(c *gin.Context) bool {
	return !auth.HasScope(c, auth.ScopeAdminAll, auth.ScopeDormtrakAdmin, auth.ScopeDormtrakFull)
}

func IsScopeAdmin(c *gin.Context) bool {
	return auth.HasScope(c, auth.ScopeAdminAll, auth.ScopeDormtrakAdmin)
}
