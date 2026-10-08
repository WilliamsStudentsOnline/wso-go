package models

import (
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/jinzhu/gorm"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func setupUserSearchFieldsDB(t *testing.T) (*gorm.DB, *zap.SugaredLogger) {
	t.Helper()

	cfg := &config.Config{
		Env:          "test",
		GinMode:      "test",
		JWTRealm:     "wso-go-test",
		DatabaseType: "sqlite3",
		DatabaseArgs: ":memory:",
		Secrets: &config.Secrets{
			JWTSecretKey: "wso-jwt-test-secret",
		},
	}

	log := zaptest.NewLogger(t).Sugar()
	db := config.LoadDatabase(cfg, log)
	err := db.AutoMigrate(
		User{},
		Tag{},
		Neighborhood{},
		Dorm{},
		DormRoom{},
		Office{},
		FactrakSurvey{},
		FactrakAgreement{},
	).Error
	testify.NoError(t, err)
	return db, log
}

// LDAP sync loads users with foreign key IDs but without associations populated.
// generateSearchFieldsByUser must look those up by dereferenced IDs (gorm rejects *uint).
func TestUserModel_generateSearchFieldsByUser_loadsAssociationsByID(t *testing.T) {
	db, log := setupUserSearchFieldsDB(t)

	office := &Office{Number: "Schow 12"}
	testify.NoError(t, db.Create(office).Error)

	dormRoom := &DormRoom{
		Dorm: &Dorm{
			Neighborhood: &Neighborhood{Name: "Currier"},
			Name:         "East",
		},
		Number: "103",
	}
	testify.NoError(t, db.Create(dormRoom).Error)

	m := NewUserModel(db, log)

	t.Run("dorm room id only", func(t *testing.T) {
		user := User{
			Type:        UserTypeStudent,
			Name:        "Foo Bar",
			UnixID:      "fbar",
			Visible:     lib.BoolToPtr(true),
			AtWilliams:  lib.BoolToPtr(true),
			DormVisible: lib.BoolToPtr(true),
			HomeVisible: lib.BoolToPtr(false),
			DormRoomID:  &dormRoom.ID,
		}

		fields, err := m.generateSearchFieldsByUser(user)
		testify.NoError(t, err)
		testify.Equal(t, "foo bar#fbar#east 103", fields)
	})

	t.Run("office id only", func(t *testing.T) {
		user := User{
			Type:        UserTypeStudent,
			Name:        "Foo Bar",
			UnixID:      "fbar",
			Visible:     lib.BoolToPtr(true),
			AtWilliams:  lib.BoolToPtr(true),
			DormVisible: lib.BoolToPtr(false),
			HomeVisible: lib.BoolToPtr(false),
			OfficeID:    &office.ID,
		}

		fields, err := m.generateSearchFieldsByUser(user)
		testify.NoError(t, err)
		testify.Equal(t, "foo bar#fbar#schow 12", fields)
	})
}

func TestUserModel_updateUserUnsafe_withExistingDormRoom(t *testing.T) {
	assert := testify.New(t)
	db, log := setupUserSearchFieldsDB(t)

	dbUser := &User{
		Type:        UserTypeStudent,
		Name:        "Old Name",
		UnixID:      "nsf1",
		Visible:     lib.BoolToPtr(true),
		AtWilliams:  lib.BoolToPtr(true),
		DormVisible: lib.BoolToPtr(true),
		HomeVisible: lib.BoolToPtr(false),
		DormRoom: &DormRoom{
			Dorm: &Dorm{
				Neighborhood: &Neighborhood{Name: "Currier"},
				Name:         "East",
			},
			Number: "103",
		},
	}
	assert.NoError(db.Create(dbUser).Error)

	// reload without associations, matching the LDAP sync First() lookup
	var stored User
	assert.NoError(db.First(&stored, dbUser.ID).Error)
	assert.NotNil(stored.DormRoomID)
	assert.Nil(stored.DormRoom)

	toUser := &User{
		Type:          UserTypeStudent,
		Name:          "Nathaniel S. Flores",
		UnixID:        "nsf1",
		WilliamsEmail: "nsf1@williams.edu",
		Visible:       lib.BoolToPtr(true),
		AtWilliams:    lib.BoolToPtr(true),
		WilliamsID:    "3125190",
		ClassYear:     lib.IntToPtr(2026),
		SUBox:         lib.StrToPtr("1234"),
	}

	m := NewUserModel(db, log)
	assert.NoError(m.updateUserUnsafe(&stored, toUser))

	var res User
	assert.NoError(db.First(&res, dbUser.ID).Error)
	assert.Equal("Nathaniel S. Flores", res.Name)
	assert.Contains(res.SearchFields, "east 103")
}
