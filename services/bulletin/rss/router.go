package rss

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// SetupRouter sets up the router for Bulletin RSS feeds
func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	c := NewController(db, cfg, log)

	// Bulletins general
	r.GET("/lostAndFound", c.ListLostAndFoundBulletins)
	r.GET("/job", c.ListJobBulletins)
	r.GET("/exchange", c.ListExchangeBulletins)
	r.GET("/announcement", c.ListAnnouncementBulletins)
	r.GET("/ride", c.ListRideBulletins)
	r.GET("/discussion", c.ListDiscussionBulletins)
}
