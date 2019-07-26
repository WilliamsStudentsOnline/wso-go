package bulletin

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// SetupRouter sets up the router for Bulletins
func SetupRouter(r gin.IRouter, db *gorm.DB) {
	c := NewController(db)
	r.GET("/", c.FetchAllBulletins)
	r.GET("/:bulletinID", c.GetBulletin)
	r.PUT("/:bulletinID", c.UpdateBulletin)
	r.POST("/bulletins", c.CreateBulletin)
	r.DELETE("/:bulletinID", c.DeleteBulletin)
}
