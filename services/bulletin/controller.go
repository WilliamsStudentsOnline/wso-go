package bulletin

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/sanitize"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// Controller refers to the struct for the bulletinModel
type Controller struct {
	services.BaseController
	bulletinModel   *models.BulletinModel
	rideModel       *models.BulletinRideModel
	discussionModel *models.DiscussionModel
	postModel       *models.PostModel
	userModel       *models.UserModel
}

// NewController constructs a new user controller
func NewController(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) *Controller {
	return &Controller{
		BaseController:  services.BaseController{Log: log},
		bulletinModel:   models.NewBulletinModel(db, log),
		rideModel:       models.NewBulletinRideModel(db, log),
		discussionModel: models.NewDiscussionModel(db, log),
		postModel:       models.NewPostModel(db, log),
		userModel:       models.NewUserModel(db, log),
	}
}

func hasUserAuth(c *gin.Context) bool {
	return auth.HasScope(c, auth.ScopeUsers, auth.ScopeAdminAll)
}

func removeUserInfoFromBulletin(c *gin.Context, bulletins []*models.Bulletin) {
	userAuth := hasUserAuth(c)
	for _, b := range bulletins {
		if userAuth {
			sanitize.User(b.User, c)
		} else {
			b.User = nil
		}
	}
}

func removeUserInfoFromRides(c *gin.Context, rides []*models.BulletinRide) {
	userAuth := hasUserAuth(c)
	for _, r := range rides {
		if userAuth {
			sanitize.User(r.User, c)
		} else {
			r.User = nil
		}
	}
}

func removeUserInfoFromDiscussions(c *gin.Context, discussions []*models.Discussion) {
	userAuth := hasUserAuth(c)
	for _, d := range discussions {
		if userAuth {
			sanitize.User(d.User, c)
		} else {
			d.User = nil
			d.ExUserName = ""
		}

		if d.Posts != nil {
			removeUserInfoFromPosts(c, d.Posts)
		}
	}
}

func removeUserInfoFromPosts(c *gin.Context, posts []*models.Post) {
	userAuth := hasUserAuth(c)
	for _, p := range posts {
		if userAuth {
			sanitize.User(p.User, c)
		} else {
			p.User = nil
			p.ExUserName = ""
		}
	}
}
