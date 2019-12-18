package models_test

import (
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	. "github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestProfessorModel_GetProfessorsByAreaOfStudy(t *testing.T) {
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

	log := zaptest.NewLogger(t).Sugar()

	db := config.LoadDatabase(cfg, log)
	db.SetLogger(gorm.Logger{LogWriter: zap.NewStdLog(log.Desugar())})
	db.LogMode(true)
	err := db.AutoMigrate(
		User{},
		Department{},
		Course{},
		AreaOfStudy{},
		FactrakSurvey{},
		FactrakAgreement{},
	).Error
	assert.NoError(err)

	// Insert test user into db
	p1 := User{
		Type:   UserTypeProfessor,
		Name:   "Professor 1",
		UnixID: "p1",
	}
	p2 := User{
		Type:   UserTypeProfessor,
		Name:   "Professor 2",
		UnixID: "p2",
	}
	// Not at williams
	p3 := User{
		Type:       UserTypeProfessor,
		Name:       "Professor 3",
		UnixID:     "p3",
		AtWilliams: lib.BoolToPtr(false),
	}
	// Other area/dept
	p4 := User{
		Type:   UserTypeProfessor,
		Name:   "Professor 4",
		UnixID: "p4",
	}
	// Staff
	s1 := User{
		Type:   UserTypeStaff,
		Name:   "Staff 1",
		UnixID: "s1",
	}
	assert.NoError(db.Create(&p1).Create(&p2).Create(&p3).Create(&p4).Create(&s1).Error)

	d1 := Department{
		Name: "Computer Science",
		AreasOfStudy: []*AreaOfStudy{
			{
				Name:         "Computer Science",
				Abbreviation: "CSCI",
			},
		},
		Users: []*User{
			&p1,
			&p2,
			&p3,
			&s1,
		},
	}
	d2 := Department{
		Name: "Economics",
		AreasOfStudy: []*AreaOfStudy{
			{
				Name:         "Economics",
				Abbreviation: "ECON",
			},
		},
		Users: []*User{
			&p4,
		},
	}
	assert.NoError(db.Create(&d1).Create(&d2).Error)

	m := NewProfessorModel(db, log)
	var resProfs []*User
	areaID := d1.AreasOfStudy[0].ID
	err = m.GetProfessorsByAreaOfStudy(areaID, &resProfs)
	assert.NoError(err)

	assert.Len(resProfs, 2)

}
