package migrations

import (
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

var CreateDepartments20190719212645 = &gormigrate.Migration{
	ID: "20190719212645_create_departments",
	Migrate: func(tx *gorm.DB) error {
		// It's a good pratice to copy the struct inside the function,
		// so side effects are prevented if the original struct changes during the time.
		// But, when the table already exists, it just adds new fields as columns, so just have a struct
		// with those fields.
		type Department struct {
			models.BaseSchema
			Name string
		}
		return tx.AutoMigrate(&Department{}).Error
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.DropTable("departments").Error
	},
}
