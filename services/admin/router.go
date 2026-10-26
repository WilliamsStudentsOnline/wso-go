package admin

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	c := NewController(db, cfg, log)

	r.GET("/tasks", c.ListTasks)
	r.GET("/tasks/:taskID", c.GetTask)
	r.POST("/tasks/:taskID/resolve", c.ResolveTasks)
	r.DELETE("/tasks/:taskID", c.DeleteTask)
}
