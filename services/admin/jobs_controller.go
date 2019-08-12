package admin

import (
	"github.com/WilliamsStudentsOnline/wso-go/jobs"
	"github.com/gin-gonic/gin"
)

// GetJobStatus godoc
// @Summary Gets status of a job
// @Description connects to Kubernetes to get the status of a running (or recently finished) job given an id
// @ID get-job-status
// @Tags admin
// @Accept  json
// @Produce  json
// @Param jobID path string true "Kubernetes Job ID"
// @Success 200 {object} k8s.io/api/batch/v1.JobStatus
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /admin/jobs/{jobID}/status [get]
func (t *Controller) GetJobStatus(c *gin.Context) {
	job, err := jobs.GetJob(t.cfg, c.Param("jobID"))

	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, job.Status)
}
