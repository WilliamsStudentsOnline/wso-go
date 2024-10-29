package coursescheduler

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	c := NewController(db, cfg, log)
	writer := r.Group("")
	writer.Use(auth.RequireScopes(auth.ScopeUsers, auth.ScopeAdminAll))

	writer.GET("/get", c.GetCourseSelectionsByUser)
	writer.POST("/set", c.SetCourseSelectionsByUser)
}
