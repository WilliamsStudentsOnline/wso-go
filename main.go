package main

import (
	"errors"
	"flag"
	"net/http"
	"os"
	"path/filepath"

	"github.com/WilliamsStudentsOnline/wso-go/services/auth"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	migrate "github.com/WilliamsStudentsOnline/wso-go/db"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	adminService "github.com/WilliamsStudentsOnline/wso-go/services/admin"
	userService "github.com/WilliamsStudentsOnline/wso-go/services/user"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/fvbock/endless"
	log "github.com/sirupsen/logrus"
)

func main() {
	/* Flags */
	var env string
	var configPath string
	var secretsPath string

	// Command-line flags
	flag.StringVar(&env, "env", "development", "environment of server")
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.StringVar(&secretsPath, "secrets", filepath.Join("config", "secrets.yml"), "path to secrets file")

	flag.Parse()

	/* Logging */
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	/* Config */
	cfg, err := config.GetConfig(env, configPath)
	if err != nil {
		log.Fatal("Config Error: " + err.Error())
	}

	if cfg.IsProduction() {
		log.SetLevel(log.WarnLevel)
	} else if cfg.IsDevelopment() {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	/* Secrets */
	if _, err := os.Stat(secretsPath); !os.IsNotExist(err) {
		// If secrets file exists, parse it
		secrets, err := config.GetSecrets(secretsPath)
		if err != nil {
			log.Fatal("Secrets Error: " + err.Error())
		}
		cfg.Secrets = secrets
	} else if !cfg.IsProduction() {
		log.Warn("Secrets file not found; generating stubs based on defaults")
		// If secrets file does not exist, but we are not in production, stub the required secrets
		cfg.Secrets = &config.Secrets{
			JWTSecretKey: "wso-jwt-" + cfg.Env + "-secret",
		}
	} else {
		// If secrets file does not exist, and we are in production, fail
		log.Fatal("Secrets file must exist in production")
	}

	/* DATABASE */
	db := config.LoadDatabase(cfg)
	defer config.CloseDatabase(db)

	/* Database Migrations */
	err = migrate.MigrateDB(db)
	if err != nil {
		log.Fatal("Migration Error: " + err.Error())
	}

	/* Gin Mode */
	switch cfg.GinMode {
	case "development":
		gin.SetMode(gin.DebugMode)
	case "test":
		gin.SetMode(gin.TestMode)
	case "production":
		gin.SetMode(gin.ReleaseMode)
	default:
		gin.SetMode(gin.DebugMode)
	}

	/* Router */
	r, err := SetupRouter(cfg, db)
	if err != nil {
		log.Fatal(err)
	}

	// Would change this to be more production-friendly in real life. I'd use something like endless to keep
	// the server running even when it crashes
	err = endless.ListenAndServe(":"+cfg.Port, r) // listen and serve on 0.0.0.0:8080
	if err != nil {
		log.Fatal("Server Error: " + err.Error())
	}
}

func SetupRouter(cfg *config.Config, db *gorm.DB) (*gin.Engine, error) {
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		log.Debugf("endpoint %v %v %v %v\n", httpMethod, absolutePath, handlerName, nuHandlers)
	}

	r := gin.New()

	// Logger middleware will write the logs to gin.DefaultWriter even if you set with GIN_MODE=release.
	// By default gin.DefaultWriter = os.Stdout
	r.Use(config.Logger(log.StandardLogger()))

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	r.Use(gin.Recovery())

	// Build JWT auth middleware
	authMiddleware, err := auth.LoadAuthMiddleware(cfg, db)
	if err != nil {
		return nil, errors.New("JWT Error: " + err.Error())
	}

	/* ROUTER */

	// Initialize login
	r.POST("/api/v1/auth/login", authMiddleware.LoginHandler)

	// Require authentication for 404s
	r.NoRoute(authMiddleware.MiddlewareFunc(), func(c *gin.Context) {
		claims := jwt.ExtractClaims(c)
		log.Infof("NoRoute claims: %#v\n", claims)
		services.Base.RespondError(http.StatusNotFound, errors.New("page not found"), c)
	})

	// Wrap everything else in authentication
	router := r.Group("")
	router.Use(authMiddleware.MiddlewareFunc())

	// Middleware to set the user and user id with the context
	router.Use(func(c *gin.Context) {
		claims := jwt.ExtractClaims(c)
		userID := uint(claims["id"].(float64))
		user := models.NewUserWithID(userID)
		c.Set("user", &user)
		c.Set("user_id", userID)
		c.Next()
	})

	// Actual API routing
	v1 := router.Group("/api/v1")
	{
		// Authentication for refresh user & other auth commands for already logged in users
		auth.SetupRouter(v1.Group("/auth"), authMiddleware)

		// User API group
		userService.SetupRouter(v1.Group("/user"), db)

		// Admin API group
		adminGroup := v1.Group("/admin")
		adminGroup.Use(auth.RequireScopes(auth.ScopeAdminAll))
		adminService.SetupRouter(adminGroup, db, cfg)
	}

	return r, nil
}
