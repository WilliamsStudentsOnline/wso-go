package user

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func SetupRouter(r gin.IRouter, db *gorm.DB) {
	c := NewController(db)
	r.GET("/", c.ListUsers)
	r.GET("/:userID", c.GetUser)
	r.Use(auth.RequireScopes(auth.ScopeWriteSelf)).PUT("/:userID", c.UpdateUser)
	r.Use(auth.RequireScopes(auth.ScopeWriteSelf)).PUT("/:userID/tags", c.UpdateUserTags)
}
