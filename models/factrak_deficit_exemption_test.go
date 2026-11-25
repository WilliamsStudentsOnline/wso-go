package models_test

import (
	"testing"
	"time"

	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	. "github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestFactrakDeficitSemester(t *testing.T) {
	assert := testify.New(t)
	cases := []struct {
		month  time.Month
		year   int
		season string
		semYr  int
	}{
		{time.January, 2026, FactrakSurveySemesterSeasonFall, 2025},
		{time.February, 2026, FactrakSurveySemesterSeasonFall, 2025},
		{time.March, 2026, FactrakSurveySemesterSeasonSpring, 2026},
		{time.September, 2026, FactrakSurveySemesterSeasonSpring, 2026},
		{time.October, 2026, FactrakSurveySemesterSeasonFall, 2026},
		{time.December, 2026, FactrakSurveySemesterSeasonFall, 2026},
	}
	for _, c := range cases {
		season, year := FactrakDeficitSemester(time.Date(c.year, c.month, 15, 12, 0, 0, 0, time.Local))
		assert.Equal(c.season, season, "%s %d", c.month, c.year)
		assert.Equal(c.semYr, year, "%s %d", c.month, c.year)
	}

	// Switches at the start of March 1 and October 1, when the recompute job runs
	season, year := FactrakDeficitSemester(time.Date(2026, time.February, 28, 23, 59, 0, 0, time.Local))
	assert.Equal(FactrakSurveySemesterSeasonFall, season)
	assert.Equal(2025, year)
	season, year = FactrakDeficitSemester(time.Date(2026, time.March, 1, 0, 0, 0, 0, time.Local))
	assert.Equal(FactrakSurveySemesterSeasonSpring, season)
	assert.Equal(2026, year)
	season, year = FactrakDeficitSemester(time.Date(2026, time.September, 30, 23, 59, 0, 0, time.Local))
	assert.Equal(FactrakSurveySemesterSeasonSpring, season)
	assert.Equal(2026, year)
	season, year = FactrakDeficitSemester(time.Date(2026, time.October, 1, 0, 0, 0, 0, time.Local))
	assert.Equal(FactrakSurveySemesterSeasonFall, season)
	assert.Equal(2026, year)
}

func TestFactrakDeficitSemesterStart(t *testing.T) {
	assert := testify.New(t)

	cases := []struct {
		now, start time.Time
	}{
		{time.Date(2026, time.October, 1, 0, 0, 0, 0, time.Local), time.Date(2026, time.October, 1, 0, 0, 0, 0, time.Local)},
		{time.Date(2026, time.December, 15, 12, 0, 0, 0, time.Local), time.Date(2026, time.October, 1, 0, 0, 0, 0, time.Local)},
		{time.Date(2027, time.February, 28, 23, 59, 0, 0, time.Local), time.Date(2026, time.October, 1, 0, 0, 0, 0, time.Local)},
		{time.Date(2027, time.March, 1, 0, 0, 0, 0, time.Local), time.Date(2027, time.March, 1, 0, 0, 0, 0, time.Local)},
		{time.Date(2027, time.September, 30, 23, 59, 0, 0, time.Local), time.Date(2027, time.March, 1, 0, 0, 0, 0, time.Local)},
	}
	for _, c := range cases {
		assert.Equal(c.start, FactrakDeficitSemesterStart(c.now), c.now.String())
	}
}

func TestFactrakDeficitExemptionModel_CreateExemptionRequest(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	m := NewFactrakDeficitExemptionModel(db, zaptest.NewLogger(t).Sugar())
	now := time.Date(2025, time.November, 15, 12, 0, 0, 0, time.Local)

	student := User{Type: UserTypeStudent, Name: "Student", UnixID: "exempt1"}
	require.NoError(t, db.Create(&student).Error)

	e, err := m.CreateExemptionRequest(student.ID, "studying abroad", now)
	require.NoError(t, err)
	assert.Equal(FactrakDeficitExemptionStatusPending, e.Status)
	assert.Equal(FactrakSurveySemesterSeasonFall, e.SemesterSeason)
	assert.Equal(2025, e.SemesterYear)

	var task ModeratorTask
	require.NoError(t, db.Where("task_reference = ?",
		TaskReference(ModeratorTaskTypeFactrakDeficitExemption, e.ID)).First(&task).Error)
	assert.Equal(ModeratorTaskTypeFactrakDeficitExemption, task.TaskType)
	require.NotNil(t, task.RequestingUserID)
	assert.Equal(student.ID, *task.RequestingUserID)
	fields, err := DecodeFactrakDeficitExemptionFields(task.TaskFields)
	require.NoError(t, err)
	assert.Equal(e.ID, fields.FactrakDeficitExemptionID)

	_, err = m.CreateExemptionRequest(student.ID, "again", now)
	assert.Equal(ErrFactrakDeficitExemptionExists, err)

	require.NoError(t, m.SetExemptionStatus(e.ID, FactrakDeficitExemptionStatusApproved))
	_, err = m.CreateExemptionRequest(student.ID, "again", now)
	assert.Equal(ErrFactrakDeficitExemptionExists, err)
	exempt, err := m.HasApprovedExemption(student.ID, now)
	assert.NoError(err)
	assert.True(exempt)

	require.NoError(t, m.SetExemptionStatus(e.ID, FactrakDeficitExemptionStatusRejected))
	_, err = m.CreateExemptionRequest(student.ID, "again", now)
	assert.NoError(err)

	_, err = m.CreateExemptionRequest(student.ID, "next semester", now.AddDate(0, 5, 0))
	assert.NoError(err)
}
