package user

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
	writer.Use(auth.RequireScopes(auth.ScopeWriteSelf))

	r.GET("", c.ListUsers)
	r.GET("/:userID", c.GetUser)
	writer.PATCH("/:userID", c.UpdateUser)
	writer.PUT("/:userID/tags", c.UpdateUserTags)
	writer.PUT("/:userID/photo", c.UploadProfilePhoto)
}
