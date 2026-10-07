package old

import (
	"github.com/WilliamsStudentsOnline/wso-go/services/auth"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func SetupRouter(r gin.IRouter, authMiddleware *jwt.GinJWTMiddleware, db *gorm.DB) {
	r.GET("/refresh-token", authMiddleware.RefreshHandler)
	r.GET("/update-token", auth.UpdateHandler(authMiddleware, db))
}
