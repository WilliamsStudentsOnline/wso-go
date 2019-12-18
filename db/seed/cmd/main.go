package main

import (
	"flag"
	"os"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	migrate "github.com/WilliamsStudentsOnline/wso-go/db"
	"github.com/WilliamsStudentsOnline/wso-go/db/seed/seeder"
	log "github.com/sirupsen/logrus"
)

func main() {
	/* Flags */
	var configPath string
	var model string
	var n int
	var disableMigrationCheck bool

	// Command-line flags
	// Note: these can be overridden by env vars
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.StringVar(&model, "model", "", "model (eg user, ephmatch, ephmatch_profile)")
	flag.IntVar(&n, "n", 0, "number of models to generate")
	flag.BoolVar(&disableMigrationCheck, "disable-migration-check", false, "don't check for outdated migrations")

	flag.Parse()

	/* Logging */
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	/* Config */
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatal("Config Error: " + err.Error())
		return
	}

	/* Set Logging Level */
	log.SetLevel(cfg.LogLevelParsed)

	/* DATABASE */
	db := config.LoadDatabase(cfg)
	defer config.CloseDatabase(db)

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

	switch model {
	case "ephmatch_profile":
		err = seeder.EphmatchProfiles(n, db)
	case "ephmatch":
		err = seeder.Ephmatches(n, db)
	default:
		log.Fatal("unknown model")
	}

	if err != nil {
		log.Fatal(err)
	}

	log.Info("Seeded!")
}
