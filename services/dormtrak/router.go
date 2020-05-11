package dormtrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	c := NewController(db, cfg, log)

	// Neighborhoods endpoint
	r.GET("/neighborhoods", c.ListNeighborhoods)
	r.GET("/neighborhoods/:neighborhoodID", c.GetNeighborhood)
	r.GET("/neighborhoods/:neighborhoodID/facts", c.GetNeighborhoodFacts)

	// Dorms endpoint
	r.GET("/dorms", c.ListDorms)
	r.GET("/dorms/:dormID", c.GetDorm)
	r.GET("/dorms/:dormID/rooms", c.ListDormRooms)
	r.GET("/dorms/:dormID/facts", c.GetDormFacts)

	// Room endpoint
	r.GET("/rooms/:dormRoomID", c.GetDormRoom)

	// Get rankings overall
	r.GET("/rankings", c.GetRankings)

	// Must be logged in:
	// Get reviews by dormID, dormRoomID, userID, pagination
	r.GET("/reviews", c.ListReviews)
	r.GET("/reviews/:reviewID", c.GetReview)

	// ScopeDormtrakWrite ensures that the person is a student and in the upperclasses
	writer := r.Group("")
	writer.Use(auth.RequireScopes(auth.ScopeDormtrakWrite))
	// Writing review operations
	writer.POST("/reviews", c.CreateReview)
	writer.PATCH("/reviews/:reviewID", c.UpdateReview)
	writer.DELETE("/reviews/:reviewID", c.DeleteReview)
}
