package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	migrate "github.com/WilliamsStudentsOnline/wso-go/db"
	"github.com/WilliamsStudentsOnline/wso-go/server"
	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func main() {
	/* Flags */
	var configPath string
	var doDevelopment bool

	// Command-line flags
	// Note: these can be overridden by env vars
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.BoolVar(&doDevelopment, "development", false, "use development config")

	flag.Parse()

	if doDevelopment && configPath == "" {
		configPath = filepath.Join("config", "environment", "development.yaml")
	}

	/* Logging */
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	/* Config */
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatal("Config Error: " + err.Error())
		return
	}

	/* Set Logging Level */
	log.SetLevel(cfg.LogLevelParsed)

	/* DATABASE */
	db := config.LoadDatabase(cfg)
	defer config.CloseDatabase(db)

	/* Database Migrations */
	err = migrate.MigrateDB(db)
	if err != nil {
		log.Fatal("Migration Error: " + err.Error())
		return
	}

	/* Gin Mode */
	gin.SetMode(cfg.GinMode)

	/* Set Gin Print Route Func */
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		log.Debugf("endpoint %v %v %v %v\n", httpMethod, absolutePath, handlerName, nuHandlers)
	}

	/* Router */
	r, err := server.SetupRouter(cfg, db)
	if err != nil {
		log.Fatal(err)
		return
	}

	err = endless.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), r) // listen and serve on 0.0.0.0:8080
	if err != nil {
		log.Fatal("Server Error: " + err.Error())
		return
	}
}
