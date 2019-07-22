package models

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/jinzhu/gorm"
	testify "github.com/stretchr/testify/assert"
	"testing"
)

func testSetup(assert *testify.Assertions) (cfg *config.Config, db *gorm.DB) {
	cfg = &config.Config{
		Env:          "test",
		GinMode:      "test",
		JWTRealm:     "wso-go-test",
		DatabaseType: "sqlite3",
		DatabaseArgs: ":memory:",
		Secrets: &config.Secrets{
			JWTSecretKey: "wso-jwt-test-secret",
		},
	}

	db = config.LoadDatabase(cfg)
	db.LogMode(true)
	err := db.AutoMigrate(
		User{},
		Department{},
		Neighborhood{},
		Dorm{},
		DormRoom{},
		Office{},
	).Error
	assert.NoError(err)
	return
}

func TestUserModel_Students(t *testing.T) {
	assert := testify.New(t)
	_, db := testSetup(assert)

	db.Create(&User{
		Type: "student",
		Name: "foo",
	})

	db.Create(&User{
		Type: "alum",
		Name: "bar",
	})

	db.Create(&User{
		Type: "student",
		Name: "baz",
	})

	userModel := &UserModel{
		BaseModel{
			DB: db,
		},
	}

	students, err := userModel.Students()
	assert.NoError(err)

	assert.Len(students, 2)
	assert.Equal(students[0].Name, "foo")
	assert.Equal(students[1].Name,"baz")
}

func TestUserModel_LDAPLookup(t *testing.T) {
	assert := testify.New(t)
	cfg, db := testSetup(assert)

	userModel := &UserModel{
		BaseModel{
			DB: db,
		},
	}

	users, err := userModel.LDAPLookup("10rem", cfg)
	assert.NoError(err)

	_ = users
}