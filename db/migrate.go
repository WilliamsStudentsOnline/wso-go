package db

import (
	"fmt"

	"github.com/WilliamsStudentsOnline/wso-go/db/migrations"
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

// LEGACY gormigrate history (frozen; MySQL uses Atlas via MigrateMySQL)
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

func MigrateDB(db *gorm.DB) error {
	if db.Dialect().GetName() != "mysql" {
		return fmt.Errorf("unsupported dialect for migrations: %s (mysql only)", db.Dialect().GetName())
	}
	return MigrateMySQL(db)
}

// Check Atlas migration version (opts kept for call-site compatibility)
func MigrationUpToDate(opts *gormigrate.Options, db *gorm.DB) (bool, error) {
	_ = opts
	if db.Dialect().GetName() != "mysql" {
		return false, fmt.Errorf("unsupported dialect for MigrationUpToDate: %s", db.Dialect().GetName())
	}
	return MySQLMigrationsUpToDate(db)
}
