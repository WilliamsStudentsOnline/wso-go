package bulletin

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func SetupRouter(r gin.IRouter, db *gorm.DB) {
	c := NewController(db)
	r.GET("/", c.FetchAllBulletins)
	r.GET("/:bulletin_id", c.GetBulletin)
	r.PUT("/:bulletin_id", c.UpdateBulletin)
	r.DELETE("/:bulletin_id", c.DeleteBulletin)
}
