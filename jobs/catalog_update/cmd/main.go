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

// Course Scheduler's JSON numbers semesters linearly from Fall 2020 (+10 per year)
const (
	fixedFallSemesterID   = 1201
	fixedFallSemesterYear = 2020
)

type options struct {
	configPath, file, factrakFile, archiveDir, catalogDir, unitsDir string
	year, previousYears, fromYear, toYear                           int
	draft, migrate, disableMigrationCheck, ingestDB, noIngestDB     bool
	skipJSON, skipCurrent, rematchInstructors, units, console       bool
}

type job struct {
	opts     options
	log      *zap.SugaredLogger
	db       *gorm.DB
	factrak  factrak.SearchFactrak
	ingestDB bool
}

func main() {
	opts := parseFlags()
	j, closeDB := setup(opts)
	defer closeDB()
	defer j.log.Sync()

	if err := j.run(); err != nil {
		j.log.Fatal(err)
	}
}

func parseFlags() options {
	var o options
	flag.StringVar(&o.configPath, "config", "", "config file; enables the database steps")
	flag.IntVar(&o.year, "year", 0, "catalog year (spring calendar year); default is the current academic year")
	flag.IntVar(&o.previousYears, "previous-years", 0, "also write JSON for this many earlier years")
	flag.IntVar(&o.fromYear, "from-year", 0, "first catalog year of a historical load (with -to-year)")
	flag.IntVar(&o.toYear, "to-year", 0, "last catalog year of a historical load (with -from-year)")
	flag.StringVar(&o.file, "file", "courses.json", "Course Scheduler JSON output")
	flag.StringVar(&o.factrakFile, "factrak-file", "", "Course Scheduler JSON with Factrak scores (current year only)")
	flag.StringVar(&o.archiveDir, "archive-dir", "", "where to save raw catalog JSON (default: <dir of -file>/catalog-archive)")
	flag.StringVar(&o.catalogDir, "catalog-dir", "", "read raw-YYYY.json / year_YYYY.json from here instead of the network")
	flag.BoolVar(&o.draft, "draft", false, "use the draft catalog (catalog.draft.williams.edu)")
	flag.BoolVar(&o.migrate, "migrate", false, "apply pending database migrations first")
	flag.BoolVar(&o.disableMigrationCheck, "disable-migration-check", false, "run even if migrations are pending")
	flag.BoolVar(&o.ingestDB, "ingest-db", false, "upsert into the database (default when -config is set)")
	flag.BoolVar(&o.noIngestDB, "no-ingest-db", false, "don't upsert into the database even with -config")
	flag.BoolVar(&o.skipJSON, "skip-json", false, "don't write JSON files")
	flag.BoolVar(&o.skipCurrent, "skip-current", false, "skip step 1 (implied by -from-year/-to-year)")
	flag.BoolVar(&o.rematchInstructors, "rematch-instructors", false, "re-link catalog instructors to WSO users")
	flag.BoolVar(&o.units, "units", false, "update areas_of_study and course_attributes from the catalog's unit pages")
	flag.StringVar(&o.unitsDir, "units-dir", "", "read unit pages (index.html, <code>.html) from here instead of the network")
	flag.BoolVar(&o.console, "console", false, "also log to the console")
	flag.Parse()

	if o.year == 0 {
		// Next year's catalog is published in spring, so switch over in March
		o.year = time.Now().Year()
		if time.Now().Month() >= time.March {
			o.year++
		}
	}
	if o.archiveDir == "" {
		o.archiveDir = filepath.Join(filepath.Dir(o.file), "catalog-archive")
	}
	if o.historical() {
		if o.fromYear == 0 || o.toYear == 0 {
			panic("both -from-year and -to-year are required for a historical load")
		}
		if o.fromYear > o.toYear {
			panic(fmt.Sprintf("-from-year (%d) must be <= -to-year (%d)", o.fromYear, o.toYear))
		}
		if o.fromYear < catalog.EarliestCatalogYear {
			panic(fmt.Sprintf("-from-year (%d) is before the earliest catalog year, %d", o.fromYear, catalog.EarliestCatalogYear))
		}
		o.skipCurrent = true
	}
	return o
}

func (o options) historical() bool { return o.fromYear != 0 || o.toYear != 0 }

func setup(o options) (*job, func()) {
	cfg := &config.Config{LogLevel: "info"}
	if o.configPath != "" {
		var err error
		if cfg, err = config.LoadConfig(o.configPath); err != nil {
			panic("Config Error: " + err.Error())
		}
	}
	if o.console {
		cfg.LogFormats = append(cfg.LogFormats, "console")
	}
	log, err := logging.SetupLog(cfg, "catalog-update")
	if err != nil {
		panic("Log Setup Error: " + err.Error())
	}

	j := &job{opts: o, log: log, ingestDB: (o.configPath != "" || o.ingestDB) && !o.noIngestDB}
	if o.configPath == "" {
		if j.ingestDB || o.rematchInstructors || o.units {
			log.Fatal("-ingest-db, -rematch-instructors and -units need -config for the database")
		}
		return j, func() {}
	}

	j.db = config.LoadDatabase(cfg, log)
	if o.migrate {
		if err := migrate.MigrateDB(j.db); err != nil {
			log.Fatal("Migration Error: " + err.Error())
		}
		log.Info("Database migrations applied")
	}
	upToDate, err := migrate.MigrationUpToDate(j.db)
	if err != nil {
		log.Fatal("Migration Checking Error: " + err.Error())
	}
	if !upToDate {
		if !o.disableMigrationCheck {
			log.Fatal("Database migrations are not up to date (pass -migrate to apply)")
		}
		log.Warn("Database migrations are not up to date")
	}
	j.factrak = factrak.NewSearchFactrak(j.db, cfg, log)
	return j, func() { config.CloseDatabase(j.db, log) }
}

func (j *job) run() error {
	o := j.opts
	rematchOnly := o.rematchInstructors && !o.historical() && o.previousYears == 0
	if rematchOnly || (o.skipCurrent && !o.historical() && o.previousYears == 0) {
		return j.finish()
	}

	year := o.year
	if !o.skipCurrent {
		latest, err := j.loadCurrent()
		if err != nil {
			return err
		}
		year = min(year, latest) // don't ask for catalogs that don't exist yet
	}
	if o.historical() {
		if err := j.loadYears(o.fromYear, o.toYear); err != nil {
			return err
		}
	} else {
		j.writePreviousYears(year)
	}
	return j.finish()
}

func (j *job) finish() error {
	if j.opts.rematchInstructors {
		if _, err := catalog.RematchOfferingInstructors(j.db, j.log); err != nil {
			return err
		}
	}
	if j.opts.units {
		units, err := catalog.LoadCatalogUnits(j.opts.unitsDir, filepath.Join(j.opts.archiveDir, "units"), j.log)
		if err != nil {
			return err
		}
		if _, err := catalog.UpdateAreasAndAttributes(j.db, units, j.log); err != nil {
			return err
		}
	}
	return nil
}

func (j *job) loadCurrent() (int, error) {
	courses, err := j.loadYear(0)
	if err != nil {
		return 0, err
	}
	if len(courses) == 0 && !j.opts.skipJSON {
		return 0, fmt.Errorf("current catalog produced zero courses")
	}
	if !j.opts.skipJSON {
		if err := j.writeJSON(courses, j.opts.file); err != nil {
			return 0, err
		}
	}
	if j.opts.factrakFile != "" {
		f, err := os.OpenFile(j.opts.factrakFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return 0, err
		}
		defer f.Close()
		if err := catalog.SaveFactrakCatalog(f, courses, j.db); err != nil {
			return 0, err
		}
		j.log.Infof("Saved Factrak-scored catalog to %s", j.opts.factrakFile)
	}
	if len(courses) == 0 {
		return j.opts.year, nil
	}
	return courses[0].Year, nil
}

func (j *job) loadYears(from, to int) error {
	j.log.Infof("Historical load: catalog years %d–%d", from, to)
	for y := from; y <= to; y++ {
		courses, err := j.loadYear(y)
		if err != nil {
			return fmt.Errorf("catalog year %d: %w", y, err)
		}
		if !j.opts.skipJSON {
			if err := j.writeJSON(courses, yearFile(j.opts.file, y)); err != nil {
				return fmt.Errorf("write catalog year %d: %w", y, err)
			}
		}
		j.log.Infof("Catalog year %d loaded", y)
	}
	return nil
}

// Also writes the given year, so courses-YYYY.json exists alongside courses.json
func (j *job) writePreviousYears(year int) {
	for i := 0; i <= j.opts.previousYears; i++ {
		courses, err := j.loadYear(year - i)
		if err != nil {
			j.log.Errorf("Failed to load catalog year %d: %v", year-i, err)
			continue
		}
		if j.opts.skipJSON {
			continue
		}
		if err := j.writeJSON(courses, yearFile(j.opts.file, year-i)); err != nil {
			j.log.Errorf("Failed to write catalog year %d: %v", year-i, err)
		}
	}
}

func (j *job) loadYear(year int) ([]catalog.Course, error) {
	body, err := j.fetchRaw(year)
	if err != nil {
		return nil, err
	}
	if err := catalog.ArchiveRawCatalog(j.opts.archiveDir, year, body); err != nil {
		return nil, fmt.Errorf("archive raw catalog: %w", err)
	}
	raw, err := catalog.DecodeCatalog(body)
	if err != nil {
		return nil, err
	}
	if j.ingestDB {
		if err := catalog.IngestCatalog(j.db, raw, j.log); err != nil {
			return nil, fmt.Errorf("ingest catalog: %w", err)
		}
	}
	if j.opts.skipJSON || len(raw) == 0 {
		return []catalog.Course{}, nil
	}
	fall := fixedFallSemesterID + 10*(raw[0].AcademicYear-fixedFallSemesterYear)
	return catalog.ParseCatalog(raw, fall, fall+1, fall+2, j.log, j.factrak)
}

func (j *job) fetchRaw(year int) ([]byte, error) {
	if j.opts.catalogDir == "" {
		j.log.Infof("Fetching catalog year %d (0 = current)", year)
		return catalog.FetchCatalogJSON(year, j.opts.draft)
	}
	if year == 0 {
		return nil, fmt.Errorf("-catalog-dir can't load the current year; pass -from-year/-to-year")
	}
	j.log.Infof("Reading catalog year %d from %s", year, j.opts.catalogDir)
	return catalog.ReadCatalogYearFile(j.opts.catalogDir, year)
}

func (j *job) writeJSON(courses []catalog.Course, path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := catalog.SaveCatalog(f, courses); err != nil {
		return err
	}
	j.log.Infof("Saved catalog to %s", path)
	return nil
}

func yearFile(file string, year int) string {
	return strings.Replace(file, ".json", fmt.Sprintf("-%d.json", year), 1)
}
