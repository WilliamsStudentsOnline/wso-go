package migrations

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

// DONT EVER IMPORT THIS. THIS IS JUST AN EXAMPLE

var Example = &gormigrate.Migration{
	ID: "20180718173200_example",
	Migrate: func(tx *gorm.DB) error {
		// it's a good pratice to copy the struct inside the function,
		// so side effects are prevented if the original struct changes during the time
		type Person struct {
			gorm.Model
			Name string
		}
		return tx.AutoMigrate(&Person{}).Error
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.DropTable("people").Error
	},
}
