package main

import (
	"flag"
	"os"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	migrate "github.com/WilliamsStudentsOnline/wso-go/db"
	"github.com/WilliamsStudentsOnline/wso-go/jobs/user_home_info"
	"github.com/WilliamsStudentsOnline/wso-go/lib/logging"
	"github.com/WilliamsStudentsOnline/wso-go/models"
)

func main() {
	var configPath string
	var disableMigrationCheck bool
	var console bool
	var filePath string

	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.BoolVar(&disableMigrationCheck, "disable-migration-check", false, "don't check for outdated migrations")
	flag.BoolVar(&console, "console", false, "print logs in console as well as in file")
	flag.StringVar(&filePath, "file", "", "path to registrar WSO info TSV (EMAIL_ADDR, SUBOX, DORM, ENTRY, CITY, STATE, COUNTRY, POSTAL)")
	flag.Parse()

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		panic("Config Error: " + err.Error())
	}

	if console {
		cfg.LogFormats = append(cfg.LogFormats, "console")
	}

	log, err := logging.SetupLog(cfg, "user-home-info")
	if err != nil {
		panic("Log Setup Error: " + err.Error())
	}
	defer log.Sync()

	db := config.LoadDatabase(cfg, log)
	defer config.CloseDatabase(db, log)

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

	if filePath == "" {
		log.Fatal("TSV file path required (-file)")
	}

	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal("File Error: " + err.Error())
	}
	defer f.Close()

	records, err := user_home_info.ParseTSV(f)
	if err != nil {
		log.Fatal("Parse Error: " + err.Error())
	}
	log.Infof("parsed %d records from %s", len(records), filePath)

	userModel := models.NewUserModel(db, log)
	stats, err := user_home_info.IngestRecords(db, userModel, log, records)
	if err != nil {
		log.Fatal("Ingest Error: " + err.Error())
	}

	log.Infof("successfully ingested user home info: updated=%d not_found=%d off_campus=%d",
		stats.Updated, stats.NotFound, stats.OffCampus)
}
