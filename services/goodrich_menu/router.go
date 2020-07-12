package goodrich_menu

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	c := NewController(db, cfg, log)

	//scope user
	goodrichUser := r.Group("")
	goodrichUser.Use(auth.RequireScopes(auth.ScopeGoodrichOrder))

	goodrichUser.GET("/api/v2/goodrich/menu", c.ListMenuItems)
	goodrichUser.GET("/api/v2/goodrich/menu/:itemID", c.GetMenuItem)


	//scope goodrich admin
	goodrichAdmin := r.Group("")
	goodrichAdmin.Use(auth.RequireScopes(auth.ScopeGoodrichAdmin))

	goodrichAdmin.GET("/api/v2/goodrich/menu", c.ListMenuItems)
	goodrichAdmin.GET("/api/v2/goodrich/menu/:itemID", c.GetMenuItem)
	goodrichAdmin.POST("/api/v2/goodrich/menu", c.CreateMenuItem)
	goodrichAdmin.PATCH("/api/v2/goodrich/menu/:itemID", c.UpdateMenuItem)

}
