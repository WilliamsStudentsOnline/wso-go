package db

import (
	"fmt"

	"github.com/WilliamsStudentsOnline/wso-go/db/migrations"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

// List all historical gormigrate migrations. Frozen: new schema changes go through
// Atlas SQL under db/atlas/migrations (see README / make atlas-diff).
// Still used to bootstrap SQLite test databases via InitSchema.
var Migrations = []*gormigrate.Migration{
	migrations.CreateUsers20190719211808,
	migrations.CreateDepartments20190719212645,
	migrations.CreateNeighborhoods20190721040940,
	migrations.CreateDorms20190721040956,
	migrations.CreateDormRooms20190721041007,
	migrations.CreateOffices20190721060106,
	migrations.CreateBulletins20190722202201,
	migrations.CreateTagsAndTagsUsers20190723012050,
	migrations.CreateAreasOfStudy20190729142313,
	migrations.CreateCourses20190729142528,
	migrations.CreateFactrakAgreements20190729142610,
	migrations.CreateFactrakSurveys20190729142626,
	migrations.CreateDormtrakReviews20190808235302,
	migrations.CreateEphcatches20190812033724,
	migrations.CreateBulletinRides20190812204834,
	migrations.CreateDiscussions20190812232511,
	migrations.CreatePosts20190812232529,
	migrations.AddNicknameToUsers20190826115256,
	migrations.CreateEphmatches20190926015114,
	migrations.CreateEphmatchProfiles20191124040019,
	migrations.MatchMessageColumn20200123225124,
	migrations.UserAddOffCampusColumn20200125225400,
	migrations.CreateEphmatchMatches20200324005606,
	migrations.CreateEphmatchLikes20200325175722,
	migrations.CurrentLocationColumns20200524014728,
	migrations.MessagingPlatformsColumns20200525010513,
	migrations.EphmatchMatchSeenColumns20200526221419,
	migrations.CampusStatusColumn20200713222414,
	migrations.FactrakSurveysCourseInfoColumns20210207234139,
	migrations.CreateNotificationSettings20210326205124,
	migrations.CreateNotificationTokens20210326205140,
	migrations.WilliamsIdColumn20210425233453,
	migrations.CreateGoodrichMenuItems20210426013057,
	migrations.CreateGoodrichOrders20210426013116,
	migrations.QuantityColumn20210507173142,
	migrations.UserAddOnCampusSemestersColumn20210406005136,
	migrations.CreateEphmatchRelations20210607211439,
	migrations.AddMhFactrakSurveyQ20211201012742,
	migrations.UserAddAreasOfStudy20220219171916,
	migrations.AddBannedUsersTable20220228163053,
	migrations.LookingForColumn20220504015544,
	migrations.CreateBookListings20221224071254,
	migrations.CreateBooks20221230032200,
}

var MigrationGormOptions = gormigrate.DefaultOptions

// MigrateDB applies schema migrations.
// MySQL uses Atlas-planned SQL (golang-migrate). SQLite (tests) uses gormigrate InitSchema.
func MigrateDB(db *gorm.DB) error {
	switch db.Dialect().GetName() {
	case "mysql":
		return MigrateMySQL(db)
	case "sqlite3":
		return migrateSQLite(db)
	default:
		return fmt.Errorf("unsupported dialect for migrations: %s", db.Dialect().GetName())
	}
}

func migrateSQLite(db *gorm.DB) error {
	m := gormigrate.New(db, MigrationGormOptions, Migrations)

	// Initializes the entire current schema. Used for the testing database.
	m.InitSchema(func(tx *gorm.DB) error {
		return AutoMigrateModels(tx)
	})

	return m.Migrate()
}

// Gets last migration id from the gormigrate migrations table (SQLite / legacy).
func LastMigration(opts *gormigrate.Options, db *gorm.DB) (string, error) {
	rows, err := db.Table(opts.TableName).Select(opts.IDColumnName).Rows()
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var migrationIDs []string

	for rows.Next() {
		migrationID := struct {
			ID string
		}{}
		err = db.ScanRows(rows, &migrationID)
		if err != nil {
			return "", err
		}

		migrationIDs = append(migrationIDs, migrationID.ID)
	}

	if len(migrationIDs) == 0 {
		return "", nil
	}

	return migrationIDs[len(migrationIDs)-1], nil
}

// MigrationUpToDate reports whether the DB has all migrations applied.
// MySQL checks Atlas/golang-migrate versions; SQLite checks the gormigrate table.
func MigrationUpToDate(opts *gormigrate.Options, db *gorm.DB) (bool, error) {
	if db.Dialect().GetName() == "mysql" {
		return MySQLMigrationsUpToDate(db)
	}
	return gormigrateUpToDate(opts, db)
}

func gormigrateUpToDate(opts *gormigrate.Options, db *gorm.DB) (bool, error) {
	var dbMigrationIDs []string

	err := db.Table(opts.TableName).Pluck(opts.IDColumnName, &dbMigrationIDs).Error
	if err != nil {
		return false, err
	}

	goMigrations := make([]string, len(Migrations)+1)
	for i := range Migrations {
		goMigrations[i] = Migrations[i].ID
	}
	goMigrations[len(goMigrations)-1] = "SCHEMA_INIT"

	for i := range dbMigrationIDs {
		if !lib.StringsContains(goMigrations, dbMigrationIDs[i]) {
			return false, nil
		}
	}

	for i := range goMigrations {
		if !lib.StringsContains(dbMigrationIDs, goMigrations[i]) {
			return false, nil
		}
	}

	return true, nil
}
