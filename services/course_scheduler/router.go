package course_scheduler

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	c := NewController(db, cfg, log)

	r.GET("/selections", c.ListCourseSchedulerSelections)
	r.POST("/selections", c.CreateCourseSchedulerSelection)
	r.DELETE("/selections/:selectionID", c.DeleteCourseSchedulerSelections)
	r.PATCH("/selections/:selectionID", c.UpdateHiddenCourseSchedulerSelection)
}
