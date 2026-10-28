package feedback

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
	writer.Use(auth.RequireScopes(auth.ScopeWriteSelf))
	writer.POST("", c.CreateFeedback)
}

func SetupAdminRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	c := NewController(db, cfg, log)

	r.GET("", c.ListFeedback)
	r.GET("/:feedbackID", c.GetFeedback)
	r.PATCH("/:feedbackID", c.UpdateFeedback)
	r.DELETE("/:feedbackID", c.DeleteFeedback)
}
