package main

import (
	"flag"
	"os"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	migrate "github.com/WilliamsStudentsOnline/wso-go/db"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	log "github.com/sirupsen/logrus"
)

func main() {
	/* Flags */
	var configPath string

	// Command-line flags
	// Note: these can be overridden by env vars
	flag.StringVar(&configPath, "config", "", "path to config file")

	flag.Parse()

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
	// NOTE: Job will not migrate anything; will fail if DB is not updated on migrations
	lastMigrationID, err := migrate.LastMigration(migrate.MigrationGormOptions, db)
	if err != nil {
		log.Fatal("Migration Checking Error: " + err.Error())
	}

	if lastMigrationID != migrate.Migrations[len(migrate.Migrations)-1].ID {
		log.Fatal("Database migrations are not up to date")
	}

	// Do the actual stuff
	userModel := models.NewUserModel(db)

	err = userModel.UpdateAllFromLDAP(cfg)
	if err != nil {
		log.Fatal("Update All From LDAP Error: " + err.Error())
	}
}
