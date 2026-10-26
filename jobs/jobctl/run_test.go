package jobctl

import (
	"testing"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/jobs/registry"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnqueueQueuesWhenActive(t *testing.T) {
	db := setupRecordDB(t)
	r := registry.New()
	require.NoError(t, r.Register(registry.Job{
		Name:   "example",
		Binary: registry.InstallDir + "/job-example_linux",
		Params: []registry.Param{
			{Name: "out", Type: registry.ParamString, Default: "/tmp/out"},
		},
	}))

	active := models.JobRun{
		JobName: "example",
		Trigger: models.JobTriggerManual,
		Status:  models.JobStatusRunning,
		Args:    "{}",
	}
	require.NoError(t, db.Create(&active).Error)

	run, err := Enqueue(db, nil, r, EnqueueOpts{JobName: "example", Values: map[string]string{}})
	require.NoError(t, err)
	assert.Equal(t, models.JobStatusPending, run.Status)
	assert.Nil(t, run.UnitName)
}

func TestEnqueueRejectsUnknownJob(t *testing.T) {
	db := setupRecordDB(t)
	r := registry.New()
	_, err := Enqueue(db, nil, r, EnqueueOpts{JobName: "nope"})
	assert.Error(t, err)
}

func TestCancelPending(t *testing.T) {
	db := setupRecordDB(t)
	run := models.JobRun{
		JobName: "example",
		Trigger: models.JobTriggerManual,
		Status:  models.JobStatusPending,
		Args:    "{}",
	}
	require.NoError(t, db.Create(&run).Error)

	got, err := Cancel(db, nil, run.ID)
	require.NoError(t, err)
	assert.Equal(t, models.JobStatusCanceled, got.Status)
	assert.NotNil(t, got.FinishedAt)
}

func TestDispatchDueSkipsFuture(t *testing.T) {
	db := setupRecordDB(t)
	r := registry.New()
	require.NoError(t, r.Register(registry.Job{
		Name:   "example",
		Binary: registry.InstallDir + "/job-example_linux",
	}))

	future := time.Now().UTC().Add(time.Hour)
	run := models.JobRun{
		JobName: "example",
		Trigger: models.JobTriggerScheduled,
		Status:  models.JobStatusPending,
		RunAt:   &future,
		Args:    "{}",
	}
	require.NoError(t, db.Create(&run).Error)

	n, err := DispatchDue(db, nil, r)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
}

func TestValidTransientUnitName(t *testing.T) {
	assert.True(t, ValidTransientUnitName("wso-job-example-r42.service", "example"))
	assert.False(t, ValidTransientUnitName("wso-job-other-r42.service", "example"))
	assert.False(t, ValidTransientUnitName("evil.service", "example"))
}
