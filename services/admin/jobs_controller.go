package admin

import (
	"github.com/WilliamsStudentsOnline/wso-go/jobs"
	_ "github.com/WilliamsStudentsOnline/wso-go/services"
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
// @Success 200 {object} k8sJobStatus "Return Job Status from Kubernetes"
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /admin/jobs/{jobID}/status [get]
func (t *Controller) GetJobStatus(c *gin.Context) {
	job, err := jobs.GetJob(t.cfg, c.Param("jobID"))

	if err != nil {
		t.RespondError(c, err)
		return
	}

	// TODO: do not return external types
	// sanitize the output and use our own struct
	t.RespondOK(c, job.Status)
}
