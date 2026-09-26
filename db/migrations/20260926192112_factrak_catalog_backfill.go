package migrations

// File generated for Courses v2 Factrak backfill

import (
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

var FactrakCatalogBackfill20260926192112 = &gormigrate.Migration{
	ID: "20260926192112_factrak_catalog_backfill",
	Migrate: func(tx *gorm.DB) error {
		type CourseSubjectAlias struct {
			models.BaseSchema
			FactrakAbbrev  string `gorm:"unique;not null;size:16"`
			CatalogSubject string `gorm:"not null;size:16"`
		}
		type FactrakSurvey struct {
			CanonicalCourseID *uint `gorm:"index:index_factrak_surveys_on_canonical_course_id"`
			OfferingID        *uint `gorm:"index:index_factrak_surveys_on_offering_id"`
		}
		if err := tx.AutoMigrate(&CourseSubjectAlias{}, &FactrakSurvey{}).Error; err != nil {
			return err
		}

		// Seed high-confidence Factrak → catalog subject renames.
		defaults := []CourseSubjectAlias{
			{FactrakAbbrev: "WGST", CatalogSubject: "WGSS"},
			{FactrakAbbrev: "ASST", CatalogSubject: "ASIA"},
		}
		for _, a := range defaults {
			var existing CourseSubjectAlias
			err := tx.Where("factrak_abbrev = ?", a.FactrakAbbrev).First(&existing).Error
			if gorm.IsRecordNotFoundError(err) {
				if err := tx.Create(&a).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			}
		}
		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		if err := tx.Table("factrak_surveys").DropColumn("canonical_course_id").Error; err != nil {
			return err
		}
		if err := tx.Table("factrak_surveys").DropColumn("offering_id").Error; err != nil {
			return err
		}
		return tx.DropTable("course_subject_aliases").Error
	},
}
