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
	authAPIService "github.com/WilliamsStudentsOnline/wso-go/services/auth/api"
	authIdentService "github.com/WilliamsStudentsOnline/wso-go/services/auth/identity"
	authOldService "github.com/WilliamsStudentsOnline/wso-go/services/auth/old"
	autocompleteService "github.com/WilliamsStudentsOnline/wso-go/services/autocomplete"

	booktrakService "github.com/WilliamsStudentsOnline/wso-go/services/booktrak"
	bulletinService "github.com/WilliamsStudentsOnline/wso-go/services/bulletin"
	bulletinRSSService "github.com/WilliamsStudentsOnline/wso-go/services/bulletin/rss"
	chatService "github.com/WilliamsStudentsOnline/wso-go/services/chat"
	clubtrakService "github.com/WilliamsStudentsOnline/wso-go/services/clubtrak"
	dormtrakService "github.com/WilliamsStudentsOnline/wso-go/services/dormtrak"
	ephcatchService "github.com/WilliamsStudentsOnline/wso-go/services/ephcatch"
	ephmatchService "github.com/WilliamsStudentsOnline/wso-go/services/ephmatch"
	factrakService "github.com/WilliamsStudentsOnline/wso-go/services/factrak"
	goodrichService "github.com/WilliamsStudentsOnline/wso-go/services/goodrich"
	notificationService "github.com/WilliamsStudentsOnline/wso-go/services/notification"
	onboardingService "github.com/WilliamsStudentsOnline/wso-go/services/onboarding"
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

	/* Initialize Authentication Middlewares  */

	// Old authentication middleware for backwards compatibility
	authOldMiddleware, err := authOldService.LoadAuthMiddleware(cfg, db, log.Named("auth.old"))
	if err != nil {
		return nil, errors.New("JWT Error: " + err.Error())
	}

	// New authentication middleware

	// This runs the identification authentication, which is involved with long-term tokens
	authIdentMiddleware, err := authIdentService.LoadMiddleware(cfg, db, log.Named("auth.ident"))
	if err != nil {
		return nil, errors.New("JWT Error: " + err.Error())
	}

	// This is the standard authentication needed to access the API
	authAPIMiddleware, err := authAPIService.LoadMiddleware(cfg, db, log.Named("auth.api"))
	if err != nil {
		return nil, errors.New("JWT Error: " + err.Error())
	}

	/* ROUTER */

	// Health check endpoint
	r.GET("/health-check", HealthCheck)

	// Token creation for identity tokens
	r.POST("/api/v2/auth/identity/token", authIdentMiddleware.LoginHandler)

	// Token creation for API tokens (requires identity token authentication)
	r.POST("/api/v2/auth/api/token", authIdentMiddleware.MiddlewareFunc(), authAPIMiddleware.LoginHandler)

	// Allow API tokens to be refreshed (updated)
	r.GET("/api/v2/auth/api/refresh", authAPIMiddleware.UpdateHandler)

	// Token creation for old authentication (backwards compatible)
	r.POST("/api/v2/auth/login", authOldMiddleware.LoginHandler)

	// Initialize words endpoint
	wordsService.SetupRouter(r.Group("/api/v2/words"), db, cfg, log.Named("words"))

	// Run API docs if it is enabled
	// NOTE: This currently requires no JWT to access.
	if cfg.EnableAPIDocs {
		// Use ginSwagger middleware to serve the API docs.
		r.GET("/docs/*any", func(c *gin.Context) {
			if c.Param("any") == "/" || c.Param("any") == "" {
				c.Redirect(http.StatusMovedPermanently, "/docs/index.html")
				c.Abort()
			}
		}, ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Require authentication for 404s
	r.NoRoute(authAPIMiddleware.MiddlewareFunc(), func(c *gin.Context) {
		services.Base.RespondErrorCode(c, http.StatusNotFound, errors.New("page not found"))
	})

	// Wrap everything else in authentication. We use the new API authentication middleware, as parsing tokens is
	//backwards compatible.
	router := r.Group("")
	router.Use(authAPIMiddleware.MiddlewareFunc())

	// Actual API routing
	v2 := router.Group("/api/v2")
	{
		// Authentication for refresh user & other auth commands for already logged in users
		// THIS IS FOR OLD DEPRECATED AUTHENTICATION
		authOldService.SetupRouter(v2.Group("/auth"), authOldMiddleware)

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

		// Booktrak Service
		booktrakGroup := v2.Group("/booktrak")
		booktrakGroup.Use(auth.RequireScopes(auth.ScopeBooktrak))
		booktrakService.SetupRouter(booktrakGroup, db, cfg, log.Named("booktrak"))

		//Clubtrak Service
		clubtrakGroup := v2.Group("/clubtrak")
		//clubtrakGroup.Use(auth.RequireScopes(auth.ScopeUsers))
		clubtrakService.SetupRouter(clubtrakGroup, db, cfg, log.Named("clubtrak"))

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
		ephmatchService.SetupRouter(ephmatchGroup, db, cfg, log.Named("ephmatch"))

		// Bulletin RSS Service
		bulletinRSSGroup := r.Group("/api/v2/bulletin/rss")
		bulletinRSSService.SetupRouter(bulletinRSSGroup, db, cfg, log.Named("bulletin").Named("rss"))

		// Chat Service
		chatGroup := v2.Group("/chat")
		chatGroup.Use(auth.RequireScopes(auth.ScopeChat))
		chatService.SetupRouter(chatGroup, db, cfg, log.Named("chat"))

		// Onboarding Service
		onboardingGroup := v2.Group("/onboarding")
		onboardingService.SetupRouter(onboardingGroup, db, cfg, log.Named("onboarding"))

		// Notifications Service
		notifGroup := v2.Group("/notification")
		notifGroup.Use(auth.RequireScopes(auth.ScopeUsers))
		notificationService.SetupRouter(notifGroup, db, cfg, log.Named("notification"))

		// Goodrich Service
		if !cfg.MissingGoodrich() {
			goodrichGroup := v2.Group("/goodrich")
			goodrichGroup.Use(auth.RequireScopes(auth.ScopeGoodrich))
			goodrichService.SetupRouter(goodrichGroup, db, cfg, log.Named("goodrich"))
		} else {
			log.Warn("Goodrich Config is not set up. Will not route Goodrich Service. ")
		}
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
