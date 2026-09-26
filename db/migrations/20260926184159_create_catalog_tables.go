package migrations

// File generated for Courses v2 catalog ingestion

import (
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

// Add the new variable to the migrations list at db/migrate.go
// If this creates a new table, or some other feature not automatically recorded in the model, add it to
// the InitSchema section of db/migrate.go

var CreateCatalogTables20260926184159 = &gormigrate.Migration{
	ID: "20260926184159_create_catalog_tables",
	Migrate: func(tx *gorm.DB) error {
		type CourseCanonical struct {
			models.BaseSchema
			CrseID      string `gorm:"unique;not null;size:32"`
			Title       string `gorm:"not null"`
			Description string `gorm:"size:65535"`
		}
		type CourseListing struct {
			models.BaseSchema
			CourseCanonicalID uint   `gorm:"index:idx_course_listings_course;unique_index:idx_course_listings_identity;not null"`
			Subject           string `gorm:"unique_index:idx_course_listings_identity;not null;size:16"`
			Number            int    `gorm:"unique_index:idx_course_listings_identity;not null"`
			Letter            string `gorm:"unique_index:idx_course_listings_identity;not null;size:8"`
			FirstValidYear    int    `gorm:"not null"`
			LastValidYear     int    `gorm:"not null"`
		}
		type Offering struct {
			models.BaseSchema
			CourseCanonicalID uint   `gorm:"index:idx_offerings_course;not null"`
			Strm              int    `gorm:"unique_index:idx_offerings_strm_class;not null"`
			Year              int    `gorm:"not null"`
			Term              string `gorm:"not null;size:16"`
			Section           string `gorm:"not null;size:16"`
			ClassNbr          int    `gorm:"unique_index:idx_offerings_strm_class;not null"`
			Title             string `gorm:"not null"`
			Component         string `gorm:"size:32"`
			Status            string `gorm:"not null;size:32"`
		}
		type OfferingMeeting struct {
			models.BaseSchema
			OfferingID uint   `gorm:"index:idx_offering_meetings_offering;not null"`
			Days       string `gorm:"size:16"`
			Start      string `gorm:"size:8"`
			End        string `gorm:"size:8"`
			Facility   string `gorm:"size:255"`
		}
		type OfferingInstructor struct {
			models.BaseSchema
			OfferingID   uint   `gorm:"index:idx_offering_instructors_offering;not null"`
			UnixID       string `gorm:"index:idx_offering_instructors_unix;size:100"`
			UserID       *uint  `gorm:"index:idx_offering_instructors_user"`
			NameAsListed string `gorm:"not null"`
		}
		if err := tx.AutoMigrate(
			&CourseCanonical{},
			&CourseListing{},
			&Offering{},
			&OfferingMeeting{},
			&OfferingInstructor{},
		).Error; err != nil {
			return err
		}
		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		if err := tx.DropTable("offering_instructors").Error; err != nil {
			return err
		}
		if err := tx.DropTable("offering_meetings").Error; err != nil {
			return err
		}
		if err := tx.DropTable("offerings").Error; err != nil {
			return err
		}
		if err := tx.DropTable("course_listings").Error; err != nil {
			return err
		}
		return tx.DropTable("courses_canonical").Error
	},
}
