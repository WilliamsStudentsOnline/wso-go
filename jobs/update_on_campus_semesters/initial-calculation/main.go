package main

import (
	"flag"
	"time"

	"go.uber.org/zap"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	migrate "github.com/WilliamsStudentsOnline/wso-go/db"
	"github.com/WilliamsStudentsOnline/wso-go/lib/logging"
	"github.com/WilliamsStudentsOnline/wso-go/models"
)

/*
This cmd file is used to generate the initial value of On Campus Semesters.
It is calculated from each student's class year and their created_at time in database.
Note that this should be run after database migration, which creates the column `OnCampusSemesters`.
*/
func main() {
	/* Flags */
	var configPath string
	var console bool

	// Command-line flags
	// Note: these can be overridden by env vars
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.BoolVar(&console, "console", false, "print logs in console as well as in ")

	flag.Parse()

	/* Config */
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		panic("Config Error: " + err.Error())
	}

	if console {
		cfg.LogFormats = append(cfg.LogFormats, "console")
	}

	/* LOGGING */
	log, err := logging.SetupLog(cfg, "update-on-campus-semesters")
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
		log.Fatal("Database migrations are not up to date")
	}

	// Do the actual stuff

	studentModel := models.NewStudentModel(db, log)

	err = InitializeOnCampusSemesters(studentModel, log)
	if err != nil {
		log.Fatal("Initialize On Campus Semesters Error: " + err.Error())
	}

	log.Info("successfully initialized on campus semesters")
}

// InitializeOnCampusSemesters sets/initializes all students' OnCampusSemesters value to their calculated value
func InitializeOnCampusSemesters(m *models.StudentModel, log *zap.SugaredLogger) (err error) {
	var students []models.User

	// We're fetching both those at_williams and not
	err = m.GetAllUsersByType(&students, models.UserTypeStudent)
	if err != nil {
		return
	}

	for _, student := range students {
		err = m.DB.Model(&student).Update("on_campus_semesters", CalculateOnCampusSemesters(&student, log)).Error
		if err != nil {
			return
		}
	}

	return
}

// CalculateOnCampusSemesters calculates a student's number of semesters on-campus according to
// their class year, created_at time and whether they are off cycle.
func CalculateOnCampusSemesters(u *models.User, log *zap.SugaredLogger) int {
	var OnCampusSemesters int
	s := u.Student()

	yearNumber := s.YearNumber()
	if *s.ClassYear-4 < s.CreatedAt.Year() {
		// possibly transfer student
		yearNumber += *s.ClassYear - 4 - s.CreatedAt.Year()

		// Log if we mark someone as transfer
		log.Debugf("[transfer student] %v (%v) '%v: On-Campus Year changed from %v to %v. Student created_at %v which should be %v \n",
			s.Name, s.UnixID, *s.ClassYear, s.YearNumber(), yearNumber, s.CreatedAt.Format("2006-Jan-2"), *s.ClassYear-4)
	} else if *s.ClassYear-4 > s.CreatedAt.Year() {
		// possibly student who took a gap year; do nothing and use current yearNumber
	}

	locTime := time.Now().Local()
	if locTime.Month() >= models.StudentCutoffMonth {
		// Fall Semester
		OnCampusSemesters = yearNumber*2 - 1
	} else {
		// Spring Semester
		OnCampusSemesters = yearNumber * 2
	}

	if *s.AtWilliams && s.Junior() {
		// Possibly on a junior year study-away program
		return (yearNumber - 1) * 2
	}

	if *s.OffCycle {
		// For those taking one semester off
		OnCampusSemesters--
	}

	return OnCampusSemesters
}
