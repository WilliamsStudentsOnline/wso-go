package rss

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/services/bulletin"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// SetupRouter sets up the router for Bulletins
func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config) {
	c := NewController(db)

	// Bulletins et all
	r.GET("/bulletins", c.ListBulletins)
	r.GET("/bulletins/:bulletinID", c.GetBulletin)
	writer.POST("/bulletins", c.CreateBulletin)
	writer.PATCH("/bulletins/:bulletinID", c.UpdateBulletin)
	writer.DELETE("/bulletins/:bulletinID", c.DeleteBulletin)

	// Rides (special)
	r.GET("/rides", c.ListRides)
	r.GET("/rides/:rideID", c.GetRide)
	writer.POST("/rides", c.CreateRide)
	writer.PATCH("/rides/:rideID", c.UpdateRide)
	writer.DELETE("/rides/:rideID", c.DeleteRide)

	// Users don't create discussions, they create posts, which have a discussion.
	// CreateDiscussion is just shorthand for CreatePost without an existing Discussion.
	// Thus, after they are created, users are unable to edit or delete discussions, only posts. However,
	// they can delete a discussion if everyone deletes their posts.
	r.GET("/discussions", c.ListDiscussions)
	r.GET("/discussions/:discussionID", c.GetDiscussion)
	r.GET("/discussions/:discussionID/posts", c.GetDiscussionPosts)
	writer.POST("/discussions", c.CreateDiscussion)
	// Only admins can delete discussions
	admin := r.Group("")
	admin.Use(auth.RequireScopes(auth.ScopeAdminAll))
	admin.DELETE("/discussions/:discussionID", c.DeleteDiscussion)

	r.GET("/posts/:postID", c.GetPost)
	writer.POST("/posts", c.CreatePost)
	writer.PATCH("/posts/:postID", c.UpdatePost)
	writer.DELETE("/posts/:postID", c.DeletePost)
}
