package db

import (
	"github.com/WilliamsStudentsOnline/wso-go/db/migrations"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func MigrateDB(db *gorm.DB) error {
	// List all migrations here.
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		migrations.CreateUsers20190719211808,
		migrations.CreateDepartments20190719212645,
		migrations.CreateNeighborhoods20190721040940,
		migrations.CreateDorms20190721040956,
		migrations.CreateDormRooms20190721041007,
		migrations.CreateOffices20190721060106,
	})

	// This initializes the entire current schema with all migrations up to day.
	// Useful for starting the testing database.
	// NOTE: If you create a model or a feature (ie foreign key, etc.) that would not be automatically included,
	// please put it here for it to run.
	m.InitSchema(func(tx *gorm.DB) error {
		err := tx.AutoMigrate(
			// Put all current models here
			&models.User{},
			&models.Department{},
			&models.Neighborhood{},
			&models.Dorm{},
			&models.DormRoom{},
			&models.Office{},
			&models.Bulletin{},
		).Error
		if err != nil {
			return err
		}

		// all other foreign keys...
		return nil
	})

	return m.Migrate()
}
