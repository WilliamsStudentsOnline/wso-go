package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubjectKeyForFactrakSurvey(t *testing.T) {
	assert.Equal(t, "factrak_survey:42", SubjectKeyForFactrakSurvey(42))
}

func TestSubjectKeyForTaskType(t *testing.T) {
	key, err := SubjectKeyForTaskType(ModeratorTaskTypeFactrakReview, `{"factrakSurveyID":7}`)
	require.NoError(t, err)
	assert.Equal(t, "factrak_survey:7", key)

	_, err = SubjectKeyForTaskType(ModeratorTaskTypeFactrakReview, `{}`)
	assert.Error(t, err)

	_, err = SubjectKeyForTaskType(ModeratorTaskTypeUndefined, `{}`)
	assert.Error(t, err)
}

func TestNewFactrakReviewTask(t *testing.T) {
	uid := uint(9)
	task, err := NewFactrakReviewTask(&uid, 3)
	require.NoError(t, err)
	assert.Equal(t, ModeratorTaskTypeFactrakReview, task.TaskType)
	assert.Equal(t, "factrak_survey:3", task.SubjectKey)
	assert.Equal(t, &uid, task.RequestingUserID)

	fields, err := DecodeFactrakReviewFields(task.TaskFields)
	require.NoError(t, err)
	assert.Equal(t, uint(3), fields.FactrakSurveyID)
}
