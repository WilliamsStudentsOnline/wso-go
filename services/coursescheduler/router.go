package coursescheduler

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	if client == nil {
		if err := ConfigureClient(cfg.RedisAddr); err != nil {
			log.Debugf("redis unavailable: %v", err)
		}
	}

	c := NewController(db, cfg, log)
	writer := r.Group("")
	writer.Use(auth.RequireScopes(auth.ScopeUsers, auth.ScopeAdminAll))

	writer.GET("/course-selections/:userID", c.GetCourseSelectionsByUser)
	writer.PUT("/course-selections/:userID", c.SetCourseSelectionsByUser)
	// We intend for the PUT endput to be used, but to avoid confusion we also enable POST (since Get does both)
	writer.POST("/course-selections/:userID", c.SetCourseSelectionsByUser)
}
