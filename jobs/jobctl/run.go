package jobctl

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/jobs/registry"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
)

type EnqueueOpts struct {
	JobName           string
	Values            map[string]string
	Trigger           string
	TriggeredByUserID *uint
	RunAt             *time.Time
	InputFileName     *string
	InputSHA256       *string
}

func Enqueue(db *gorm.DB, sd *Systemd, r *registry.Registry, opts EnqueueOpts) (*models.JobRun, error) {
	j, ok := r.Get(opts.JobName)
	if !ok {
		return nil, fmt.Errorf("unknown job %q", opts.JobName)
	}
	args, err := j.BuildArgs(opts.Values)
	if err != nil {
		return nil, err
	}
	argsJSON, err := json.Marshal(opts.Values)
	if err != nil {
		return nil, err
	}
	if opts.Values == nil {
		argsJSON = []byte("{}")
	}

	trigger := opts.Trigger
	if trigger == "" {
		if opts.RunAt != nil {
			trigger = models.JobTriggerScheduled
		} else {
			trigger = models.JobTriggerManual
		}
	}

	run := models.JobRun{
		JobName:           j.Name,
		Trigger:           trigger,
		TriggeredByUserID: opts.TriggeredByUserID,
		Args:              string(argsJSON),
		InputFileName:     opts.InputFileName,
		InputSHA256:       opts.InputSHA256,
		Status:            models.JobStatusPending,
		RunAt:             opts.RunAt,
	}
	if err := db.Create(&run).Error; err != nil {
		return nil, err
	}

	if err := tryStart(db, sd, j, &run, args); err != nil {
		return &run, err
	}
	return &run, nil
}

func DispatchDue(db *gorm.DB, sd *Systemd, r *registry.Registry) (started int, err error) {
	now := time.Now().UTC()
	var pending []models.JobRun
	q := db.Where("status = ?", models.JobStatusPending).
		Where("run_at IS NULL OR run_at <= ?", now).
		Order("id asc")
	if err := q.Find(&pending).Error; err != nil {
		return 0, err
	}

	seenBusy := map[string]bool{}
	for i := range pending {
		run := &pending[i]
		if seenBusy[run.JobName] {
			continue
		}
		j, ok := r.Get(run.JobName)
		if !ok {
			_ = db.Model(run).Updates(map[string]interface{}{
				"status":      models.JobStatusFailed,
				"finished_at": now,
			}).Error
			continue
		}
		active, err := hasActive(db, run.JobName)
		if err != nil {
			return started, err
		}
		if active {
			seenBusy[run.JobName] = true
			continue
		}
		var values map[string]string
		if run.Args != "" {
			_ = json.Unmarshal([]byte(run.Args), &values)
		}
		args, err := j.BuildArgs(values)
		if err != nil {
			_ = db.Model(run).Updates(map[string]interface{}{
				"status":      models.JobStatusFailed,
				"finished_at": now,
			}).Error
			continue
		}
		if err := tryStart(db, sd, j, run, args); err != nil {
			return started, err
		}
		if run.Status == models.JobStatusRunning || run.UnitName != nil {
			started++
			seenBusy[run.JobName] = true
		}
	}
	return started, nil
}

func Cancel(db *gorm.DB, sd *Systemd, runID uint) (*models.JobRun, error) {
	var run models.JobRun
	if err := db.First(&run, runID).Error; err != nil {
		return nil, err
	}
	now := time.Now().UTC()

	switch run.Status {
	case models.JobStatusPending:
		if err := db.Model(&run).Updates(map[string]interface{}{
			"status":      models.JobStatusCanceled,
			"finished_at": now,
		}).Error; err != nil {
			return nil, err
		}
		run.Status = models.JobStatusCanceled
		run.FinishedAt = &now
		return &run, nil

	case models.JobStatusRunning:
		if run.UnitName != nil && *run.UnitName != "" {
			if err := sd.StopUnit(*run.UnitName); err != nil {
				return nil, err
			}
		}
		if err := db.Model(&run).Updates(map[string]interface{}{
			"status":      models.JobStatusCanceled,
			"finished_at": now,
		}).Error; err != nil {
			return nil, err
		}
		run.Status = models.JobStatusCanceled
		run.FinishedAt = &now
		return &run, nil

	default:
		return &run, fmt.Errorf("run %d is %s; cannot cancel", runID, run.Status)
	}
}

func tryStart(db *gorm.DB, sd *Systemd, j registry.Job, run *models.JobRun, args []string) error {
	if run.RunAt != nil && run.RunAt.After(time.Now().UTC()) {
		return nil
	}
	active, err := hasActive(db, j.Name)
	if err != nil {
		return err
	}
	if active {
		return nil
	}
	if sd == nil {
		return fmt.Errorf("systemd client required to start runs")
	}

	unitName, err := sd.StartJobRun(j, run.ID, args)
	if err != nil {
		return err
	}
	if err := db.Model(run).Update("unit_name", unitName).Error; err != nil {
		return err
	}
	run.UnitName = &unitName
	return nil
}

func hasActive(db *gorm.DB, jobName string) (bool, error) {
	var n int
	err := db.Model(&models.JobRun{}).
		Where("job_name = ? AND status = ?", jobName, models.JobStatusRunning).
		Count(&n).Error
	return n > 0, err
}
