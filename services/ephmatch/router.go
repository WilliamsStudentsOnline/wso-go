package ephmatch

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config) {
	c := NewController(db, cfg)

	r.GET("/profiles", c.ListProfiles)
	r.GET("/profiles/:profileUserID", c.GetProfile)
	r.POST("/profiles/:profileUserID/like", c.LikeProfile)
	r.POST("/profiles/:profileUserID/unlike", c.UnlikeProfile)

	// Get self profile
	r.GET("/profile", c.GetSelfProfile)
	// Create profile
	r.POST("/profile", c.CreateProfile)
	// Edit profile
	r.PATCH("/profile", c.UpdateProfile)
	// Delete profile
	r.DELETE("/profile", c.DeleteProfile)

	r.GET("/matches", c.ListMatches) // TODO: set seen:true

}
