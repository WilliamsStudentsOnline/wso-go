package models

import (
	"fmt"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/jinzhu/gorm"
	testify "github.com/stretchr/testify/assert"
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

	userModel := NewUserModel(db)

	students, err := userModel.Students()
	assert.NoError(err)

	assert.Len(students, 2)
	assert.Equal(students[0].Name, "foo")
	assert.Equal(students[1].Name, "baz")
}

func ExampleUserModel_LDAPLookup() {
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

	db := config.LoadDatabase(cfg)
	db.LogMode(true)
	err := db.AutoMigrate(
		User{},
		Department{},
		Neighborhood{},
		Dorm{},
		DormRoom{},
		Office{},
	).Error
	if err != nil {
		panic(err)
	}

	userModel := NewUserModel(db)

	users, err := userModel.LDAPLookup("al15", cfg)
	if err != nil {
		panic(err)
	}

	fmt.Println(users)
}
