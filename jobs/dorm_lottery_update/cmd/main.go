package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	migrate "github.com/WilliamsStudentsOnline/wso-go/db"
	"github.com/WilliamsStudentsOnline/wso-go/jobs/dorm_lottery_update"
	"github.com/WilliamsStudentsOnline/wso-go/lib/logging"
)

func main() {
	/* Flags */
	var lotteryRecordsURL string
	var configPath string
	var filename string
	var disableMigrationCheck bool

	// Command-line flags
	// Note: these can be overridden by env vars
	flag.StringVar(&lotteryRecordsURL, "url", "https://drive.google.com/uc?export=download&id=1prfSIbLPp5tGgvLtCanWmyBrBhrTyuvU", "where to pull lottery records")
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.StringVar(&filename, "file", "dorm_lottery_beds.json", "where to save the dorm lottery beds JSON file")
	flag.BoolVar(&disableMigrationCheck, "disable-migration-check", false, "don't check for outdated migrations")

	flag.Parse()

	/* Config */
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		panic("Config Error: " + err.Error())
	}

	/* LOGGING */
	log, err := logging.SetupLog(cfg, "dorm-lottery-update")
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

	httpResp, err := http.Get(lotteryRecordsURL)
	if err != nil {
		log.Fatal("HTTP Get Records Error: " + err.Error())
	}

	respBytes, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		log.Fatal("HTTP Get Records Read Error: " + err.Error())
	}

	respReader := bytes.NewReader(respBytes)

	// Do the actual stuff
	rawBeds, err := dorm_lottery_update.ParseLotteryPDF(respReader, respReader.Size())
	if err != nil {
		log.Fatal("Parse Lottery PDF Error: " + err.Error())
	}

	beds, err := dorm_lottery_update.ParseLotteryDormBedRows(rawBeds, db)
	if err != nil {
		log.Fatal("Parse Lottery Dorm Bed Rows Error: " + err.Error())
	}

	// Open a file to save it as
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer f.Close()

	err = dorm_lottery_update.SaveLottery(f, beds)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Infof("Saved dorm lottery to %s", filename)
}
