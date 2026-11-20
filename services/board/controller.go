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

type Controller struct {
	services.BaseController
	threadModel        *models.BoardThreadModel
	postModel          *models.BoardPostModel
	pollModel          *models.BoardPollModel
	userModel          *models.UserModel
	moderatorTaskModel *models.ModeratorTaskModel
}

func NewController(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) *Controller {
	return &Controller{
		BaseController:     services.BaseController{Log: log},
		threadModel:        models.NewBoardThreadModel(db, log),
		postModel:          models.NewBoardPostModel(db, log),
		pollModel:          models.NewBoardPollModel(db, log),
		userModel:          models.NewUserModel(db, log),
		moderatorTaskModel: models.NewModeratorTaskModel(db, log),
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
	sanitizePosts(c, t.Posts, t.UserID)
	redactAuthor(c, t.IsAnonymous(), &t.UserID, &t.User, &t.ExUserName)
}

func sanitizeThreads(c *gin.Context, threads []*models.BoardThread) {
	for _, t := range threads {
		sanitizeThread(c, t)
	}
}

func sanitizePosts(c *gin.Context, posts []*models.BoardPost, opUserID uint) {
	viewerID := services.GetUserID(c)
	for _, p := range posts {
		if p == nil {
			continue
		}
		p.ByOP = p.UserID == opUserID
		p.IsMine = viewerID != 0 && p.UserID == viewerID
		sanitizeThread(c, p.Thread)
		redactAuthor(c, p.IsAnonymous(), &p.UserID, &p.User, &p.ExUserName)
	}
}

func redactAuthor(c *gin.Context, anonymous bool, userID *uint, user **models.User, exUserName *string) {
	if (anonymous && !isAdmin(c)) || !hasUserAuth(c) {
		*userID = 0
		*user = nil
		*exUserName = ""
		return
	}
	sanitize.User(*user, c)
}

func canMutate(c *gin.Context, ownerID uint) bool {
	return services.GetUserID(c) == ownerID || isAdmin(c)
}
