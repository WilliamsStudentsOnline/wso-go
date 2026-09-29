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
	gone := c.RespondGone

	r.GET("/neighborhoods", gone)
	r.GET("/neighborhoods/:neighborhoodID", gone)
	r.GET("/neighborhoods/:neighborhoodID/facts", gone)

	r.GET("/dorms", gone)
	r.GET("/dorms/:dormID", gone)
	r.GET("/dorms/:dormID/rooms", gone)
	r.GET("/dorms/:dormID/facts", gone)

	r.GET("/rooms/:roomID/photos", gone)

	r.GET("/rankings", gone)

	r.GET("/reviews", gone)
	r.GET("/reviews/:reviewID", gone)
	r.GET("/reviews/:reviewID/photos", gone)

	writer := r.Group("")
	writer.Use(auth.RequireScopes(auth.ScopeDormtrakWrite))
	writer.POST("/reviews", gone)
	writer.PATCH("/reviews/:reviewID", gone)
	writer.DELETE("/reviews/:reviewID", gone)
	writer.PUT("/reviews/:reviewID/photo", gone)
}
