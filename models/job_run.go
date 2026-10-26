package models

import (
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type JobRunModel struct {
	*BaseModel
}

func NewJobRunModel(db *gorm.DB, log *zap.SugaredLogger) *JobRunModel {
	return &JobRunModel{BaseModel: NewBaseModel(db, log)}
}

func (m *JobRunModel) FindByID(id uint) (*JobRun, error) {
	var run JobRun
	err := m.DB.First(&run, id).Error
	if err != nil {
		return nil, err
	}
	return &run, nil
}

func (m *JobRunModel) FindByInvocationID(invocationID string) (*JobRun, error) {
	var run JobRun
	err := m.DB.Where("invocation_id = ?", invocationID).First(&run).Error
	if err != nil {
		return nil, err
	}
	return &run, nil
}

// HasActiveRun reports whether jobName has a running row (concurrency gate).
func (m *JobRunModel) HasActiveRun(jobName string) (bool, error) {
	var n int
	err := m.DB.Model(&JobRun{}).
		Where("job_name = ? AND status = ?", jobName, JobStatusRunning).
		Count(&n).Error
	return n > 0, err
}
