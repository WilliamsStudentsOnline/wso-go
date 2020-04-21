package main

import (
	"flag"
	"os"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	catalog "github.com/WilliamsStudentsOnline/wso-go/jobs/catalog_update"
	"github.com/WilliamsStudentsOnline/wso-go/lib/logging"
)

const (
	FixedFallSemesterID   = 1201
	FixedFallSemesterYear = 2019
)

func main() {
	var configPath string
	var year int
	var academicYear int
	var fallSemesterID int
	var winterSemesterID int
	var springSemesterID int
	var filename string
	var draftCatalog bool

	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.IntVar(&year, "year", 0, "the calendar year; set this to the year of fall semester")
	flag.IntVar(&academicYear, "academic-year", 0, "academic year of courses (eg 1819, 1920)")
	flag.IntVar(&fallSemesterID, "fall", 0, "fall courses semester id")
	flag.IntVar(&winterSemesterID, "winter", 0, "winter courses semester id")
	flag.IntVar(&springSemesterID, "spring", 0, "spring courses semester id")
	flag.StringVar(&filename, "file", "courses.json", "where to save the courses JSON file")
	flag.BoolVar(&draftCatalog, "draft", false, "get the draft catalog at catalog.draft.williams.edu")

	flag.Parse()

	/* Flag Defaults */

	// Default year is now, but if it is march or before (early-mid 2nd semester), set it to the previous year
	if year == 0 {
		year = time.Now().Year()
		if time.Now().Month() <= time.March {
			year = year - 1
		}
	}

	// Set the academic year from the last 2 digits of the year and the last 2 digits of the next year
	if academicYear == 0 {
		// Converts a real year's 2018 to 1819 (aabb to bb(bb+1))
		academicYear = (year%100)*100 + (year % 100) + 1
	}

	// Set the fall semester ID to be a linear scale (+10 every year) starting at a fixed point
	if fallSemesterID == 0 {
		fallSemesterID = FixedFallSemesterID + 10*(year-FixedFallSemesterYear)
	}
	// Set winter semester ID to be one more than fall semester ID
	if winterSemesterID == 0 {
		winterSemesterID = fallSemesterID + 1
	}
	// Set spring semester ID to be two more than fall semester ID
	if springSemesterID == 0 {
		springSemesterID = fallSemesterID + 2
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

	/* LOGGING */
	log, err := logging.SetupLog(cfg, "catalog-update")
	if err != nil {
		panic("Log Setup Error: " + err.Error())
		return
	}
	defer log.Sync()

	/* Command Code */

	rawCourses, err := catalog.GetCatalog(academicYear, draftCatalog)
	if err != nil {
		log.Fatal(err)
		return
	}

	courses, err := catalog.ParseCatalog(rawCourses, fallSemesterID, winterSemesterID, springSemesterID)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Open a file to save it as
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer f.Close()

	err = catalog.SaveCatalog(f, courses)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Infof("Saved parsed course catalog to %s", filename)
}
