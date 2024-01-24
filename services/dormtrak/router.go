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

	// Full dormtrak scoping:
	full := r.Group("")
	full.Use(auth.RequireScopes(auth.ScopeDormtrakFull))

	// Neighborhoods endpoint
	r.GET("/neighborhoods", c.ListNeighborhoods)
	r.GET("/neighborhoods/:neighborhoodID", c.GetNeighborhood)
	full.GET("/neighborhoods/:neighborhoodID/facts", c.GetNeighborhoodFacts)

	// Dorms endpoint
	r.GET("/dorms", c.ListDorms)
	r.GET("/dorms/:dormID", c.GetDorm)
	r.GET("/dorms/:dormID/rooms", c.GetDormRooms)
	full.GET("/dorms/:dormID/facts", c.GetDormFacts)

	// Rooms endpoint
	r.GET("/rooms/:roomID/photos", c.GetRoomPhotos)

	// Get rankings overall
	full.GET("/rankings", c.GetRankings)

	// Must be logged in:
	// Get reviews by dormID, dormRoomID, userID, pagination
	full.GET("/reviews", c.ListReviews)
	full.GET("/reviews/:reviewID", c.GetReview)
	full.GET("/reviews/:reviewID/photos", c.GetReviewPhotos)

	// ScopeDormtrakWrite ensures that the person is a student and in the upperclasses
	writer := r.Group("")
	writer.Use(auth.RequireScopes(auth.ScopeDormtrakWrite))
	// Writing review operations
	writer.POST("/reviews", c.CreateReview)
	writer.PATCH("/reviews/:reviewID", c.UpdateReview)
	writer.DELETE("/reviews/:reviewID", c.DeleteReview)
	writer.PUT("/reviews/:reviewID/photo", c.UploadDormRoomPhoto)
}
