package bulletin

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func SetupRouter(r gin.IRouter, db *gorm.DB) {
	c := NewController(db)
	r.GET("/", c.FetchAllBulletins)
	r.GET("/:bulletinID", c.GetBulletin)
	r.PUT("/:bulletinID", c.UpdateBulletin)
	r.DELETE("/:bulletinID", c.DeleteBulletin)
}
