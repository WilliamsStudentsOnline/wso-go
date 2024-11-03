package migrations

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

var RenameClubTrakTable = &gormigrate.Migration{
	ID: "20241103143000_rename_clubtrak_table",
	Migrate: func(tx *gorm.DB) error {
		return tx.Exec("ALTER TABLE ClubTrak RENAME TO clubs").Error
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Exec("ALTER TABLE ClubTrak RENAME TO clubs").Error
	},
}
