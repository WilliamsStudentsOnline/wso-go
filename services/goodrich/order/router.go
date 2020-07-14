package order

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

//SetupRouter defines routes endpoints to functions and requires appropriate scopes
func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	c := NewController(db, cfg, log)

	// User side
	goodrichUser := r.Group("")
	goodrichUser.Use(auth.RequireScopes(auth.ScopeGoodrichUser, auth.ScopeGoodrichAdmin))
	goodrichUser.POST("/api/v2/goodrich/orders", c.CreateOrder)
	goodrichUser.GET("/api/v2/goodrich/:user/orders", c.ListUserOrders)
	goodrichUser.GET("/api/v2/goodrich/orders/:orderID", c.GetOrder)

	//Admin side
	goodrichAdmin := r.Group("")
	goodrichAdmin.Use(auth.RequireScopes(auth.ScopeGoodrichAdmin))
	goodrichAdmin.GET("/api/v2/goodrich/orders", c.ListOrders)
	goodrichUser.GET("/api/v2/goodrich/orders/:orderID", c.GetOrderAdmin)
	goodrichAdmin.PATCH("/api/v2/goodrich/orders/:orderID", c.UpdateOrder)

}
