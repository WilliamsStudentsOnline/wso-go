package ephmatch

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config) {
	c := NewController(db, cfg)

	r.GET("/ephmatchers", c.ListEphmatchers)
	r.GET("/ephmatchers/:ephmatcherID", c.GetEphmatcher)
	r.POST("/ephmatchers/:ephmatcherID/like", c.LikeEphmatcher)
	r.POST("/ephmatchers/:ephmatcherID/unlike", c.UnlikeEphmatcher)

	r.GET("/matches", c.ListMatches) // TODO: set seen:true

}
