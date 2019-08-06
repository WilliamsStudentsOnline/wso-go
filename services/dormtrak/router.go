package dormtrak

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func SetupRouter(r gin.IRouter, db *gorm.DB) {
	c := NewController(db)

	// ALL: must be logged in, must have dormtrak auth

	// Models:
	// Neighborhoods, dorms, dorm rooms, dormtrak reviews

	// Neighborhoods endpoint
	r.GET("/neighborhoods", c.ListNeighborhoods)
	r.GET("/neighborhoods/:neighborhoodID", c.GetNeighborhood)

	r.GET("/dorms")
	r.GET("/dorms/:dormID")
	r.GET("/dorms/:dormID/rooms")

	// Get rankings by dormID, neighborhoodID, overall
	r.GET("/rankings")

	// Must be logged in:
	// Get reviews by dormID, neighborhoodID
	r.GET("/reviews")
	r.GET("/reviews/:reviewID")
	// Must be student, ensure_upperclassman, ensure_dorm, ensure_new_review
	r.POST("/reviews")
	//
	r.PATCH("/reviews/:reviewID")
	r.DELETE("/reviews/:reviewID")

}
