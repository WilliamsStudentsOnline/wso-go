package main

import (
	"flag"
	"os"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	migrate "github.com/WilliamsStudentsOnline/wso-go/db"
	factrakbackfill "github.com/WilliamsStudentsOnline/wso-go/jobs/backfills/factrak"
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
		linkSourcesOnly       bool
	)

	flag.StringVar(&configPath, "config", "", "path to config file (required)")
	flag.StringVar(&reportPath, "report", "backfill_factrak-report.json", "where to write the JSON report")
	flag.BoolVar(&dryRun, "dry-run", true, "compute matches without writing (default true)")
	flag.BoolVar(&apply, "apply", false, "write course_canonical_id / course_offering_id (implies not dry-run)")
	flag.BoolVar(&skipLinked, "skip-linked", true, "skip surveys that already have course_canonical_id")
	flag.BoolVar(&fullResults, "full-results", false, "include every survey in the report (default: unmatched+ambiguous only)")
	flag.BoolVar(&runMigrations, "migrate", false, "run pending DB migrations first")
	flag.BoolVar(&disableMigrationCheck, "disable-migration-check", false, "don't require up-to-date migrations")
	flag.BoolVar(&console, "console", false, "also log to console")
	flag.BoolVar(&linkSourcesOnly, "link-sources-only", false, "only recompute link sources and My courses from trusted reviews (run after catalog loads; needs -apply); no matching")
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

	log, err := logging.SetupLog(cfg, "backfill_factrak")
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

	upToDate, err := migrate.MigrationUpToDate(db)
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

	if linkSourcesOnly {
		if dryRun {
			log.Info("dry run: -link-sources-only would recompute link sources and add My courses; pass -apply to write")
			return
		}
		if _, err := factrakbackfill.AssignCourseLinkVerificationSources(db, log); err != nil {
			log.Fatal(err)
		}
		if _, _, err := factrakbackfill.SyncReviewedUserCourses(db, log); err != nil {
			log.Fatal(err)
		}
		return
	}

	report, err := factrakbackfill.RunBackfill(db, log, factrakbackfill.Options{
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
	if err := factrakbackfill.WriteReportJSON(f, report); err != nil {
		log.Fatal(err)
	}
	log.Infof("Wrote report to %s", reportPath)
}
