package main

import (
	"flag"
	"github.com/WilliamsStudentsOnline/wso-go/config"
	migrate "github.com/WilliamsStudentsOnline/wso-go/db"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
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
	if _, err := os.Stat(secretsPath); os.IsNotExist(err) {
		log.Fatal("Secrets file must exist")

	}

	// If secrets file exists, parse it
	secrets, err := config.GetSecrets(secretsPath)
	if err != nil {
		log.Fatal("Secrets Error: " + err.Error())
	}
	cfg.Secrets = secrets

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
