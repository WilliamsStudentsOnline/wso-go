package admin

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/jobs/jobctl"
	"github.com/WilliamsStudentsOnline/wso-go/jobs/registry"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

const maxJobUploadBytes = 32 << 20 // 32 MiB

type jobInfo struct {
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	Params       []registry.Param   `json:"params"`
	Schedule     *registry.Schedule `json:"schedule,omitempty"`
	NextRunAt    *time.Time         `json:"nextRunAt,omitempty"`
	HasActiveRun bool               `json:"hasActiveRun"`
}

type createJobRunRequest struct {
	Params map[string]string `json:"params"`
	RunAt  *time.Time        `json:"runAt"`
}

// ListJobs godoc
// @Summary List batch jobs
// @Description list registered batch jobs and next timer elapse when available
// @ID listAdminJobs
// @Tags admin-jobs
// @Produce json
// @Success 200 {array} admin.jobInfo
// @Failure 403 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /admin/jobs [get]
func (t *Controller) ListJobs(c *gin.Context) {
	r := registry.Default()
	var sd *jobctl.Systemd
	if s, err := jobctl.NewSystemd(); err == nil {
		sd = s
		defer sd.Close()
	}

	out := make([]jobInfo, 0, len(r.All()))
	for _, j := range r.All() {
		info := jobInfo{
			Name:        j.Name,
			Description: j.Description,
			Params:      j.Params,
			Schedule:    j.Schedule,
		}
		active, err := t.jobRunModel.HasActiveRun(j.Name)
		if err != nil {
			t.RespondError(c, err)
			return
		}
		info.HasActiveRun = active
		if sd != nil && j.Schedule != nil {
			if next, err := sd.TimerNextElapse(j.Name); err == nil {
				info.NextRunAt = next
			}
		}
		out = append(out, info)
	}
	t.RespondOK(c, out)
}

// ListJobRuns godoc
// @Summary List job runs
// @Description list job run history with optional filters
// @ID listAdminJobRuns
// @Tags admin-jobs
// @Produce json
// @Param job query string false "job name"
// @Param status query string false "run status"
// @Param limit query int false "max rows" default(50)
// @Success 200 {array} models.JobRun
// @Failure 403 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /admin/jobs/runs [get]
func (t *Controller) ListJobRuns(c *gin.Context) {
	q := t.db.Model(&models.JobRun{}).Order("id desc")
	if job := c.Query("job"); job != "" {
		q = q.Where("job_name = ?", job)
	}
	if status := c.Query("status"); status != "" {
		q = q.Where("status = ?", status)
	}
	limit := 50
	if s := c.Query("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	var runs []models.JobRun
	if err := q.Limit(limit).Find(&runs).Error; err != nil {
		t.RespondError(c, err)
		return
	}
	t.RespondOK(c, runs)
}

// GetJobRun godoc
// @Summary Get a job run
// @ID getAdminJobRun
// @Tags admin-jobs
// @Produce json
// @Param id path int true "run id"
// @Success 200 {object} models.JobRun
// @Failure 404 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /admin/jobs/runs/{id} [get]
func (t *Controller) GetJobRun(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		t.RespondAPIError(c, lib.ErrorMalformedRequestData)
		return
	}
	run, err := t.jobRunModel.FindByID(uint(id))
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			t.RespondAPIError(c, lib.ErrorRecordNotFound)
			return
		}
		t.RespondError(c, err)
		return
	}
	t.RespondOK(c, run)
}

// GetJobRunLogs godoc
// @Summary Get job run logs
// @ID getAdminJobRunLogs
// @Tags admin-jobs
// @Produce text/plain
// @Param id path int true "run id"
// @Success 200 {string} string
// @Failure 404 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /admin/jobs/runs/{id}/logs [get]
func (t *Controller) GetJobRunLogs(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		t.RespondAPIError(c, lib.ErrorMalformedRequestData)
		return
	}
	run, err := t.jobRunModel.FindByID(uint(id))
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			t.RespondAPIError(c, lib.ErrorRecordNotFound)
			return
		}
		t.RespondError(c, err)
		return
	}

	if run.LogPath != nil && *run.LogPath != "" {
		if !safeLogPath(*run.LogPath) {
			t.RespondAPIError(c, lib.ErrorNoScopeAuthorization)
			return
		}
		c.File(*run.LogPath)
		return
	}

	if run.InvocationID != nil && *run.InvocationID != "" {
		out, err := journalForInvocation(*run.InvocationID)
		if err != nil {
			t.RespondError(c, err)
			return
		}
		c.Data(200, "text/plain; charset=utf-8", out)
		return
	}

	t.RespondAPIError(c, lib.ErrorRecordNotFound)
}

// CreateJobRun godoc
// @Summary Enqueue or schedule a job run
// @Description JSON body or multipart (params JSON + optional file). Queues if the job is already running.
// @ID createAdminJobRun
// @Tags admin-jobs
// @Accept json
// @Produce json
// @Param name path string true "job name"
// @Param body body admin.createJobRunRequest false "JSON params"
// @Success 200 {object} models.JobRun
// @Failure 400 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /admin/jobs/{name}/runs [post]
func (t *Controller) CreateJobRun(c *gin.Context) {
	name := c.Param("name")
	reg := registry.Default()
	j, ok := reg.Get(name)
	if !ok {
		t.RespondAPIError(c, lib.ErrorRecordNotFound)
		return
	}

	values := map[string]string{}
	var runAt *time.Time
	var inputName *string
	var inputHash *string

	ct := c.ContentType()
	if strings.HasPrefix(ct, "multipart/") {
		if params := c.PostForm("params"); params != "" {
			if err := json.Unmarshal([]byte(params), &values); err != nil {
				t.RespondBadBind(c, err)
				return
			}
		}
		if s := c.PostForm("runAt"); s != "" {
			tm, err := time.Parse(time.RFC3339, s)
			if err != nil {
				t.RespondBadBind(c, err)
				return
			}
			tm = tm.UTC()
			runAt = &tm
		}
		if fh, err := c.FormFile("file"); err == nil {
			path, orig, hash, err := saveJobUpload(fh)
			if err != nil {
				t.RespondError(c, err)
				return
			}
			inputName = &orig
			inputHash = &hash
			for _, p := range j.Params {
				if p.Type == registry.ParamFile {
					values[p.Name] = path
				}
			}
		}
	} else {
		var body createJobRunRequest
		if err := c.ShouldBindJSON(&body); err != nil {
			t.RespondBadBind(c, err)
			return
		}
		if body.Params != nil {
			values = body.Params
		}
		runAt = body.RunAt
	}

	userID := services.GetUserID(c)
	sd, err := jobctl.NewSystemd()
	if err != nil {
		t.RespondError(c, err)
		return
	}
	defer sd.Close()

	run, err := jobctl.Enqueue(t.db, sd, reg, jobctl.EnqueueOpts{
		JobName:           name,
		Values:            values,
		TriggeredByUserID: &userID,
		RunAt:             runAt,
		InputFileName:     inputName,
		InputSHA256:       inputHash,
	})
	if err != nil {
		t.Log.Warnf("admin create job run user=%d job=%s: %v", userID, name, err)
		t.RespondError(c, err)
		return
	}
	t.Log.Infof("admin create job run user=%d run=%d job=%s status=%s", userID, run.ID, name, run.Status)
	t.RespondOK(c, run)
}

// CancelJobRun godoc
// @Summary Cancel a job run
// @Description cancel a pending run or stop a running one
// @ID cancelAdminJobRun
// @Tags admin-jobs
// @Produce json
// @Param id path int true "run id"
// @Success 200 {object} models.JobRun
// @Failure 400 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /admin/jobs/runs/{id}/cancel [post]
func (t *Controller) CancelJobRun(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		t.RespondAPIError(c, lib.ErrorMalformedRequestData)
		return
	}
	sd, err := jobctl.NewSystemd()
	if err != nil {
		t.RespondError(c, err)
		return
	}
	defer sd.Close()

	run, err := jobctl.Cancel(t.db, sd, uint(id))
	if err != nil {
		t.RespondError(c, err)
		return
	}
	t.Log.Infof("admin cancel job run user=%d run=%d status=%s", services.GetUserID(c), run.ID, run.Status)
	t.RespondOK(c, run)
}

func safeLogPath(p string) bool {
	clean := filepath.Clean(p)
	return strings.HasPrefix(clean, filepath.Clean(jobctl.LogDir)+string(os.PathSeparator))
}

func journalForInvocation(invocationID string) ([]byte, error) {
	if strings.ContainsAny(invocationID, " \t\n;'\"`$\\") {
		return nil, fmt.Errorf("invalid invocation id")
	}
	return exec.Command("journalctl", "_SYSTEMD_INVOCATION_ID="+invocationID, "--no-pager", "-o", "short-iso").Output()
}

func saveJobUpload(fh *multipart.FileHeader) (path, originalName, shaHex string, err error) {
	if fh.Size > maxJobUploadBytes {
		return "", "", "", fmt.Errorf("file too large")
	}
	originalName = filepath.Base(fh.Filename)
	if originalName == "." || originalName == "" {
		return "", "", "", fmt.Errorf("invalid filename")
	}
	if err := os.MkdirAll(jobctl.InputsDir, 0700); err != nil {
		return "", "", "", err
	}
	dstPath := filepath.Join(jobctl.InputsDir, uuid.New().String())
	src, err := fh.Open()
	if err != nil {
		return "", "", "", err
	}
	defer src.Close()

	dst, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0600)
	if err != nil {
		return "", "", "", err
	}
	defer dst.Close()

	h := sha256.New()
	limited := io.LimitReader(src, maxJobUploadBytes+1)
	n, err := io.Copy(io.MultiWriter(dst, h), limited)
	if err != nil {
		os.Remove(dstPath)
		return "", "", "", err
	}
	if n > maxJobUploadBytes {
		os.Remove(dstPath)
		return "", "", "", fmt.Errorf("file too large")
	}
	return dstPath, originalName, hex.EncodeToString(h.Sum(nil)), nil
}
