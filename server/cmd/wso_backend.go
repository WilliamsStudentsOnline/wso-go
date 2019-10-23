package main

import (
	"flag"
	"fmt"
	"io"
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

	/* LOGGING */
	log.SetLevel(cfg.LogLevelParsed)
	if cfg.LogPath != "" {
		logWriter, err := logSetup(cfg)
		if err != nil {
			log.Error(err)
			log.Warn("Using stdout as log output")
		} else {
			log.SetOutput(logWriter)
		}
	}

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

	addr := fmt.Sprintf(":%d", cfg.Port)

	if cfg.EnableTLS {
		err = endless.ListenAndServeTLS(addr, cfg.TLSCertPath, cfg.TLSKeyPath, r)
	} else {
		err = endless.ListenAndServe(addr, r)
	}
	if err != nil {
		log.Fatal("Server Error: " + err.Error())
		return
	}
}

func logSetup(cfg *config.Config) (io.Writer, error) {
	logPath, err := filepath.Abs(cfg.LogPath)
	if err != nil {
		return nil, err
	}
	return os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
}
