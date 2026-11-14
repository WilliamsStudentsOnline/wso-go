package models

import "time"

const (
	JobTriggerTimer     = "timer"
	JobTriggerManual    = "manual"
	JobTriggerScheduled = "scheduled"

	JobStatusPending   = "pending"
	JobStatusRunning   = "running"
	JobStatusSucceeded = "succeeded"
	JobStatusFailed    = "failed"
	JobStatusCanceled  = "canceled"
	JobStatusLost      = "lost"
)

type JobRun struct {
	BaseSchema

	JobName string `gorm:"size:128;not null;index:idx_job_runs_job_name" json:"jobName"`
	Trigger string `gorm:"size:32;not null" json:"trigger"`

	TriggeredByUserID *uint `gorm:"index" json:"triggeredByUserID,omitempty"`
	TriggeredByUser   *User `json:"triggeredByUser,omitempty"`

	Args string `gorm:"type:longtext" json:"args"`

	InputFileName *string `gorm:"size:255" json:"inputFileName,omitempty"`
	InputSHA256   *string `gorm:"size:64" json:"inputSha256,omitempty"`

	Status string `gorm:"size:32;not null;index:idx_job_runs_status" json:"status"`

	RunAt *time.Time `gorm:"index:idx_job_runs_run_at" json:"runAt,omitempty"`

	StartedAt  *time.Time `json:"startedAt,omitempty"`
	FinishedAt *time.Time `json:"finishedAt,omitempty"`

	ExitStatus *int `json:"exitStatus,omitempty"`

	InvocationID *string `gorm:"size:64;index" json:"invocationID,omitempty"`
	UnitName     *string `gorm:"size:128" json:"unitName,omitempty"`
	LogPath      *string `gorm:"size:512" json:"logPath,omitempty"`
}

func (*JobRun) TableName() string {
	return "job_runs"
}
