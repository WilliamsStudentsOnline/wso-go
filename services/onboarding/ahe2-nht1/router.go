package exercise

import (
	"fmt"
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	// var controller *Controller
	controller := NewController(db, cfg, log)

	rUser := r.Group("")
	rUser.Use(auth.RequireScopes(auth.ScopeUsers))

	rUser.GET("/:unixID", controller.GetUserByUnix)

}

// Hints:
//
// 1. Create a group to deal with requiring the user scope: rUser := r.Group("")
// 2. To require the user scope, do rUser.Use(auth.RequireScopes(auth.ScopeUsers))
// 3. To map a URL to a function, do rUser.GET("/:fooBar", controller.FunctionGoesHere)
// , where fooBar can be a parameter passed by the user that is called fooBar (for
// this task, you might want to try doing it with unixID.
