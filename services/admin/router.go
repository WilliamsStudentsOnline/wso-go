package admin

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	c := NewController(db, cfg, log)

	r.GET("/tasks/task-reference", c.GetTaskReference)
	r.GET("/tasks", c.ListTasks)
	r.GET("/tasks/:taskID", c.GetTask)
	r.POST("/tasks/by-reference/:taskReference/resolve", c.ResolveTasksByReference)
	r.DELETE("/tasks/by-reference/:taskReference", c.DiscardTasksByReference)
}
