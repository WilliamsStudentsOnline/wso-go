package migrations

import (
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

// Add the new variable to the migrations list at db/migrate.go
// If this creates a new table, or some other feature not automatically recorded in the model, add it to
// the InitSchema section of db/migrate.go

var CreateUsers20190719211808 = &gormigrate.Migration{
	ID: "20190719211808_create_users",
	Migrate: func(tx *gorm.DB) error {
		// It's a good pratice to copy the struct inside the function,
		// so side effects are prevented if the original struct changes during the time.
		// But, when the table already exists, it just adds new fields as columns, so just have a struct
		// with those fields.
		type User struct {
			models.BaseSchema
			Type           string
			Name           string
			CellPhone      *string
			CampusPhoneExt *string
			UnixID         string `gorm:"unique;"`
			WilliamsEmail  string
			Title          *string
			Visible        bool
			ClassYear      *int `gorm:"size:4"`

			// Equivalent to belongs_to Department
			DepartmentID *uint

			DormVisible bool `gorm:"DEFAULT:true"`
			HomeTown    *string
			HomeZip     *string
			HomePhone   *string
			HomeState   *string
			HomeCountry *string
			HomeVisible bool `gorm:"DEFAULT:true"`

			Major                     *string
			SUBox                     *string
			Entry                     *string
			Admin                     bool `gorm:"DEFAULT:false"`
			FactrakAdmin              bool `gorm:"DEFAULT:false"`
			HasAcceptedFactrakPolicy  bool `gorm:"DEFAULT:false"`
			HasAcceptedDormtrakPolicy bool `gorm:"DEFAULT:false"`

			OfficeID *uint

			DormRoomID *uint `gorm:"index_rooms_on_dorm_room_id"`

			Pronoun              *string
			AtWilliams           bool `gorm:"DEFAULT:true"`
			OffCycle             bool `gorm:"DEFAULT:false"`
			FactrakSurveyDeficit *int

			OptOutEphcatch      bool `gorm:"DEFAULT:false"`
			EphcatchEligibility bool `gorm:"DEFAULT:false"`
		}
		return tx.AutoMigrate(&User{}).Error
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.DropTable("users").Error
	},
}
