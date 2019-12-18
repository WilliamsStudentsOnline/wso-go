package server

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/docs"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	// Services
	adminService "github.com/WilliamsStudentsOnline/wso-go/services/admin"
	authService "github.com/WilliamsStudentsOnline/wso-go/services/auth"
	autocompleteService "github.com/WilliamsStudentsOnline/wso-go/services/autocomplete"
	bulletinService "github.com/WilliamsStudentsOnline/wso-go/services/bulletin"
	dormtrakService "github.com/WilliamsStudentsOnline/wso-go/services/dormtrak"
	ephcatchService "github.com/WilliamsStudentsOnline/wso-go/services/ephcatch"
	ephmatchService "github.com/WilliamsStudentsOnline/wso-go/services/ephmatch"
	factrakService "github.com/WilliamsStudentsOnline/wso-go/services/factrak"
	userService "github.com/WilliamsStudentsOnline/wso-go/services/user"
	wordsService "github.com/WilliamsStudentsOnline/wso-go/services/words"
)

// @title WSO API
// @version 2.0.0
// @description API for WSO services like factrak, facebook, dormtrak, course scheduler, and others.

// @contact.name WSO Dev
// @contact.email wso-dev@wso.williams.edu

// @host wso.williams.edu
// @BasePath /api/v2

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization

func SetupRouter(cfg *config.Config, db *gorm.DB, log *zap.SugaredLogger) (*gin.Engine, error) {
	docs.SwaggerInfo.Host = fmt.Sprintf("%s:%d", cfg.Hostname, cfg.Port)
	if cfg.EnableTLS {
		docs.SwaggerInfo.Schemes = []string{"https"}
	} else {
		docs.SwaggerInfo.Schemes = []string{"http"}
	}

	r := gin.New()

	// Logger middleware will write the logs to gin.DefaultWriter even if you set with GIN_MODE=release.
	// By default gin.DefaultWriter = os.Stdout
	r.Use(config.Logger(log))

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	// We also write to Slack if there is any internal server error
	r.Use(SlackRecovery(cfg, log))
	stdLog, err := zap.NewStdLogAt(log.Desugar(), zapcore.ErrorLevel)
	if err != nil {
		return nil, err
	}
	r.Use(gin.RecoveryWithWriter(io.MultiWriter(
		os.Stderr,
		stdLog.Writer(),
	)))

	// CORS config for react app
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")
	corsConfig.AllowAllOrigins = true
	r.Use(cors.New(corsConfig))

	// Build JWT auth middleware
	authMiddleware, err := authService.LoadAuthMiddleware(cfg, db, log.Named("auth"))
	if err != nil {
		return nil, errors.New("JWT Error: " + err.Error())
	}

	/* ROUTER */

	// Health check endpoint
	r.GET("/health-check", HealthCheck)

	// Initialize login
	r.POST("/api/v2/auth/login", authMiddleware.LoginHandler)

	// Initialize words endpoint
	wordsService.SetupRouter(r.Group("/api/v2/words"), db, cfg, log.Named("words"))

	// Run API docs if it is enabled
	// NOTE: This currently requires no JWT to access.
	if cfg.EnableAPIDocs {
		// Use ginSwagger middleware to serve the API docs.
		r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Require authentication for 404s
	r.NoRoute(authMiddleware.MiddlewareFunc(), func(c *gin.Context) {
		services.Base.RespondErrorCode(c, http.StatusNotFound, errors.New("page not found"))
	})

	// Wrap everything else in authentication
	router := r.Group("")
	router.Use(authMiddleware.MiddlewareFunc())

	// Actual API routing
	v2 := router.Group("/api/v2")
	{
		// Authentication for refresh user & other auth commands for already logged in users
		authService.SetupRouter(v2.Group("/auth"), authMiddleware)

		// User Service
		userGroup := v2.Group("/users")
		userGroup.Use(auth.RequireScopes(auth.ScopeUsers))
		userService.SetupRouter(userGroup, db, cfg, log.Named("users"))

		// Admin Service
		adminGroup := v2.Group("/admin")
		adminGroup.Use(auth.RequireScopes(auth.ScopeAdminAll))
		adminService.SetupRouter(adminGroup, db, cfg, log.Named("admin"))

		// Factrak Service
		factrakGroup := v2.Group("/factrak")
		factrakGroup.Use(auth.RequireScopes(auth.ScopeFactrakLimited, auth.ScopeFactrakFull))
		factrakService.SetupRouter(factrakGroup, db, cfg, log.Named("factrak"))

		// Dormtrak Service
		dormtrakGroup := v2.Group("/dormtrak")
		dormtrakGroup.Use(auth.RequireScopes(auth.ScopeDormtrak))
		dormtrakService.SetupRouter(dormtrakGroup, db, cfg, log.Named("dormtrak"))

		// Bulletin Service
		bulletinGroup := v2.Group("/bulletin")
		bulletinGroup.Use(auth.RequireScopes(auth.ScopeBulletin))
		bulletinService.SetupRouter(bulletinGroup, db, cfg, log.Named("bulletin"))

		// Ephcatch Service
		ephcatchGroup := v2.Group("/ephcatch")
		ephcatchGroup.Use(auth.RequireScopes(auth.ScopeEphcatch, auth.ScopeAdminAll))
		ephcatchService.SetupRouter(ephcatchGroup, db, cfg, log.Named("ephcatch"))

		// Autocomplete Service
		autocompleteGroup := v2.Group("/autocomplete")
		autocompleteService.SetupRouter(autocompleteGroup, db, cfg, log.Named("autocomplete"))

		// Ephmatch Service
		ephmatchGroup := v2.Group("/ephmatch")
		ephmatchGroup.Use(auth.RequireScopes(auth.ScopeEphmatch, auth.ScopeAdminAll))
		ephmatchService.SetupRouter(ephmatchGroup, db, cfg, log.Named("ephmatch"))
	}

	return r, nil
}

type HealthCheckResponse struct {
	OK bool `json:"ok"`
}

// HealthCheck godoc
// @Summary Health check
// @Description Check server health
// @ID health-check
// @Accept  json
// @Produce  json
// @Success 200 {object} server.HealthCheckResponse
// @Security Bearer
// @Router /health-check [get]
func HealthCheck(t *gin.Context) {
	t.JSON(http.StatusOK, HealthCheckResponse{OK: true})
}
