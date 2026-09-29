package ephmatch

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	c := NewController(db, cfg, log)
	gone := c.RespondGone

	r.GET("/availability", gone)

	selfGroup := r.Group("", auth.RequireScopes(auth.ScopeEphmatch, auth.ScopeAdminAll))
	selfGroup.GET("/profile", gone)
	selfGroup.POST("/profile", gone)
	selfGroup.PATCH("/profile", gone)
	selfGroup.DELETE("/profile", gone)
	selfGroup.PUT("/profile/photo", gone)
	selfGroup.DELETE("/profile/photo", gone)

	matchesGroup := selfGroup.Group("", auth.RequireScopes(auth.ScopeEphmatchMatches, auth.ScopeAdminAll))
	matchesGroup.GET("/matches", gone)
	matchesGroup.GET("/matches-count", gone)
	matchesGroup.DELETE("/matches/:matchUserID", gone)

	profilesGroup := matchesGroup.Group("", auth.RequireScopes(auth.ScopeEphmatchProfiles, auth.ScopeAdminAll))
	profilesGroup.GET("/profiles", gone)
	profilesGroup.GET("/profiles/:profileUserID", gone)
	profilesGroup.PUT("/profiles/:profileUserID/relation", gone)
}
