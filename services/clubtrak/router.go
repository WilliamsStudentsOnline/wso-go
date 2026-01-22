package clubtrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// SetupRouter sets up the router for clubtrak
func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	c := NewController(db, cfg, log)

	//User Scope Endpoints
	userGroup := r.Group("/clubs")
	userGroup.Use(auth.RequireScopes(auth.ScopeUsers))
	userGroup.GET("", c.GetAllClubs)
	userGroup.DELETE("/:clubID", c.DeleteClub)
	userGroup.PATCH("/:clubID", c.UpdateClub)

	//Admin Endpoints
	adminGroup := r.Group("/admin/clubs")
	adminGroup.Use(auth.RequireScopes(auth.ScopeAdminAll))
	adminGroup.POST("", c.CreateClub)
	adminGroup.GET("", c.AdminGetAllClubs)
	adminGroup.DELETE("/:clubID", c.AdminDeleteClub)
	adminGroup.PATCH("/:clubID", c.AdminUpdateClub)

}
