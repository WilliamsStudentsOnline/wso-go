package main

import (
	"flag"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	migrate "github.com/WilliamsStudentsOnline/wso-go/db"
	areasbackfill "github.com/WilliamsStudentsOnline/wso-go/jobs/backfills/areas"
	"github.com/WilliamsStudentsOnline/wso-go/lib/logging"
)

func main() {
	var (
		configPath            string
		apply                 bool
		disableMigrationCheck bool
		console               bool
	)

	flag.StringVar(&configPath, "config", "", "path to config file (required)")
	flag.BoolVar(&apply, "apply", false, "add the areas (default: dry run, only log what would be added)")
	flag.BoolVar(&disableMigrationCheck, "disable-migration-check", false, "don't require up-to-date migrations")
	flag.BoolVar(&console, "console", false, "also log to console")
	flag.Parse()

	if configPath == "" {
		panic("-config is required")
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		panic("Config Error: " + err.Error())
	}
	if console {
		cfg.LogFormats = append(cfg.LogFormats, "console")
	}

	log, err := logging.SetupLog(cfg, "backfill_areas")
	if err != nil {
		panic("Log Setup Error: " + err.Error())
	}
	defer log.Sync()

	db := config.LoadDatabase(cfg, log)
	defer config.CloseDatabase(db, log)

	upToDate, err := migrate.MigrationUpToDate(db)
	if err != nil {
		log.Fatal("Migration Checking Error: " + err.Error())
	}
	if !upToDate {
		if disableMigrationCheck {
			log.Warn("Database migrations are not up to date")
		} else {
			log.Fatal("Database migrations are not up to date")
		}
	}

	res, err := areasbackfill.Run(db, log, areasbackfill.Options{Apply: apply})
	if err != nil {
		log.Fatal("Areas backfill error: ", err)
	}
	verb := "would add"
	if apply {
		verb = "added"
	}
	log.Infof("areas backfill complete: %s=%v existing=%v missing-department=%v",
		verb, res.Created, res.Existing, res.MissingDepartment)
	if !apply && len(res.Created) > 0 {
		log.Info("dry run: pass -apply to write")
	}
}
