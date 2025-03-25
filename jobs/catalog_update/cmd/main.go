package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	migrate "github.com/WilliamsStudentsOnline/wso-go/db"
	catalog "github.com/WilliamsStudentsOnline/wso-go/jobs/catalog_update"
	"github.com/WilliamsStudentsOnline/wso-go/lib/logging"
	"github.com/WilliamsStudentsOnline/wso-go/lib/search/factrak"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

const (
	FixedFallSemesterID   = 1201
	FixedFallSemesterYear = 2019
)

var (
	configPath            string
	filename              string
	factrakFilename       string
	savePreviousYears     int
	draftCatalog          bool
	disableMigrationCheck bool
	console               bool

	log           *zap.SugaredLogger
	searchFactrak factrak.SearchFactrak
)

func main() {
	var year int

	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.IntVar(&year, "year", 0, "the calendar year; set this to the year of spring semester")
	flag.IntVar(&savePreviousYears, "previous-years", 0, "parse n previous years and save to public JSONs")
	flag.StringVar(&filename, "file", "courses.json", "where to save the courses JSON file")
	flag.StringVar(&factrakFilename, "factrak-file", "", "where to save the JSON aggregated with Factrak info")
	flag.BoolVar(&draftCatalog, "draft", false, "get the draft catalog at catalog.draft.williams.edu")
	flag.BoolVar(&disableMigrationCheck, "disable-migration-check", false, "don't check for outdated migrations")
	flag.BoolVar(&console, "console", false, "print logs in console as well as in ")

	flag.Parse()

	/* Flag Defaults */

	// Year defaults to spring of current academic year
	if year == 0 {
		year = time.Now().Year()
		if time.Now().Month() >= time.March { // load next year's catalog options in mar-aug, or this year's in sep-dec
			year += 1
		}
	}

	/* CONFIG */
	var cfg *config.Config
	var err error
	if configPath != "" {
		/* Config */
		cfg, err = config.LoadConfig(configPath)
		if err != nil {
			panic("Config Error: " + err.Error())
		}
	} else {
		cfg = &config.Config{
			LogLevel: "info",
		}
	}

	if console {
		cfg.LogFormats = append(cfg.LogFormats, "console")
	}

	/* LOGGING */
	log, err = logging.SetupLog(cfg, "catalog-update")
	if err != nil {
		panic("Log Setup Error: " + err.Error())
	}
	defer log.Sync()

	searchFactrak = nil
	var factrakDB *gorm.DB = nil

	/* LOAD DB AND FACTRAK SEARCH ENGINE IF ENABLED */

	// If no config path, don't load db or do search
	// If config path, load up the db and do the search for factrak profs
	if configPath != "" {
		/* DATABASE */
		db := config.LoadDatabase(cfg, log)
		factrakDB = db
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

		searchFactrak = factrak.NewSearchFactrak(db, cfg, log)
	}

	courses, err := getCatalogCourses(0) // 0 = current year
	if err != nil {
		log.Fatal(err)
		return
	}

	// Save current year
	err = writeCatalogFile(courses, filename)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Save current year with Factrak reviews
	f_factrak, err := os.OpenFile(factrakFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer f_factrak.Close()
	err = catalog.SaveFactrakCatalog(f_factrak, courses, factrakDB)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Infof("Successfully saved factrak-aggregated course catalog to %s", factrakFilename)

	// Saves previous years to processed JSON, optionally cross listings as well
	// Will write extra files for current year for compatibility (e.g. would save courses.json and courses-2025.json in AY 2024-5)
	for i := 0; i <= savePreviousYears; i++ {
		courses, err := getCatalogCourses(year - i)
		if err != nil {
			log.Fatal(err)
			return
		}

		err = writeCatalogFile(courses, strings.Replace(filename, ".json", fmt.Sprintf("-%v.json", year-i), 1))
		if err != nil {
			log.Fatal(err)
			return
		}
	}

}

func getCatalogCourses(year int) (courses []catalog.Course, err error) {
	// Set the fall semester ID to be a linear scale (+10 every year) starting at a fixed point
	fallSemesterID := FixedFallSemesterID + 10*(year-FixedFallSemesterYear)
	// Set winter semester ID to be one more than fall semester ID
	winterSemesterID := fallSemesterID + 1
	// Set spring semester ID to be two more than fall semester ID
	springSemesterID := fallSemesterID + 2

	rawCourses, err := catalog.GetCatalog(year, draftCatalog)
	if err != nil {
		return
	}

	courses, err = catalog.ParseCatalog(rawCourses, fallSemesterID, winterSemesterID, springSemesterID, log, searchFactrak)
	if err != nil {
		return
	}
	log.Infof("Successfully loaded catalog courses for %v", year)
	return
}

func writeCatalogFile(courses []catalog.Course, filename string) (err error) {
	f_course, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	defer f_course.Close()

	err = catalog.SaveCatalog(f_course, courses)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Infof("Successfully saved parsed catalog to %v", filename)
	return nil
}
