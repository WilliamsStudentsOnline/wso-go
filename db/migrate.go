package db

import (
	"github.com/WilliamsStudentsOnline/wso-go/db/migrations"
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func MigrateDB(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		migrations.CreateUsers20190719211808,
		migrations.CreateDepartments20190719212645,
	})

	return m.Migrate()
}
