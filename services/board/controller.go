package board

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

// Controller for the board service.
type Controller struct {
	services.BaseController
	threadModel *models.BoardThreadModel
	postModel   *models.BoardPostModel
	flagModel   *models.FlagModel
	userModel   *models.UserModel
}

func NewController(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) *Controller {
	return &Controller{
		BaseController: services.BaseController{Log: log},
		threadModel:    models.NewBoardThreadModel(db, log),
		postModel:      models.NewBoardPostModel(db, log),
		flagModel:      models.NewFlagModel(db, log),
		userModel:      models.NewUserModel(db, log),
	}
}

func hasUserAuth(c *gin.Context) bool {
	return auth.HasScope(c, auth.ScopeUsers, auth.ScopeAdminAll)
}

func isAdmin(c *gin.Context) bool {
	return auth.HasScope(c, auth.ScopeAdminAll)
}

func sanitizeThread(c *gin.Context, t *models.BoardThread) {
	if t == nil {
		return
	}
	if hasUserAuth(c) {
		sanitize.User(t.User, c)
	} else {
		t.User = nil
		t.ExUserName = ""
	}
	sanitizePosts(c, t.Posts)
}

func sanitizeThreads(c *gin.Context, threads []*models.BoardThread) {
	for _, t := range threads {
		sanitizeThread(c, t)
	}
}

func sanitizePosts(c *gin.Context, posts []*models.BoardPost) {
	userAuth := hasUserAuth(c)
	for _, p := range posts {
		if p == nil {
			continue
		}
		if userAuth {
			sanitize.User(p.User, c)
		} else {
			p.User = nil
			p.ExUserName = ""
		}
	}
}

func canMutate(c *gin.Context, ownerID uint) bool {
	return services.GetUserID(c) == ownerID || isAdmin(c)
}
