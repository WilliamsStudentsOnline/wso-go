package models_test

import (
	"fmt"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	migrate "github.com/WilliamsStudentsOnline/wso-go/db"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	. "github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestUserModel_Students(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	db.Create(&User{
		Type:   UserTypeStudent,
		Name:   "foo",
		UnixID: "u1",
	})

	db.Create(&User{
		Type:   UserTypeAlum,
		Name:   "bar",
		UnixID: "u2",
	})

	db.Create(&User{
		Type:   UserTypeStudent,
		Name:   "baz",
		UnixID: "u3",
	})

	userModel := NewUserModel(db, zaptest.NewLogger(t).Sugar())

	students, err := userModel.Students()
	assert.NoError(err)

	assert.Len(students, 2)
	assert.Equal(students[0].Name, "foo")
	assert.Equal(students[1].Name, "baz")
}

func TestUserModel_PopulateSearchFields(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	u1 := &User{
		Type:        UserTypeStudent,
		Name:        "Foo",
		UnixID:      "user1",
		Title:       lib.StrToPtr("Research Assistant"),
		ClassYear:   lib.IntToPtr(2022),
		Major:       lib.StrToPtr("Computer Science"),
		SUBox:       lib.StrToPtr("2885"),
		Entry:       lib.StrToPtr("AP3"),
		DormVisible: lib.BoolToPtr(true),
		DormRoom: &DormRoom{
			Dorm: &Dorm{
				Neighborhood: &Neighborhood{
					Name: "Currier",
				},
				Name: "East",
			},
			Number: "103",
		},
		HomeVisible: lib.BoolToPtr(true),
		HomeTown:    lib.StrToPtr("Palo Alto"),
		HomeState:   lib.StrToPtr("California"),
		HomeCountry: lib.StrToPtr("United States"),
	}
	assert.NoError(db.Create(&u1).Error)

	m := NewUserModel(db, zaptest.NewLogger(t).Sugar())
	assert.NoError(m.PopulateSearchFields(u1.ID))

	var res User
	assert.NoError(db.First(&res, u1.ID).Error)

	assert.Equal("foo#user1#research assistant#2022#computer science#2885#ap3#east 103#palo alto, california", res.SearchFields)
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

	db := config.LoadDatabase(cfg, zap.S())
	db.LogMode(true)
	err := db.AutoMigrate(
		User{},
		Department{},
		Neighborhood{},
		Dorm{},
		DormRoom{},
		Office{},
		Bulletin{},
	).Error
	if err != nil {
		panic(err)
	}

	userModel := NewUserModel(db, zap.S())

	users, err := userModel.LDAPLookup("al15", cfg)
	if err != nil {
		panic(err)
	}

	fmt.Println(users)
}

func TestUserModel_DoesUserExist(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	db.Create(&User{
		Type:   UserTypeStudent,
		Name:   "foo",
		UnixID: "u1",
	})

	db.Create(&User{
		Type:   UserTypeAlum,
		Name:   "bar",
		UnixID: "u2",
	})

	db.Create(&User{
		Type:   UserTypeStudent,
		Name:   "baz",
		UnixID: "u3",
	})

	m := NewUserModel(db, zaptest.NewLogger(t).Sugar())

	t.Run("does exist", func(t *testing.T) {
		exists, err := m.DoesUserExist(2)
		testify.NoError(t, err)
		testify.True(t, exists)
	})

	t.Run("does not exist", func(t *testing.T) {
		exists, err := m.DoesUserExist(5)
		testify.NoError(t, err)
		testify.False(t, exists)
	})
}

func BenchmarkUserModel_DoesUserExist(b *testing.B) {
	cfg := utils.SetupConfig()

	log := zaptest.NewLogger(b).Sugar()

	db := config.LoadDatabase(cfg, log)
	_ = migrate.MigrateDB(db)

	db.Create(&User{
		Type:   UserTypeStudent,
		Name:   "foo",
		UnixID: "u1",
	})

	db.Create(&User{
		Type:   UserTypeAlum,
		Name:   "bar",
		UnixID: "u2",
	})

	db.Create(&User{
		Type:   UserTypeStudent,
		Name:   "baz",
		UnixID: "u3",
	})

	m := NewUserModel(db, log)

	b.Run("does exist", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			_, _ = m.DoesUserExist(2)
		}
	})

	b.Run("does not exist", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			_, _ = m.DoesUserExist(5)
		}
	})
}
