package ephcatch

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config) {
	c := NewController(db, cfg)

	r.GET("/ephcatchers", c.ListEphcatchers)
	r.GET("/ephcatchers/:ephcatcherID", c.GetEphcatcher)
	r.POST("/ephcatchers/:ephcatcherID/like", c.LikeEphcatcher)
	r.POST("/ephcatchers/:ephcatcherID/unlike", c.UnlikeEphcatcher)

	r.GET("/matches", c.ListMatches) // set seen:true

}
