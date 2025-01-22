package migrations

import (
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

var CreateClubtrakTable = &gormigrate.Migration{
	ID: "20250111153045_create_club",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&models.Club{}).Error
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.DropTable("clubs").Error
	},
}
