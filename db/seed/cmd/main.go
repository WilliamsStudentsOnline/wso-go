package main

import (
	"flag"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	migrate "github.com/WilliamsStudentsOnline/wso-go/db"
	"github.com/WilliamsStudentsOnline/wso-go/db/seed/seeder"
	"github.com/WilliamsStudentsOnline/wso-go/lib/logging"
)

func main() {
	var configPath string
	var model string
	var n int
	var disableMigrationCheck bool

	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.StringVar(&model, "model", "all", "model: all | user | board | ephmatch | ephmatch_profile | dorm")
	flag.IntVar(&n, "n", 0, "number of models to generate (0 = seeder default)")
	flag.BoolVar(&disableMigrationCheck, "disable-migration-check", false, "deprecated: migrate always runs before seed")
	flag.Parse()
	_ = disableMigrationCheck

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		panic("Config Error: " + err.Error())
	}

	log, err := logging.SetupLog(cfg, "seed")
	if err != nil {
		panic("Log Setup Error: " + err.Error())
	}
	defer log.Sync()

	db := config.LoadDatabase(cfg, log)
	defer config.CloseDatabase(db, log)

	// Schema first, then data — same order as devenv (empty → migrate → seed).
	if err := migrate.MigrateDB(db); err != nil {
		log.Fatal("Migration Error: " + err.Error())
	}

	switch model {
	case "all", "":
		err = seeder.Development(db)
	case "user", "users":
		err = seeder.Users(n, db)
	case "board":
		err = seeder.BoardThreads(n, db)
	case "dorm", "dorms":
		err = seeder.Dorms(n, db)
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
