package ahe2nht1

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	// var controller *Controller
	controller := NewController(db, cfg, log)

	rUser := r.Group("")
	rUser.Use(auth.RequireScopes(auth.ScopeUsers))

	rUser.GET("/:unixID", controller.GetUserByUnix)

}
