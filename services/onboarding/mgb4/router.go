package mgb4

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	// Initialize the controller and group
	controller := NewController(db, cfg, log)
	rUser := r.Group("")

	// Route requests
	rUser.Use(auth.RequireScopes(auth.ScopeUsers))
	{
		rUser.GET("/:unixID", controller.GetUserByUnix)
	}
}
