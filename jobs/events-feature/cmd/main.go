package main

import (
	//"fmt"
	"flag"
	"log"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	events "github.com/WilliamsStudentsOnline/wso-go/jobs/events-feature"
	"github.com/WilliamsStudentsOnline/wso-go/lib/logging"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "path to config file")

	flag.Parse()

	rawCategories, err := events.GetRawCategories()
	if err != nil {
		log.Fatalf("error getting raw categories: %s\n", err)
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		panic("Config Error: " + err.Error())
	}

	/* LOGGING */
	logger, err := logging.SetupLog(cfg, "wso-backend")
	if err != nil {
		log.Fatalln("Log Setup Error: " + err.Error())
	}
	defer logger.Sync()

	dbLog := logger.Named("database")
	db := config.LoadDatabase(cfg, dbLog)
	defer db.Close()

	categories, err := events.ParseDailyMessages(rawCategories)
	if err != nil {
		log.Fatalf("error parsing daily messages: %s\n", err)
	}
	err = events.SaveToDB(categories, db)
	if err != nil {
		log.Fatalf("error saving to DB: %s\n", err)
	}
}
