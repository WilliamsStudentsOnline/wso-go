package auth

import (
	jwt "github.com/WilliamsStudentsOnline/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func SetupRouter(r gin.IRouter, authMiddleware *jwt.GinJWTMiddleware) {
	r.GET("/refresh-token", authMiddleware.RefreshHandler)
	r.GET("/update-token", authMiddleware.UpdateHandler)
}
