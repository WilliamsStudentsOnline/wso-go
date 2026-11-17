package models_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	. "github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

type testClock struct {
	now time.Time
}

func (c testClock) Now() time.Time {
	return c.now
}

func TestStudentModel_UpdateFactrakSurveyDeficit(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	log := zaptest.NewLogger(t).Sugar()

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
			Type:   UserTypeStudent,
			Name:   "Student",
			UnixID: "s1",
			BaseSchema: BaseSchema{
				// So we can properly Initialize OnCampusSemesters
				CreatedAt: time.Date(
					m.SeniorYear()-1,
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
			Type:   UserTypeStudent,
			Name:   "Student",
			UnixID: "s1",
			BaseSchema: BaseSchema{
				// So we can properly Initialize OnCampusSemesters
				CreatedAt: time.Date(
					m.SeniorYear()-1,
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
					m.SeniorYear()-2,
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
					m.SeniorYear()-2,
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

func TestStudentModel_UpdateFactrakSurveyDeficit_PerSemester(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	now := time.Date(2025, time.November, 15, 12, 0, 0, 0, time.Local)
	thisSem := time.Date(2025, time.October, 1, 0, 0, 0, 0, time.Local)
	lastSem := thisSem.Add(-time.Second)

	m := NewStudentModel(db, zaptest.NewLogger(t).Sugar())
	m.Clock = testClock{now}

	n := 0
	newStudent := func(onCampusSemesters int, classYear *int) *User {
		n++
		student := &User{
			Type:              UserTypeStudent,
			Name:              "Student",
			UnixID:            fmt.Sprintf("deficit%d", n),
			ClassYear:         classYear,
			OnCampusSemesters: onCampusSemesters,
		}
		assert.NoError(db.Create(student).Error)
		return student
	}
	writeSurveys := func(student *User, createdAt time.Time, season *string, year *int, count int) {
		for i := 0; i < count; i++ {
			survey := FactrakSurvey{UserID: student.ID, SemesterSeason: season, SemesterYear: year}
			survey.CreatedAt = createdAt
			assert.NoError(db.Create(&survey).Error)
		}
	}
	deficit := func(student *User) int {
		assert.NoError(m.UpdateFactrakSurveyDeficit(student))
		var got User
		assert.NoError(db.First(&got, student.ID).Error)
		if !assert.NotNil(got.FactrakSurveyDeficit) {
			return -1
		}
		return *got.FactrakSurveyDeficit
	}
	fall, spring := FactrakSurveySemesterSeasonFall, FactrakSurveySemesterSeasonSpring

	t.Run("pre-frosh", func(t *testing.T) {
		student := newStudent(0, lib.IntToPtr(2029))
		assert.Equal(0, deficit(student))
	})

	t.Run("first semester", func(t *testing.T) {
		student := newStudent(1, lib.IntToPtr(2029))
		assert.Equal(0, deficit(student))
	})

	t.Run("later semester, no reviews written this semester", func(t *testing.T) {
		student := newStudent(5, lib.IntToPtr(2027))
		// Reviews written before the semester began don't count, even about this semester
		writeSurveys(student, lastSem, nil, nil, 10)
		writeSurveys(student, lastSem, &fall, lib.IntToPtr(2025), 2)
		assert.Equal(2, deficit(student))
	})

	t.Run("later semester, one review written this semester", func(t *testing.T) {
		student := newStudent(3, lib.IntToPtr(2028))
		writeSurveys(student, thisSem, &spring, lib.IntToPtr(2025), 1)
		assert.Equal(1, deficit(student))
	})

	t.Run("later semester, two or more reviews written this semester", func(t *testing.T) {
		// The survey's own semester doesn't matter
		two := newStudent(3, lib.IntToPtr(2028))
		writeSurveys(two, thisSem, &fall, lib.IntToPtr(2020), 1)
		writeSurveys(two, now, nil, nil, 1)
		assert.Equal(0, deficit(two))

		three := newStudent(3, lib.IntToPtr(2028))
		writeSurveys(three, now, &spring, lib.IntToPtr(2025), 3)
		assert.Equal(0, deficit(three))
	})

	t.Run("no class year", func(t *testing.T) {
		student := newStudent(4, nil)
		assert.Equal(2, deficit(student))
	})

	t.Run("approved exemption", func(t *testing.T) {
		student := newStudent(4, lib.IntToPtr(2027))
		exemption := FactrakDeficitExemption{
			UserID: student.ID, SemesterSeason: FactrakSurveySemesterSeasonFall, SemesterYear: 2025,
			Reason: "abroad", Status: FactrakDeficitExemptionStatusPending,
		}
		assert.NoError(db.Create(&exemption).Error)
		assert.Equal(2, deficit(student))

		other := FactrakDeficitExemption{
			UserID: student.ID, SemesterSeason: FactrakSurveySemesterSeasonSpring, SemesterYear: 2025,
			Reason: "abroad", Status: FactrakDeficitExemptionStatusApproved,
		}
		assert.NoError(db.Create(&other).Error)
		assert.Equal(2, deficit(student))

		assert.NoError(db.Model(&exemption).Update("status", FactrakDeficitExemptionStatusApproved).Error)
		assert.Equal(0, deficit(student))
	})

	t.Run("recomputing gives the same result", func(t *testing.T) {
		student := newStudent(4, lib.IntToPtr(2027))
		writeSurveys(student, now, nil, nil, 1)
		assert.Equal(1, deficit(student))
		assert.Equal(1, deficit(student))

		// Only the ID set, as the moderator path passes it
		bare := &User{}
		bare.ID = student.ID
		assert.Equal(1, deficit(bare))
	})
}
