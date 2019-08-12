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

	r.GET("/bulletins", c.ListBulletins)
	r.GET("/bulletins/:bulletinID", c.GetBulletin)
	writer.POST("/bulletins", c.CreateBulletin)
	writer.PATCH("/bulletins/:bulletinID", c.UpdateBulletin)
	writer.DELETE("/bulletins/:bulletinID", c.DeleteBulletin)

	r.GET("/rides", c.ListBulletins)
	r.GET("/rides/:rideID", c.GetBulletin)
	writer.POST("/rides", c.CreateBulletin)
	writer.PATCH("/rides/:ridesID", c.UpdateBulletin)
	writer.DELETE("/rides/:ridesID", c.DeleteBulletin)
}
