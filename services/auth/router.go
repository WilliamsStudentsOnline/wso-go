package auth

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func SetupRouter(r gin.IRouter, authMiddleware *jwt.GinJWTMiddleware) {
	r.GET("/refresh_token", authMiddleware.RefreshHandler)
}
