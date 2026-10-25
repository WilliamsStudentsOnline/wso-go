package banner

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

func ensureClient(cfg *config.Config, log *zap.SugaredLogger) {
	if client != nil {
		return
	}
	if err := ConfigureClient(cfg.RedisAddr); err != nil {
		log.Debugf("redis unavailable: %v", err)
	}
}

// SetupPublicRouter mounts the unauthenticated GET banner route.
func SetupPublicRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	ensureClient(cfg, log)
	c := NewController(db, cfg, log)
	r.GET("", c.GetBanner)
}

// SetupAdminRouter mounts the admin-only POST banner route.
func SetupAdminRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	ensureClient(cfg, log)
	c := NewController(db, cfg, log)

	admin := r.Group("")
	admin.Use(auth.RequireScopes(auth.ScopeAdminAll))
	admin.POST("", c.SetBanner)
}
