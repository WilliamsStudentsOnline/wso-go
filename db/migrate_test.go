package db

import (
	"testing"

	"gopkg.in/gormigrate.v1"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/assert"
)

func TestLastMigration(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	assert.NoError(t, err)

	db.LogMode(true)

	migrationID := "201608301430"

	// Copied this from documentation. Example stuff.
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		// create persons table
		{
			ID: "201608301400",
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
		},
		// add age column to persons
		{
			ID: "201608301415",
			Migrate: func(tx *gorm.DB) error {
				// when table already exists, it just adds fields as columns
				type Person struct {
					Age int
				}
				return tx.AutoMigrate(&Person{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Table("people").DropColumn("age").Error
			},
		},
		// add pets table
		{
			ID: migrationID,
			Migrate: func(tx *gorm.DB) error {
				type Pet struct {
					gorm.Model
					Name     string
					PersonID int
				}
				return tx.AutoMigrate(&Pet{}).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.DropTable("pets").Error
			},
		},
	})

	err = m.Migrate()
	assert.NoError(t, err)

	currMigrationID, err := LastMigration(gormigrate.DefaultOptions, db)
	assert.NoError(t, err)

	assert.Equal(t, migrationID, currMigrationID)
}
