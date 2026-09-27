package main

import (
	"flag"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	migrate "github.com/WilliamsStudentsOnline/wso-go/db"
	boardbackfill "github.com/WilliamsStudentsOnline/wso-go/jobs/backfills/board"
	"github.com/WilliamsStudentsOnline/wso-go/lib/logging"
)

func main() {
	var configPath string
	var disableMigrationCheck bool
	var console bool

	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.BoolVar(&disableMigrationCheck, "disable-migration-check", false, "don't check for outdated migrations")
	flag.BoolVar(&console, "console", false, "print logs in console as well")
	flag.Parse()

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		panic("Config Error: " + err.Error())
	}
	if console {
		cfg.LogFormats = append(cfg.LogFormats, "console")
	}

	log, err := logging.SetupLog(cfg, "board-backfill")
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

	res, err := boardbackfill.Run(db, log)
	if err != nil {
		log.Fatal("Board backfill error: ", err)
	}
	log.Infof("board backfill complete: discussions=%d bulletins=%d rides=%d skipped=%d",
		res.DiscussionsMigrated, res.BulletinsMigrated, res.RidesMigrated, res.Skipped)
}
