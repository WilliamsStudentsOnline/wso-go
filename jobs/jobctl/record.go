package jobctl

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
)

func RecordStart(db *gorm.DB) error {
	jobName := os.Getenv("WSO_JOB_NAME")
	if jobName == "" {
		return fmt.Errorf("WSO_JOB_NAME is required")
	}
	invocationID := os.Getenv("INVOCATION_ID")
	unitName := os.Getenv("WSO_JOB_UNIT")
	if unitName == "" {
		unitName = "wso-job-" + jobName + ".service"
	}

	now := time.Now().UTC()

	if idStr := os.Getenv("WSO_JOB_RUN_ID"); idStr != "" {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return fmt.Errorf("WSO_JOB_RUN_ID: %w", err)
		}
		var run models.JobRun
		if err := db.First(&run, id).Error; err != nil {
			return fmt.Errorf("lookup run %d: %w", id, err)
		}
		if run.JobName != jobName {
			return fmt.Errorf("run %d job_name %q != %q", id, run.JobName, jobName)
		}
		updates := map[string]interface{}{
			"status":     models.JobStatusRunning,
			"started_at": now,
		}
		if invocationID != "" {
			updates["invocation_id"] = invocationID
		}
		if unitName != "" {
			updates["unit_name"] = unitName
		}
		return db.Model(&run).Updates(updates).Error
	}

	run := models.JobRun{
		JobName:      jobName,
		Trigger:      models.JobTriggerTimer,
		Status:       models.JobStatusRunning,
		StartedAt:    &now,
		InvocationID: strPtr(invocationID),
		UnitName:     strPtr(unitName),
		Args:         "{}",
	}
	return db.Create(&run).Error
}

func RecordFinish(db *gorm.DB) error {
	jobName := os.Getenv("WSO_JOB_NAME")
	invocationID := os.Getenv("INVOCATION_ID")
	serviceResult := os.Getenv("SERVICE_RESULT")
	exitStatusStr := os.Getenv("EXIT_STATUS")

	var run models.JobRun
	var err error

	if idStr := os.Getenv("WSO_JOB_RUN_ID"); idStr != "" {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return fmt.Errorf("WSO_JOB_RUN_ID: %w", err)
		}
		err = db.First(&run, id).Error
		if err != nil {
			return fmt.Errorf("lookup run %d: %w", id, err)
		}
	} else if invocationID != "" {
		err = db.Where("invocation_id = ?", invocationID).Order("id desc").First(&run).Error
		if gorm.IsRecordNotFoundError(err) && jobName != "" {
			now := time.Now().UTC()
			run = models.JobRun{
				JobName:      jobName,
				Trigger:      models.JobTriggerTimer,
				Status:       models.JobStatusRunning,
				StartedAt:    &now,
				InvocationID: strPtr(invocationID),
				Args:         "{}",
			}
			if err := db.Create(&run).Error; err != nil {
				return err
			}
		} else if err != nil {
			return fmt.Errorf("lookup by invocation: %w", err)
		}
	} else {
		return fmt.Errorf("need WSO_JOB_RUN_ID or INVOCATION_ID")
	}

	now := time.Now().UTC()
	status := models.JobStatusFailed
	if run.Status == models.JobStatusCanceled {
		status = models.JobStatusCanceled
	} else {
		switch serviceResult {
		case "success":
			status = models.JobStatusSucceeded
		case "timeout", "exit-code", "signal", "core-dump", "resources", "failure", "protocol", "exit-signal":
			status = models.JobStatusFailed
		}
		if serviceResult == "success" || exitStatusStr == "0" {
			status = models.JobStatusSucceeded
		}
	}

	var exitStatus *int
	if exitStatusStr != "" && exitStatusStr != "n/a" {
		if n, err := strconv.Atoi(exitStatusStr); err == nil {
			exitStatus = &n
			if n == 0 && status != models.JobStatusCanceled {
				status = models.JobStatusSucceeded
			}
		}
	}

	logPath := filepath.Join(LogDir, fmt.Sprintf("%d.log", run.ID))
	if err := os.MkdirAll(LogDir, 0750); err != nil {
		return fmt.Errorf("mkdir log dir: %w", err)
	}
	if invocationID != "" {
		if copyErr := copyJournalToFile(invocationID, logPath); copyErr != nil {
			logPath = ""
			fmt.Fprintf(os.Stderr, "jobctl record-finish: journal copy: %v\n", copyErr)
		}
	}

	updates := map[string]interface{}{
		"status":      status,
		"finished_at": now,
	}
	if exitStatus != nil {
		updates["exit_status"] = *exitStatus
	}
	if logPath != "" {
		updates["log_path"] = logPath
	}
	if invocationID != "" && (run.InvocationID == nil || *run.InvocationID == "") {
		updates["invocation_id"] = invocationID
	}

	return db.Model(&run).Updates(updates).Error
}

func copyJournalToFile(invocationID, path string) error {
	if strings.ContainsAny(invocationID, " \t\n;'\"`$\\") {
		return fmt.Errorf("invalid invocation id")
	}
	cmd := exec.Command("journalctl", "_SYSTEMD_INVOCATION_ID="+invocationID, "--no-pager", "-o", "short-iso")
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0640)
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
