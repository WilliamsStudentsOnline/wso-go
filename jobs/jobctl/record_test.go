package jobctl

import (
	"os"
	"strconv"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupRecordDB(t *testing.T) *gorm.DB {
	t.Helper()
	cfg := &config.Config{
		Env:          "test",
		DatabaseType: "sqlite3",
		DatabaseArgs: ":memory:",
		Secrets:      &config.Secrets{},
	}
	db := config.LoadDatabase(cfg, zap.NewNop().Sugar())
	require.NoError(t, db.AutoMigrate(&models.JobRun{}).Error)
	t.Cleanup(func() { config.CloseDatabase(db, zap.NewNop().Sugar()) })
	return db
}

func TestRecordStartTimerCreatesRow(t *testing.T) {
	db := setupRecordDB(t)

	t.Setenv("WSO_JOB_NAME", "library-hours-update")
	t.Setenv("INVOCATION_ID", "inv-timer-1")
	t.Setenv("WSO_JOB_UNIT", "wso-job-library-hours-update.service")
	os.Unsetenv("WSO_JOB_RUN_ID")

	require.NoError(t, RecordStart(db))

	var run models.JobRun
	require.NoError(t, db.First(&run).Error)
	assert.Equal(t, "library-hours-update", run.JobName)
	assert.Equal(t, models.JobTriggerTimer, run.Trigger)
	assert.Equal(t, models.JobStatusRunning, run.Status)
	require.NotNil(t, run.InvocationID)
	assert.Equal(t, "inv-timer-1", *run.InvocationID)
}

func TestRecordStartUpdatesPending(t *testing.T) {
	db := setupRecordDB(t)

	pending := models.JobRun{
		JobName: "library-hours-update",
		Trigger: models.JobTriggerManual,
		Status:  models.JobStatusPending,
		Args:    `{"out":"/tmp/x"}`,
	}
	require.NoError(t, db.Create(&pending).Error)

	t.Setenv("WSO_JOB_NAME", "library-hours-update")
	t.Setenv("WSO_JOB_RUN_ID", strconv.FormatUint(uint64(pending.ID), 10))
	t.Setenv("INVOCATION_ID", "inv-manual-1")

	require.NoError(t, RecordStart(db))

	var run models.JobRun
	require.NoError(t, db.First(&run, pending.ID).Error)
	assert.Equal(t, models.JobStatusRunning, run.Status)
	require.NotNil(t, run.InvocationID)
	assert.Equal(t, "inv-manual-1", *run.InvocationID)
	assert.NotNil(t, run.StartedAt)
}

func TestRecordFinishSuccess(t *testing.T) {
	db := setupRecordDB(t)
	LogDir = t.TempDir()

	inv := "inv-finish-1"
	run := models.JobRun{
		JobName:      "library-hours-update",
		Trigger:      models.JobTriggerTimer,
		Status:       models.JobStatusRunning,
		InvocationID: &inv,
		Args:         "{}",
	}
	require.NoError(t, db.Create(&run).Error)

	t.Setenv("WSO_JOB_NAME", "library-hours-update")
	t.Setenv("INVOCATION_ID", inv)
	t.Setenv("SERVICE_RESULT", "success")
	t.Setenv("EXIT_STATUS", "0")
	os.Unsetenv("WSO_JOB_RUN_ID")

	require.NoError(t, RecordFinish(db))

	var got models.JobRun
	require.NoError(t, db.First(&got, run.ID).Error)
	assert.Equal(t, models.JobStatusSucceeded, got.Status)
	assert.NotNil(t, got.FinishedAt)
}
