package models

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUserModel_LDAPLookup(t *testing.T) {
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
	err := db.AutoMigrate(
		User{},
		Department{},
		Neighborhood{},
		Dorm{},
		DormRoom{},
		Office{},
	).Error
	assert.NoError(t, err)

	userModel := &UserModel{
		BaseModel{
			DB: db,
		},
	}

	users, err := userModel.LDAPLookup("10rem", cfg)
	assert.NoError(t, err)

	_ = users
}