package bulletin

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// SetupRouter sets up the router for Bulletins
func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config) {
	c := NewController(db)

	writer := r.Group("")
	writer.Use(auth.RequireScopes(auth.ScopeWriteSelf))

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
}
