package rss

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// SetupRouter sets up the router for Bulletins
func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config) {
	c := NewController(db, cfg)

	// Bulletins general
	r.GET("/lostAndFound", c.ListLostAndFoundBulletins)
	r.GET("/job", c.ListJobBulletins)
	r.GET("/exchange", c.ListExchangeBulletins)
	r.GET("/announcement", c.ListAnnouncementBulletins)
	r.GET("/ride", c.ListRideBulletins)
	r.GET("/discussion", c.ListDiscussionBulletins)
}
