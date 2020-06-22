package rmn1

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// With consultation from the canonical package
func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	// create a Controller
	controller := NewController(db, cfg, log)

	// create a group to deal with requiring the user scope
	rUser := r.Group("")

	// require the user scope
	rUser.Use(auth.RequireScopes(auth.ScopeUsers))

	// map a URL to a function
	rUser.GET("/:unixID", controller.GetUserByUnix)
}
