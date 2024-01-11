package migrations

import (
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

var CreateEnrollments20240111120400 = &gormigrate.Migration{
	ID: "20240111120400_create_enrollments",
	Migrate: func(tx *gorm.DB) error {
		type Enrollment struct {
			models.BaseSchema

			Student   *models.User   `json:"student"`
			Course    *models.Course `json:"course"`
			StudentID *uint          `json:"studentID"`
			CourseID  *uint          `json:"courseID"`
			Hidden    bool           `json:"hidden"`

			SemesterID   *uint  `json:"semesterID"`
			SemesterType string `json:"semesterType"`
			Year         *uint  `json:"year"`
		}
		return tx.AutoMigrate(&Enrollment{}).Error
	},
	Rollback: func(tx *gorm.DB) error {
		err := tx.Table("enrollment").DropColumn("student").Error
		if err != nil {
			return err
		}
		err = tx.Table("enrollment").DropColumn("course").Error
		if err != nil {
			return err
		}
		err = tx.Table("enrollment").DropColumn("studentID").Error
		if err != nil {
			return err
		}
		err = tx.Table("enrollment").DropColumn("courseID").Error
		if err != nil {
			return err
		}
		err = tx.Table("enrollment").DropColumn("hidden").Error
		if err != nil {
			return err
		}
		err = tx.Table("enrollment").DropColumn("semesterID").Error
		if err != nil {
			return err
		}
		err = tx.Table("enrollment").DropColumn("semesterType").Error
		if err != nil {
			return err
		}
		err = tx.Table("enrollment").DropColumn("year").Error
		return err
	},
}
