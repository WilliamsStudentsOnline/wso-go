package admin

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/jobs/jobctl"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	if err := jobctl.ConfigureRedis(cfg.RedisAddr); err != nil {
		log.Debugf("redis unavailable: %v", err)
	}
	c := NewController(db, cfg, log)

	r.GET("/tasks", c.ListTasks)
	r.GET("/tasks/:taskID", c.GetTask)
	r.POST("/tasks/:taskID/resolve", c.ResolveTasks)
	r.DELETE("/tasks/:taskID", c.DeleteTask)

	r.GET("/jobs", c.ListJobs)
	r.GET("/jobs/runs", c.ListJobRuns)
	r.GET("/jobs/runs/:id", c.GetJobRun)
	r.GET("/jobs/runs/:id/logs", c.GetJobRunLogs)
	r.PATCH("/jobs/:name", c.UpdateJob)
	r.POST("/jobs/:name/runs", c.CreateJobRun)
	r.POST("/jobs/runs/:id/cancel", c.CancelJobRun)
}
