package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
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
	FixedFallSemesterYear = 2020
)

var (
	configPath            string
	filename              string
	factrakFilename       string
	archiveDir            string
	catalogDir            string
	savePreviousYears     int
	fromYear              int
	toYear                int
	draftCatalog          bool
	disableMigrationCheck bool
	runMigrations         bool
	ingestDB              bool
	noIngestDB            bool
	skipJSON              bool
	skipCurrent           bool
	rematchInstructors    bool
	console               bool

	log           *zap.SugaredLogger
	searchFactrak factrak.SearchFactrak
	db            *gorm.DB
)

func main() {
	var year int

	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.IntVar(&year, "year", 0, "the calendar year; set this to the year of spring semester")
	flag.IntVar(&savePreviousYears, "previous-years", 0, "parse n previous years and save to public JSONs")
	flag.IntVar(&fromYear, "from-year", 0, "inclusive start year for historical ingest (with -to-year); ingested ascending")
	flag.IntVar(&toYear, "to-year", 0, "inclusive end year for historical ingest (with -from-year)")
	flag.StringVar(&filename, "file", "courses.json", "where to save the courses JSON file")
	flag.StringVar(&factrakFilename, "factrak-file", "", "where to save the JSON aggregated with Factrak info")
	flag.StringVar(&archiveDir, "archive-dir", "", "directory to archive raw catalog JSON (default: <dir of -file>/catalog-archive)")
	flag.StringVar(&catalogDir, "catalog-dir", "", "load year_YYYY.json / raw-YYYY.json from this dir instead of the network")
	flag.BoolVar(&draftCatalog, "draft", false, "get the draft catalog at catalog.draft.williams.edu")
	flag.BoolVar(&disableMigrationCheck, "disable-migration-check", false, "don't check for outdated migrations")
	flag.BoolVar(&runMigrations, "migrate", false, "run pending DB migrations before ingest")
	flag.BoolVar(&ingestDB, "ingest-db", false, "upsert catalog rows into DB (default on when -config is set)")
	flag.BoolVar(&noIngestDB, "no-ingest-db", false, "skip DB upsert even when -config is set")
	flag.BoolVar(&skipJSON, "skip-json", false, "skip writing courses.json outputs (DB ingest / archive only)")
	flag.BoolVar(&skipCurrent, "skip-current", false, "skip the year/current fetch (useful with -from-year/-to-year)")
	flag.BoolVar(&rematchInstructors, "rematch-instructors", false, "re-link offering_instructors.user_id via unix_id then normalized name; can run alone")
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

	if archiveDir == "" {
		archiveDir = filepath.Join(filepath.Dir(filename), "catalog-archive")
	}

	historical := fromYear != 0 || toYear != 0
	if historical {
		if fromYear == 0 || toYear == 0 {
			panic("both -from-year and -to-year are required for historical load")
		}
		if fromYear > toYear {
			panic(fmt.Sprintf("-from-year (%d) must be <= -to-year (%d)", fromYear, toYear))
		}
		if fromYear < catalog.EarliestCatalogYear {
			panic(fmt.Sprintf("-from-year (%d) is before earliest served year %d", fromYear, catalog.EarliestCatalogYear))
		}
		// Historical loads should not also pull "current" unless explicitly wanted.
		skipCurrent = true
	}

	doIngest := false
	if configPath != "" && !noIngestDB {
		doIngest = true
	}
	if ingestDB {
		doIngest = true
	}
	if noIngestDB {
		doIngest = false
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
		db = config.LoadDatabase(cfg, log)
		factrakDB = db
		defer config.CloseDatabase(db, log)

		if runMigrations {
			if err := migrate.MigrateDB(db); err != nil {
				log.Fatal("Migration Error: " + err.Error())
			}
			log.Info("Database migrations applied")
		}

		/* Database Migrations */
		// NOTE: Job will not migrate anything unless -migrate; will fail if db is not updated on migrations
		dbUpToDate, err := migrate.MigrationUpToDate(migrate.MigrationGormOptions, db)
		if err != nil {
			log.Fatal("Migration Checking Error: " + err.Error())
		}

		if !dbUpToDate {
			if disableMigrationCheck {
				log.Warnf("Database migrations are not up to date")
			} else {
				log.Fatalf("Database migrations are not up to date (pass -migrate to apply)")
			}
		}

		searchFactrak = factrak.NewSearchFactrak(db, cfg, log)
	}

	if doIngest && db == nil {
		log.Fatal("-ingest-db requires -config so a database can be loaded")
	}
	if rematchInstructors && db == nil {
		log.Fatal("-rematch-instructors requires -config so a database can be loaded")
	}

	// Bare -rematch-instructors (no year range / previous-years): rematch and exit.
	if rematchInstructors && !historical && savePreviousYears == 0 {
		if _, err := catalog.RematchOfferingInstructors(db, log); err != nil {
			log.Fatal(err)
		}
		return
	}

	var currentCourses []catalog.Course

	if !skipCurrent {
		currentCourses, err = getCatalogCourses(0, doIngest)
		if err != nil {
			log.Fatal(err)
			return
		}

		if !skipJSON {
			err = writeCatalogFile(currentCourses, filename)
			if err != nil {
				log.Fatal(err)
				return
			}
		}

		// Save current year with Factrak reviews
		if factrakFilename != "" {
			f_factrak, err := os.OpenFile(factrakFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				log.Fatal(err)
				return
			}
			defer f_factrak.Close()
			err = catalog.SaveFactrakCatalog(f_factrak, currentCourses, factrakDB)
			if err != nil {
				log.Fatal(err)
				return
			}
			log.Infof("Successfully saved factrak-aggregated course catalog to %s", factrakFilename)
		}

		if len(currentCourses) == 0 {
			log.Fatal("current catalog produced zero courses")
			return
		}

		year = min(year, currentCourses[0].Year) // prevent trying to access future catalogs
	}

	if historical {
		log.Infof("Historical load: ingesting years %d–%d (ascending)", fromYear, toYear)
		for y := fromYear; y <= toYear; y++ {
			courses, err := getCatalogCourses(y, doIngest)
			if err != nil {
				log.Fatalf("Failed historical catalog year %d: %v", y, err)
			}
			if !skipJSON {
				out := strings.Replace(filename, ".json", fmt.Sprintf("-%v.json", y), 1)
				if err := writeCatalogFile(courses, out); err != nil {
					log.Fatalf("Failed to write catalog for %d: %v", y, err)
				}
			}
			if skipJSON {
				log.Infof("Historical year %d ingested", y)
			} else {
				log.Infof("Historical year %d done (%d parsed courses)", y, len(courses))
			}
		}
		log.Infof("Historical load complete (%d–%d)", fromYear, toYear)
		if rematchInstructors {
			if _, err := catalog.RematchOfferingInstructors(db, log); err != nil {
				log.Fatal(err)
			}
		}
		return
	}

	// Saves previous years to processed JSON (legacy -previous-years path)
	// Will write extra files for current year for compatibility (e.g. would save courses.json and courses-2025.json in AY 2024-5)
	for i := 0; i <= savePreviousYears; i++ {
		courses, err := getCatalogCourses(year-i, doIngest)
		if err != nil {
			log.Errorf("Failed to parse catalog: %v", err)
			continue
		}

		if skipJSON {
			continue
		}
		err = writeCatalogFile(courses, strings.Replace(filename, ".json", fmt.Sprintf("-%v.json", year-i), 1))
		if err != nil {
			log.Error("Failed to write catalog: %v", err)
			continue
		}
	}

}

func getCatalogCourses(year int, doIngest bool) (courses []catalog.Course, err error) {
	log.Infof("Fetching catalog for %v", year)

	var body []byte
	if catalogDir != "" {
		if year == 0 {
			return nil, fmt.Errorf("-catalog-dir cannot load year/current; pass an explicit year")
		}
		body, err = catalog.ReadCatalogYearFile(catalogDir, year)
		if err != nil {
			return
		}
		log.Infof("Loaded catalog for %v from %s", year, catalogDir)
	} else {
		body, err = catalog.FetchCatalogJSON(year, draftCatalog)
		if err != nil {
			return
		}
	}

	if err = catalog.ArchiveRawCatalog(archiveDir, year, body); err != nil {
		return nil, fmt.Errorf("archive raw catalog: %w", err)
	}
	log.Infof("Archived raw catalog for %v to %s", year, archiveDir)

	rawCourses, err := catalog.DecodeCatalog(body)
	if err != nil {
		return
	}

	if doIngest {
		if err = catalog.IngestCatalog(db, rawCourses, log); err != nil {
			return nil, fmt.Errorf("ingest catalog: %w", err)
		}
	}

	if skipJSON {
		// DB ingest / archive only — skip Factrak professor matching in ParseCatalog.
		return []catalog.Course{}, nil
	}

	if len(rawCourses) == 0 {
		return []catalog.Course{}, nil
	}

	yearForSemID := rawCourses[0].AcademicYear
	// Set the fall semester ID to be a linear scale (+10 every year) starting at a fixed point
	// (kept for JSON ParseCatalog compatibility; DB ingest uses TermFromStrm instead).
	fallSemesterID := FixedFallSemesterID + 10*(yearForSemID-FixedFallSemesterYear)
	winterSemesterID := fallSemesterID + 1
	springSemesterID := fallSemesterID + 2

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
