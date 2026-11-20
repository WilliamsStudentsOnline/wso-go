package banner

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/lib/redis_util"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

func ensureClient(cfg *config.Config, log *zap.SugaredLogger) {
	if store != nil {
		return
	}
	store = redis_util.NewLazy(cfg.RedisAddr, redis_util.AdminDatabaseID)
	if store.Client() == nil {
		log.Debugf("redis unavailable, will retry")
	}
}

func SetupPublicRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	ensureClient(cfg, log)
	c := NewController(db, cfg, log)
	r.GET("", c.GetBanner)
}

func SetupAdminRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	ensureClient(cfg, log)
	c := NewController(db, cfg, log)

	admin := r.Group("")
	admin.Use(auth.RequireScopes(auth.ScopeAdminAll))
	admin.POST("", c.SetBanner)
}
