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
	var pronounsFile string

	// Command-line flags
	// Note: these can be overridden by env vars
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.BoolVar(&disableMigrationCheck, "disable-migration-check", false, "don't check for outdated migrations")
	flag.BoolVar(&console, "console", false, "print logs in console as well as in ")
	flag.StringVar(&pronounsFile, "file", "", "path to pronoun info csv file")

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
	log, err := logging.SetupLog(cfg, "user-pronouns")
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

	if pronounsFile == "" {
		log.Fatal("Pronouns file required")
		return
	}

	// Do the actual stuff

	// Load pronouns file
	pnounsFile, err := os.Open(pronounsFile)
	if err != nil {
		log.Fatal("Pronouns File Error: " + err.Error())
		return
	}

	// CSV
	r := csv.NewReader(pnounsFile)

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("Pronouns Parse Error: " + err.Error())
			break
		}
		if len(record) != 2 {
			log.Warn("Pronouns Parse Warning: too many parsed elements in line")
			continue
		}

		err = db.Model(&models.User{}).Where("unix_id = ?", record[0]).Update("pronoun", record[1]).Error
		if err != nil {
			log.Fatal("User Database Error: " + err.Error())
			break
		}
	}

	log.Info("successfully updated all user pronouns")
}
