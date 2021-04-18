package models

import (
	"testing"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/jinzhu/gorm"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

type testClock struct {
	now time.Time
}

func (c testClock) Now() time.Time {
	return c.now
}

func TestStudentModel_UpdateFactrakSurveyDeficit(t *testing.T) {
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
	testify.NoError(t, err)

	fsM := NewFactrakSurveyModel(db, log)

	t.Run("prefrosh", func(t *testing.T) {
		assert := testify.New(t)
		m := NewStudentModel(db, zaptest.NewLogger(t).Sugar())

		// Set time to be may
		m.Clock = testClock{time.Date(
			time.Now().Year(),
			time.May,
			1,
			1,
			1,
			1,
			1,
			time.Now().Location(),
		)}

		// Test user
		student := User{
			Type:      UserTypeStudent,
			Name:      "Student",
			UnixID:    "s1",
			ClassYear: lib.IntToPtr(4 + m.SeniorYear()),
		}
		// Create, update deficit, and get student
		assert.NoError(db.Create(&student).Error)
		assert.NoError(m.UpdateFactrakSurveyDeficit(&student))
		assert.NoError(db.First(&student).Error)

		// Assert the number of surveys
		assert.Equal(0, *student.FactrakSurveyDeficit)

		// Cleanup
		assert.NoError(db.Unscoped().Delete(&student).Error)
	})

	t.Run("freshman pre-fall", func(t *testing.T) {
		assert := testify.New(t)
		m := NewStudentModel(db, zaptest.NewLogger(t).Sugar())

		// Test user
		student := User{
			Type:      UserTypeStudent,
			Name:      "Student",
			UnixID:    "s1",
			ClassYear: lib.IntToPtr(3 + m.SeniorYear()),
		}

		// Set time to be may
		m.Clock = testClock{time.Date(
			*student.ClassYear-4,
			time.July,
			1,
			1,
			1,
			1,
			1,
			time.Now().Location(),
		)}

		// Create, update deficit, and get student
		assert.NoError(db.Create(&student).Error)
		assert.NoError(m.UpdateFactrakSurveyDeficit(&student))
		assert.NoError(db.First(&student).Error)

		// Assert the number of surveys
		assert.Equal(0, *student.FactrakSurveyDeficit)

		// Cleanup
		assert.NoError(db.Unscoped().Delete(&student).Error)
	})

	t.Run("freshman fall", func(t *testing.T) {
		assert := testify.New(t)
		m := NewStudentModel(db, zaptest.NewLogger(t).Sugar())

		// Test user
		student := User{
			Type:      UserTypeStudent,
			Name:      "Student",
			UnixID:    "s1",
			BaseSchema: BaseSchema{
				// So we can properly Initialize OnCampusSemesters
				CreatedAt: time.Date(
					m.SeniorYear() - 1,
					time.September,
					1,
					1,
					1,
					1,
					1,
					time.Now().Location(),
				),
			},
			ClassYear: lib.IntToPtr(3 + m.SeniorYear()),
		}

		// Set time to be may
		m.Clock = testClock{time.Date(
			*student.ClassYear-4,
			time.November,
			1,
			1,
			1,
			1,
			1,
			time.Now().Location(),
		)}

		// Create, update deficit, and get student
		assert.NoError(db.Create(&student).Error)
		assert.NoError(m.InitializeOnCampusSemesters(&student))
		assert.NoError(m.UpdateFactrakSurveyDeficit(&student))
		assert.NoError(db.First(&student).Error)

		// Assert the number of surveys
		assert.Equal(0, *student.FactrakSurveyDeficit)

		// Cleanup
		assert.NoError(db.Unscoped().Delete(&student).Error)
	})

	t.Run("freshman january", func(t *testing.T) {
		assert := testify.New(t)
		m := &StudentModel{
			UserModel: NewUserModel(db, zaptest.NewLogger(t).Sugar()),
		}

		// Test user
		student := User{
			Type:      UserTypeStudent,
			Name:      "Student",
			UnixID:    "s1",
			BaseSchema: BaseSchema{
				// So we can properly Initialize OnCampusSemesters
				CreatedAt: time.Date(
					m.SeniorYear() - 1,
					time.September,
					1,
					1,
					1,
					1,
					1,
					time.Now().Location(),
				),
			},
			ClassYear: lib.IntToPtr(3 + m.SeniorYear()),
		}

		m.Clock = testClock{time.Date(
			*student.ClassYear-3,
			time.January,
			1,
			1,
			1,
			1,
			1,
			time.Now().Location(),
		)}

		// Create, update deficit, and get student
		assert.NoError(db.Create(&student).Error)
		assert.NoError(m.InitializeOnCampusSemesters(&student))
		assert.NoError(m.UpdateFactrakSurveyDeficit(&student))
		assert.NoError(db.First(&student).Error)

		surveyCount, err := fsM.CountSurveysByUser(student.ID)
		assert.NoError(err)
		assert.Equal(0, surveyCount)

		// Assert the number of surveys
		assert.Equal(2, *student.FactrakSurveyDeficit)

		// Cleanup
		assert.NoError(db.Unscoped().Delete(&student).Error)
	})

	t.Run("when new survey created", func(t *testing.T) {
		assert := testify.New(t)
		m := &StudentModel{
			UserModel: NewUserModel(db, zaptest.NewLogger(t).Sugar()),
		}

		// Test user
		student := User{
			Type:   UserTypeStudent,
			Name:   "Student",
			UnixID: "s1",
			// Sophomore year
			BaseSchema: BaseSchema{
				// So we can properly Initialize OnCampusSemesters
				CreatedAt: time.Date(
					m.SeniorYear() - 2,
					time.September,
					1,
					1,
					1,
					1,
					1,
					time.Now().Location(),
				),
			},
			ClassYear: lib.IntToPtr(2 + m.SeniorYear()),
		}

		m.Clock = testClock{time.Date(
			*student.ClassYear-3,
			time.January,
			1,
			1,
			1,
			1,
			1,
			time.Now().Location(),
		)}

		// Create, update deficit, and get student
		assert.NoError(db.Create(&student).Error)
		assert.NoError(m.InitializeOnCampusSemesters(&student))
		assert.NoError(m.UpdateFactrakSurveyDeficit(&student))
		assert.NoError(db.First(&student).Error)

		surveyCount, err := fsM.CountSurveysByUser(student.ID)
		assert.NoError(err)
		assert.Equal(0, surveyCount)

		// Assert the number of surveys
		assert.Equal(2, *student.FactrakSurveyDeficit)

		// make a factrak survey
		fs1 := FactrakSurvey{
			UserID: student.ID,
		}
		assert.NoError(m.DB.Create(&fs1).Error)

		// recalculate deficit
		assert.NoError(m.UpdateFactrakSurveyDeficit(&student))
		assert.NoError(db.First(&student).Error)

		// assert one less
		assert.Equal(1, *student.FactrakSurveyDeficit)

		// Cleanup
		assert.NoError(db.Unscoped().Delete(&student).Error)
		assert.NoError(db.Unscoped().Delete(&fs1).Error)
	})

	t.Run("when survey is deleted", func(t *testing.T) {
		assert := testify.New(t)
		m := &StudentModel{
			UserModel: NewUserModel(db, zaptest.NewLogger(t).Sugar()),
		}

		// Test user
		student := User{
			Type:   UserTypeStudent,
			Name:   "Student",
			UnixID: "s1",
			// Sophomore year
			BaseSchema: BaseSchema{
				// So we can properly Initialize OnCampusSemesters
				CreatedAt: time.Date(
					m.SeniorYear() - 2,
					time.September,
					1,
					1,
					1,
					1,
					1,
					time.Now().Location(),
				),
			},
			ClassYear: lib.IntToPtr(2 + m.SeniorYear()),
		}

		m.Clock = testClock{time.Date(
			*student.ClassYear-3,
			time.January,
			1,
			1,
			1,
			1,
			1,
			time.Now().Location(),
		)}

		// Create, update deficit, and get student
		assert.NoError(db.Create(&student).Error)

		// make a factrak survey
		fs1 := FactrakSurvey{
			UserID: student.ID,
		}
		assert.NoError(m.DB.Create(&fs1).Error)
		assert.NoError(m.InitializeOnCampusSemesters(&student))
		assert.NoError(m.UpdateFactrakSurveyDeficit(&student))
		assert.NoError(db.First(&student).Error)

		surveyCount, err := fsM.CountSurveysByUser(student.ID)
		assert.NoError(err)
		assert.Equal(1, surveyCount)

		// Assert the number of surveys
		assert.Equal(1, *student.FactrakSurveyDeficit)

		// Delete
		assert.NoError(db.Delete(&fs1).Error)

		// Recalculate deficit
		assert.NoError(m.UpdateFactrakSurveyDeficit(&student))
		assert.NoError(db.First(&student).Error)

		// assert one less
		assert.Equal(2, *student.FactrakSurveyDeficit)

		// Cleanup
		assert.NoError(db.Unscoped().Delete(&student).Error)
	})
}
