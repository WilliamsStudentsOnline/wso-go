package user

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func SetupRouter(r gin.IRouter, db *gorm.DB) {
	c := NewController(db)
	r.GET("/", c.FetchAllUsers)
	r.GET("/:userID", c.GetUser)
	r.PUT("/:userID", c.UpdateUser)
	r.PUT("/:userID/tags", c.UpdateUserTags)
}
