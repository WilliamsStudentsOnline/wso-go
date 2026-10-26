package admin

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	c := NewController(db, cfg, log)

	// Register static paths before :taskID
	r.GET("/tasks/subject-key", c.GetSubjectKey)
	r.GET("/tasks", c.ListTasks)
	r.GET("/tasks/:taskID", c.GetTask)
	r.POST("/tasks/by-subject/:subjectKey/resolve", c.ResolveTasksBySubject)
	r.DELETE("/tasks/by-subject/:subjectKey", c.DiscardTasksBySubject)
}
