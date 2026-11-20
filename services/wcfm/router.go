package wcfm

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/redis_util"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ensureClient(cfg *config.Config, log *zap.SugaredLogger) {
	if store != nil {
		return
	}
	store = redis_util.NewLazy(cfg.RedisAddr, redis_util.CacheDatabaseID)
	if store.Client() == nil {
		log.Debugf("redis unavailable, will retry")
	}
}

func SetupRouter(r gin.IRouter, cfg *config.Config, log *zap.SugaredLogger) {
	ensureClient(cfg, log)
	c := NewController(cfg, log)
	r.GET("/now", c.GetNow)
	r.GET("/schedule", c.GetSchedule)
}
