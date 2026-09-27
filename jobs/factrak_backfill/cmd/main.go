package main

import (
	"flag"
	"os"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	migrate "github.com/WilliamsStudentsOnline/wso-go/db"
	"github.com/WilliamsStudentsOnline/wso-go/jobs/factrak_backfill"
	"github.com/WilliamsStudentsOnline/wso-go/lib/logging"
)

func main() {
	var (
		configPath            string
		reportPath            string
		dryRun                bool
		apply                 bool
		skipLinked            bool
		fullResults           bool
		runMigrations         bool
		disableMigrationCheck bool
		console               bool
	)

	flag.StringVar(&configPath, "config", "", "path to config file (required)")
	flag.StringVar(&reportPath, "report", "factrak-backfill-report.json", "where to write the JSON report")
	flag.BoolVar(&dryRun, "dry-run", true, "compute matches without writing (default true)")
	flag.BoolVar(&apply, "apply", false, "write canonical_course_id / offering_id (implies not dry-run)")
	flag.BoolVar(&skipLinked, "skip-linked", true, "skip surveys that already have canonical_course_id")
	flag.BoolVar(&fullResults, "full-results", false, "include every survey in the report (default: unmatched+ambiguous only)")
	flag.BoolVar(&runMigrations, "migrate", false, "run pending DB migrations first")
	flag.BoolVar(&disableMigrationCheck, "disable-migration-check", false, "don't require up-to-date migrations")
	flag.BoolVar(&console, "console", false, "also log to console")
	flag.Parse()

	if configPath == "" {
		panic("-config is required")
	}
	if apply {
		dryRun = false
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		panic("Config Error: " + err.Error())
	}
	if console {
		cfg.LogFormats = append(cfg.LogFormats, "console")
	}

	log, err := logging.SetupLog(cfg, "factrak-backfill")
	if err != nil {
		panic("Log Setup Error: " + err.Error())
	}
	defer log.Sync()

	db := config.LoadDatabase(cfg, log)
	defer config.CloseDatabase(db, log)

	if runMigrations {
		if err := migrate.MigrateDB(db); err != nil {
			log.Fatal("Migration Error: " + err.Error())
		}
		log.Info("Database migrations applied")
	}

	upToDate, err := migrate.MigrationUpToDate(migrate.MigrationGormOptions, db)
	if err != nil {
		log.Fatal("Migration Checking Error: " + err.Error())
	}
	if !upToDate {
		if disableMigrationCheck {
			log.Warn("Database migrations are not up to date")
		} else {
			log.Fatal("Database migrations are not up to date (pass -migrate)")
		}
	}

	report, err := factrak_backfill.RunBackfill(db, log, factrak_backfill.Options{
		DryRun:         dryRun,
		SkipLinked:     skipLinked,
		IncludeResults: fullResults,
	})
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.OpenFile(reportPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := factrak_backfill.WriteReportJSON(f, report); err != nil {
		log.Fatal(err)
	}
	log.Infof("Wrote report to %s", reportPath)
}
