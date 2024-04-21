package course_scheduler

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	c := NewController(db, cfg, log)

	writer := r.Group("")
	writer.Use(auth.RequireScopes(auth.ScopeCourseSchedulerAdmin))

	r.GET("/selections", c.ListCourseSchedulerSelections)

	writer.POST("/selections", c.CreateCourseSchedulerSelection)
	writer.DELETE("/selections/:selectionID", c.DeleteCourseSchedulerSelections)
	writer.PATCH("/selections/:selectionID", c.UpdateHiddenCourseSchedulerSelection)
}
