package models_test

import (
	"testing"
	"time"

	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	. "github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestModeratorTaskModel_ResolveOpenTasksByTaskReference(t *testing.T) {
	assert := assert.New(t)
	db := utils.SetupServiceTest(assert)
	m := NewModeratorTaskModel(db, zap.NewNop().Sugar())

	uid1, uid2 := uint(1), uint(2)
	t1, err := NewFactrakReviewTask(&uid1, 10)
	require.NoError(t, err)
	t2, err := NewFactrakReviewTask(&uid2, 10)
	require.NoError(t, err)
	other, err := NewFactrakReviewTask(&uid1, 11)
	require.NoError(t, err)
	require.NoError(t, m.CreateTask(t1))
	require.NoError(t, m.CreateTask(t2))
	require.NoError(t, m.CreateTask(other))

	n, err := m.ResolveOpenTasksByTaskReference(
		TaskReference(ModeratorTaskTypeFactrakReview, 10),
		ModeratorTaskResolutionApproved,
		99,
		time.Now().UTC(),
	)
	require.NoError(t, err)
	assert.Equal(int64(2), n)

	var open []*ModeratorTask
	require.NoError(t, m.GetAllTasks(&open, nil))
	require.Len(t, open, 1)
	assert.Equal(other.ID, open[0].ID)

	var resolved []*ModeratorTask
	anyTime := time.Now().UTC()
	key := TaskReference(ModeratorTaskTypeFactrakReview, 10)
	require.NoError(t, m.GetAllTasks(&resolved, &GetAllModeratorTasksOptions{
		TaskReference: &key,
		ResolvedAt:    &anyTime,
	}))
	require.Len(t, resolved, 2)
	for _, task := range resolved {
		require.NotNil(t, task.ResolvedAt)
		assert.Equal(ModeratorTaskResolutionApproved, *task.Resolution)
		assert.Equal(uint(99), *task.ResolvedByUserID)
	}
}

func TestModeratorTaskModel_SoftDeleteOpenTasksByTaskReference(t *testing.T) {
	assert := assert.New(t)
	db := utils.SetupServiceTest(assert)
	m := NewModeratorTaskModel(db, zap.NewNop().Sugar())

	uid := uint(5)
	task, err := NewFactrakReviewTask(&uid, 20)
	require.NoError(t, err)
	require.NoError(t, m.CreateTask(task))

	keep, err := NewFactrakReviewTask(&uid, 21)
	require.NoError(t, err)
	require.NoError(t, m.CreateTask(keep))

	n, err := m.SoftDeleteOpenTasksByTaskReference(TaskReference(ModeratorTaskTypeFactrakReview, 20))
	require.NoError(t, err)
	assert.Equal(int64(1), n)

	var open []*ModeratorTask
	require.NoError(t, m.GetAllTasks(&open, nil))
	require.Len(t, open, 1)
	assert.Equal(keep.ID, open[0].ID)

	var count int
	assert.NoError(db.Unscoped().Model(&ModeratorTask{}).Where("id = ?", task.ID).Count(&count).Error)
	assert.Equal(1, count)
}

func TestModeratorTaskModel_GetAllTasks_DefaultsToOpen(t *testing.T) {
	assert := assert.New(t)
	db := utils.SetupServiceTest(assert)
	m := NewModeratorTaskModel(db, zap.NewNop().Sugar())

	uid := uint(1)
	openTask, err := NewFactrakReviewTask(&uid, 1)
	require.NoError(t, err)
	require.NoError(t, m.CreateTask(openTask))

	closedTask, err := NewFactrakReviewTask(&uid, 2)
	require.NoError(t, err)
	require.NoError(t, m.CreateTask(closedTask))
	_, err = m.ResolveOpenTasksByTaskReference(
		closedTask.TaskReference,
		ModeratorTaskResolutionRejected,
		7,
		time.Now().UTC(),
	)
	require.NoError(t, err)

	var open []*ModeratorTask
	require.NoError(t, m.GetAllTasks(&open, &GetAllModeratorTasksOptions{}))
	require.Len(t, open, 1)
	assert.Equal(openTask.ID, open[0].ID)

	count, err := m.CountAllTasks(&GetAllModeratorTasksOptions{})
	require.NoError(t, err)
	assert.Equal(1, count)
}

func TestModeratorTaskModel_GetOpenTasksByTaskReference(t *testing.T) {
	assert := assert.New(t)
	db := utils.SetupServiceTest(assert)
	m := NewModeratorTaskModel(db, zap.NewNop().Sugar())

	uid := uint(1)
	a, err := NewFactrakReviewTask(&uid, 3)
	require.NoError(t, err)
	b, err := NewFactrakReviewTask(&uid, 3)
	require.NoError(t, err)
	require.NoError(t, m.CreateTask(a))
	require.NoError(t, m.CreateTask(b))

	var tasks []*ModeratorTask
	require.NoError(t, m.GetOpenTasksByTaskReference(TaskReference(ModeratorTaskTypeFactrakReview, 3), &tasks))
	require.Len(t, tasks, 2)
	assert.True(tasks[0].CreatedAt.Before(tasks[1].CreatedAt) || tasks[0].CreatedAt.Equal(tasks[1].CreatedAt))
}
