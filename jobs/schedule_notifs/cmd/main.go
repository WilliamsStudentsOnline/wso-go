package main

import (
	"flag"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	migrate "github.com/WilliamsStudentsOnline/wso-go/db"
	"github.com/WilliamsStudentsOnline/wso-go/jobs/schedule_notifs"
	"github.com/WilliamsStudentsOnline/wso-go/lib/logging"
)

func main() {
	/* Flags */
	var configPath string
	var disableMigrationCheck bool
	var console bool
	var notif string

	// Command-line flags
	// Note: these can be overridden by env vars
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.BoolVar(&disableMigrationCheck, "disable-migration-check", false, "don't check for outdated migrations")
	flag.BoolVar(&console, "console", false, "print logs in console as well as in ")
	flag.StringVar(&notif, "notif", "", "scheduled notification job to run")

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
	log, err := logging.SetupLog(cfg, "scheduled-notifications")
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

	var notifErr error
	switch notif {
	case "salmon":
		notifErr = schedule_notifs.SalmonNotify(cfg, db, log)
	case "food":
		notifErr = schedule_notifs.FoodNotify(cfg, db, log)
	default:
		log.Fatal("unknown scheduled notification job")
	}
	if notifErr != nil {
		log.Fatal("Notification Error: "+notifErr.Error(), notifErr)
	}

	log.Info("successfully ran scheduled notification")
}
