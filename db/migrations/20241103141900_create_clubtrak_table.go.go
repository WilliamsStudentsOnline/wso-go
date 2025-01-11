package migrations

import (
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

var CreateClubTrakTable = &gormigrate.Migration{
	ID: "20241103141900_create_clubtrak_table",
	Migrate: func(tx *gorm.DB) error {
		// Use AutoMigrate to create the ClubTrak table if it doesn’t exist
		return tx.AutoMigrate(&models.Club{}).Error
	},
	Rollback: func(tx *gorm.DB) error {
		// Use DropTable directly, as there’s no Migrator in GORM v1
		return tx.DropTable("club_traks").Error // Ensure the table name matches your model’s table name
	},
}
