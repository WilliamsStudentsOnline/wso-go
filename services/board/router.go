package board

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// SetupRouter mounts board routes on the given router group
func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	c := NewController(db, cfg, log)

	writer := r.Group("")
	writer.Use(auth.RequireScopes(auth.ScopeWriteSelf, auth.ScopeBulletinWrite))

	r.GET("/threads", c.ListThreads)
	r.GET("/threads/:threadID", c.GetThread)
	writer.POST("/threads", c.CreateThread)
	writer.PATCH("/threads/:threadID", c.UpdateThread)
	writer.DELETE("/threads/:threadID", c.DeleteThread)

	writer.POST("/threads/:threadID/replies", c.CreateReply)
	writer.PATCH("/replies/:replyID", c.UpdateReply)
	writer.DELETE("/replies/:replyID", c.DeleteReply)

	writer.POST("/threads/:threadID/flag", c.FlagThread)
	writer.POST("/replies/:replyID/flag", c.FlagReply)
}
