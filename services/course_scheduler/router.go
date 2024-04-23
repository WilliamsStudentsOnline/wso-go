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

	students := r.Group("")
	students.Use(auth.RequireScopes(auth.ScopeCourseSchedulerFull, auth.ScopeCourseSchedulerAdmin))

	students.GET("/selections", c.ListCourseSchedulerSelections)
	students.POST("/selections", c.CreateCourseSchedulerSelection)
	students.DELETE("/selections/:selectionID", c.DeleteCourseSchedulerSelections)
	students.PATCH("/selections/:selectionID", c.UpdateHiddenCourseSchedulerSelection)

}
