package ks19

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"

	"go.uber.org/zap"
)

func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	c := NewController(db, cfg, log) //intialize NewController
	rUser := r.Group("")             //group is empty
	rUser.Use(auth.RequireScopes(auth.ScopeUsers))
	rUser.GET("/:unixID", c.GetUserByUnix) //set path
}
