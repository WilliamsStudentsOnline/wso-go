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
			Type           string  `json:"type"`
			Name           string  `json:"name"`
			CellPhone      *string `json:"cellPhone"`
			CampusPhoneExt *string `json:"campusPhoneEXT"`
			UnixID         string  `gorm:"unique;not null;" json:"unixID"`
			WilliamsEmail  string  `json:"williamsEmail"`
			Title          *string `json:"title"`
			Visible        *bool   `gorm:"DEFAULT:true;not null" json:"visible"`
			ClassYear      *int    `gorm:"size:4" json:"classYear"`

			// Equivalent to belongs_to Department
			DepartmentID *uint `json:"departmentID"`

			DormVisible *bool   `gorm:"DEFAULT:true;not null" json:"dormVisible"`
			HomeTown    *string `json:"homeTown"`
			HomeZip     *string `json:"homeZip"`
			HomePhone   *string `json:"homePhone"`
			HomeState   *string `json:"homeState"`
			HomeCountry *string `json:"homeCountry"`
			HomeVisible *bool   `gorm:"DEFAULT:true;not null" json:"homeVisible"`

			Major                     *string `json:"major"`
			SUBox                     *string `json:"suBox"`
			Entry                     *string `json:"entry"`
			Admin                     *bool   `gorm:"DEFAULT:false;not null" json:"admin"`
			FactrakAdmin              *bool   `gorm:"DEFAULT:false;not null" json:"factrakAdmin"`
			HasAcceptedFactrakPolicy  *bool   `gorm:"DEFAULT:false;not null" json:"hasAcceptedFactrakPolicy"`
			HasAcceptedDormtrakPolicy *bool   `gorm:"DEFAULT:false;not null" json:"hasAcceptedDormtrakPolicy"`

			// belongs_to Office
			OfficeID *uint `json:"officeID"`

			// belongs_to Dorm Room
			DormRoomID *uint `gorm:"index:index_rooms_on_dorm_room_id" json:"dormRoomID"`

			Pronoun              *string `json:"pronoun"`
			AtWilliams           *bool   `gorm:"DEFAULT:true;not null" json:"atWilliams"`
			OffCycle             *bool   `gorm:"DEFAULT:false;not null" json:"offCycle"`
			FactrakSurveyDeficit *int    `json:"factrakSurveyDeficit"`

			OptOutEphcatch      *bool `gorm:"DEFAULT:false;not null" json:"optOutEphcatch"`
			EphcatchEligibility *bool `gorm:"DEFAULT:false;not null" json:"ephcatchEligibility"`

			// Keep this in here as long as we want to maintain this type of searching.
			SearchFields string `gorm:"default:'';not null'" json:"-"`
		}
		return tx.AutoMigrate(&User{}).Error
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.DropTable("users").Error
	},
}
