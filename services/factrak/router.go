package factrak

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func SetupRouter(r gin.IRouter, db *gorm.DB) {
	c := NewController(db)
	// Example route:
	/*
	r.GET("/", c.FetchAllUsers)
	*/
}
