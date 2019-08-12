package server

import (
	"errors"
	"net/http"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	_ "github.com/WilliamsStudentsOnline/wso-go/docs"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	jwt "github.com/aidanlloydtucker/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// Services
	adminService "github.com/WilliamsStudentsOnline/wso-go/services/admin"
	authService "github.com/WilliamsStudentsOnline/wso-go/services/auth"
	bulletinService "github.com/WilliamsStudentsOnline/wso-go/services/bulletin"
	dormtrakService "github.com/WilliamsStudentsOnline/wso-go/services/dormtrak"
	factrakService "github.com/WilliamsStudentsOnline/wso-go/services/factrak"
	userService "github.com/WilliamsStudentsOnline/wso-go/services/user"
)

// @title WSO API
// @version 0.1.0
// @description API for WSO services like factrak, facebook, dormtrak, course scheduler, and others.

// @contact.name WSO Dev
// @contact.email wso-dev@wso.williams.edu

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization

func SetupRouter(cfg *config.Config, db *gorm.DB) (*gin.Engine, error) {
	r := gin.New()

	// Logger middleware will write the logs to gin.DefaultWriter even if you set with GIN_MODE=release.
	// By default gin.DefaultWriter = os.Stdout
	r.Use(config.Logger(log.StandardLogger()))

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	r.Use(gin.Recovery())

	// Build JWT auth middleware
	authMiddleware, err := authService.LoadAuthMiddleware(cfg, db)
	if err != nil {
		return nil, errors.New("JWT Error: " + err.Error())
	}

	/* ROUTER */

	// Initialize login
	r.POST("/api/v1/auth/login", authMiddleware.LoginHandler)

	// Run API docs if it is enabled
	if cfg.EnableAPIDocs {
		// Use ginSwagger middleware to serve the API docs.
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Require authentication for 404s
	r.NoRoute(authMiddleware.MiddlewareFunc(), func(c *gin.Context) {
		claims := jwt.ExtractClaims(c)
		log.Infof("NoRoute claims: %#v\n", claims)
		services.Base.RespondErrorCode(c, http.StatusNotFound, errors.New("page not found"))
	})

	// Wrap everything else in authentication
	router := r.Group("")
	router.Use(authMiddleware.MiddlewareFunc())

	// Actual API routing
	v1 := router.Group("/api/v1")
	{
		// Authentication for refresh user & other auth commands for already logged in users
		authService.SetupRouter(v1.Group("/auth"), authMiddleware)

		// User Service
		userGroup := v1.Group("/users")
		userGroup.Use(auth.RequireScopes(auth.ScopeUsers))
		userService.SetupRouter(userGroup, db, cfg)

		// Admin Service
		adminGroup := v1.Group("/admin")
		adminGroup.Use(auth.RequireScopes(auth.ScopeAdminAll))
		adminService.SetupRouter(adminGroup, db, cfg)

		// Factrak Service
		factrakGroup := v1.Group("/factrak")
		factrakGroup.Use(auth.RequireScopes(auth.ScopeFactrakLimited, auth.ScopeFactrakFull))
		factrakService.SetupRouter(factrakGroup, db, cfg)

		// Dormtrak Service
		dormtrakGroup := v1.Group("/dormtrak")
		dormtrakGroup.Use(auth.RequireScopes(auth.ScopeDormtrak))
		dormtrakService.SetupRouter(dormtrakGroup, db, cfg)

		// Bulletin Service
		bulletinGroup := v1.Group("/bulletin")
		bulletinGroup.Use(auth.RequireScopes(auth.ScopeBulletin))
		bulletinService.SetupRouter(dormtrakGroup, db, cfg)
	}

	return r, nil
}
