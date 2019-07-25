package models

import (
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	testify "github.com/stretchr/testify/assert"
)

func TestCourseModel_FindByAbbrevAndNumber(t *testing.T) {
	assert := testify.New(t)

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
		Course{},
		AreaOfStudy{},
		FactrakSurvey{},
		Neighborhood{},
		FactrakAgreement{},
	).Error
	assert.NoError(err)

	a1 := AreaOfStudy{
		Name:         "Computer Science",
		Abbreviation: "CSCI",
		Department: &Department{
			Name: "Computer Science",
		},
	}

	// Populate
	c1 := Course{
		Number:      "136",
		AreaOfStudy: &a1,
	}
	c2 := Course{
		Number: "136", // Same number
		AreaOfStudy: &AreaOfStudy{
			Name:         "Economics",
			Abbreviation: "ECON",
			Department: &Department{
				Name: "Economics",
			},
		},
	}
	c3 := Course{
		Number:      "256",
		AreaOfStudy: &a1, // Same dept
	}

	err = db.Create(&c1).Create(&c2).Create(&c3).Error
	assert.NoError(err)

	m := NewCourseModel(db)
	var resC Course
	err = m.FindByAbbrevAndNumber("CSCI", "136", &resC)
	assert.NoError(err)

	assert.Equal(c1.ID, resC.ID)
	assert.Equal("136", resC.Number)
	assert.Equal("CSCI", resC.AreaOfStudy.Abbreviation)
	assert.Equal("Computer Science", resC.AreaOfStudy.Department.Name)
}
