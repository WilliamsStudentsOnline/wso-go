//go:build go1.8
// +build go1.8

package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	migrate "github.com/WilliamsStudentsOnline/wso-go/db"
	"github.com/WilliamsStudentsOnline/wso-go/lib/logging"
	"github.com/WilliamsStudentsOnline/wso-go/server"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	/* Flags */
	var configPath string
	var doDevelopment bool
	var beVerbose bool

	// Command-line flags
	// Note: these can be overridden by env vars
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.BoolVar(&doDevelopment, "development", false, "use development config")
	flag.BoolVar(&beVerbose, "verbose", false, "enable log output to both standard output and syslog (for debugging logging)")

	flag.Parse()

	if doDevelopment && configPath == "" {
		configPath = filepath.Join("config", "environment", "development.yaml")
	} else if beVerbose && configPath == "" {
		configPath = filepath.Join("config", "environment", "verbose.yaml")
	}

	/* Config */
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		panic("Config Error: " + err.Error())
		return
	}

	/* LOGGING */
	log, err := logging.SetupLog(cfg, "wso-backend")
	if err != nil {
		panic("Log Setup Error: " + err.Error())
		return
	}
	defer log.Sync()
	if beVerbose {
		log.Debug("note: logs will be extra verbose")
	}

	/* DATABASE */
	dbLog := log.Named("database")
	db := config.LoadDatabase(cfg, dbLog)
	defer config.CloseDatabase(db, dbLog)

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
		log.Debugf("endpoint %s %s %s %d", httpMethod, absolutePath, handlerName, nuHandlers)
	}

	/* Router */
	r, err := server.SetupRouter(cfg, db, log)
	if err != nil {
		log.Fatal(err)
		return
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	gracefulServe(addr, cfg, r, log)
}

func gracefulServe(addr string, cfg *config.Config, handler http.Handler, log *zap.SugaredLogger) {
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	if cfg.EnableTLS {
		go func() {
			// service connections
			if err := srv.ListenAndServeTLS(cfg.TLSCertPath, cfg.TLSKeyPath); err != nil && err != http.ErrServerClosed {
				log.Fatalf("listen and serve error: %s\n", err)
			}
		}()
	} else {
		go func() {
			// service connections
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("listen and serve error: %s\n", err)
			}
		}()
	}

	log.Infof("server running on %s", addr)

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Warn("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server shutdown error:", err)
	}
	// catching ctx.Done(). timeout of 5 seconds.
	select {
	case <-ctx.Done():
		log.Info("5 seconds timeout complete")
	}
	log.Info("server exited")
}
