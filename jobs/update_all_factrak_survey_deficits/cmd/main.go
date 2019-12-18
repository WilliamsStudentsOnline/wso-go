package main

import (
	"flag"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	migrate "github.com/WilliamsStudentsOnline/wso-go/db"
	"github.com/WilliamsStudentsOnline/wso-go/lib/logging"
	"github.com/WilliamsStudentsOnline/wso-go/models"
)

func main() {
	/* Flags */
	var configPath string

	// Command-line flags
	// Note: these can be overridden by env vars
	flag.StringVar(&configPath, "config", "", "path to config file")

	flag.Parse()

	/* Config */
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		panic("Config Error: " + err.Error())
	}

	/* LOGGING */
	log, err := logging.SetupLog(cfg, "update-all-factrak-survey-deficits")
	if err != nil {
		panic("Log Setup Error: " + err.Error())
		return
	}
	defer log.Sync()

	/* DATABASE */
	db := config.LoadDatabase(cfg, log)
	defer config.CloseDatabase(db, log)

	/* Database Migrations */
	// NOTE: Job will not migrate anything; will fail if db is not updated on migrations
	dbUpToDate, err := migrate.MigrationUpToDate(migrate.MigrationGormOptions, db)
	if err != nil {
		log.Fatal("Migration Checking Error: " + err.Error())
	}

	if !dbUpToDate {
		log.Fatal("Database migrations are not up to date")
	}

	// Do the actual stuff
	studentModel := models.NewStudentModel(db, log)

	err = studentModel.UpdateAllFactrakSurveyDeficits()
	if err != nil {
		log.Fatal("Update All Factrak Survey Deficits Error: " + err.Error())
	}
}
