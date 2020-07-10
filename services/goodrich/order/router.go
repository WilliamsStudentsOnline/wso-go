package order

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	c := NewController(db, cfg, log)

	goodrich_user := r.Group("")
	goodrich_user.User(auth.RequireScopes(auth.ScopeGoodrichUser))
	// User side
	goodrich_user.POST("/api/v2/goodrich/orders", c.CreateOrder)
	goodrich_user.GET("/api/v2/goodrich/:user/orders", c.ListUserOrders)
	goodrich_user.GET("/api/v2/goodrich/orders/:orderID", c.GetOrder)


	//Admin side
	goodrich_admin := r.Group("")
	goodrich_admin.Use(auth.RequireScopes(auth.ScopeGoodrichAdmin))

	goodrich_admin.GET("/api/v2/goodrich/orders", c.ListOrders)
	goodrich_admin.GET("/api/v2/goodrich/orders/:orderID", c.GetOrder)
	goodrich_admin.PATCH("/api/v2/goodrich/orders/:orderID", c.UpdateOrder)

}
