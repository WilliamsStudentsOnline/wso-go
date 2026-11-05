package seeder

import (
	"fmt"

	"github.com/WilliamsStudentsOnline/wso-go/db/seed"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
)

const defaultExtraUsers = 24

// KnownUnixIDs are stable local-login accounts when disable_ldap is true
// (any password works). Prefer these in docs / manual testing.
var KnownUnixIDs = []string{"student", "admin"}

// Users inserts a small synthetic user set. When n <= 0, seeds known accounts
// plus defaultExtraUsers random students.
func Users(n int, db *gorm.DB) error {
	extra := n
	if extra <= 0 {
		extra = defaultExtraUsers
	}

	if err := ensureKnownUsers(db); err != nil {
		return err
	}

	for i := 0; i < extra; i++ {
		u := seed.GenerateStudent()
		// Keep unix IDs clearly synthetic and unique.
		u.UnixID = fmt.Sprintf("seed%03d", i)
		u.WilliamsEmail = u.UnixID + "@williams.edu"
		u.Name = fmt.Sprintf("Seed User %03d", i)
		if err := db.Create(u).Error; err != nil {
			return err
		}
	}
	return nil
}

func ensureKnownUsers(db *gorm.DB) error {
	year := 2027

	student := &models.User{
		Type:                      models.UserTypeStudent,
		Name:                      "Seed Student",
		UnixID:                    "student",
		WilliamsEmail:             "student@williams.edu",
		Visible:                   lib.TruePtr(),
		ClassYear:                 &year,
		DormVisible:               lib.TruePtr(),
		HomeVisible:               lib.TruePtr(),
		Admin:                     lib.FalsePtr(),
		FactrakAdmin:              lib.FalsePtr(),
		HasAcceptedFactrakPolicy:  lib.TruePtr(),
		HasAcceptedDormtrakPolicy: lib.TruePtr(),
		AtWilliams:                lib.TruePtr(),
	}
	if err := createUserIfMissing(db, student); err != nil {
		return err
	}

	admin := &models.User{
		Type:                      models.UserTypeStaff,
		Name:                      "Seed Admin",
		UnixID:                    "admin",
		WilliamsEmail:             "admin@williams.edu",
		Visible:                   lib.TruePtr(),
		DormVisible:               lib.TruePtr(),
		HomeVisible:               lib.TruePtr(),
		Admin:                     lib.TruePtr(),
		FactrakAdmin:              lib.TruePtr(),
		HasAcceptedFactrakPolicy:  lib.TruePtr(),
		HasAcceptedDormtrakPolicy: lib.TruePtr(),
		AtWilliams:                lib.TruePtr(),
	}
	return createUserIfMissing(db, admin)
}

func createUserIfMissing(db *gorm.DB, u *models.User) error {
	var count int
	if err := db.Model(&models.User{}).Where("unix_id = ?", u.UnixID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return db.Create(u).Error
}

// UserCount returns the number of users in the database.
func UserCount(db *gorm.DB) (int, error) {
	var count int
	err := db.Model(&models.User{}).Count(&count).Error
	return count, err
}
