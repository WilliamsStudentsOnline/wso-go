package main

import (
	"encoding/csv"
	"flag"
	"io"
	"os"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	migrate "github.com/WilliamsStudentsOnline/wso-go/db"
	"github.com/WilliamsStudentsOnline/wso-go/lib/logging"
	"github.com/WilliamsStudentsOnline/wso-go/models"
)

func main() {
	/* Flags */
	var configPath string
	var disableMigrationCheck bool
	var console bool
	var csvFilePath string

	// Command-line flags
	// Note: these can be overridden by env vars
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.BoolVar(&disableMigrationCheck, "disable-migration-check", false, "don't check for outdated migrations")
	flag.BoolVar(&console, "console", false, "print logs in console as well as in ")
	flag.StringVar(&csvFilePath, "file", "", "path to user data csv file")

	flag.Parse()

	/* Config */
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		panic("Config Error: " + err.Error())
	}

	if console {
		cfg.LogFormats = append(cfg.LogFormats, "console")
	}

	/* LOGGING */
	log, err := logging.SetupLog(cfg, "user-csv-data")
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
		if disableMigrationCheck {
			log.Warnf("Database migrations are not up to date")
		} else {
			log.Fatalf("Database migrations are not up to date")
		}
	}

	if csvFilePath == "" {
		log.Fatal("CSV file path required")
		return
	}

	// Do the actual stuff

	// Load CSV file
	csvFile, err := os.Open(csvFilePath)
	if err != nil {
		log.Fatal("CSV File Error: " + err.Error())
		return
	}

	// CSV
	r := csv.NewReader(csvFile)

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("CSV Parse Error: " + err.Error())
			break
		}
		if len(record) != 3 {
			log.Warn("CSV Parse Warning: too many parsed elements in line")
			continue
		}

		err = db.Model(&models.User{}).Where("unix_id = ?", record[0]).Update("pronoun", record[1], "off_cycle", record[2] == "T").Error
		if err != nil {
			log.Fatal("User Database Error: " + err.Error())
			break
		}
	}

	log.Info("successfully updated all user data from CSV")
}
