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

	goodrichUser := r.Group("")
	goodrichUser.User(auth.RequireScopes(auth.ScopeGoodrichUser))
	// User side
	goodrichUser.POST("/api/v2/goodrich/orders", c.CreateOrder)
	goodrichUser.GET("/api/v2/goodrich/:user/orders", c.ListUserOrders)
	goodrichUser.GET("/api/v2/goodrich/orders/:orderID", c.GetOrder)

	//Admin side
	goodrichAdmin := r.Group("")
	goodrichAdmin.Use(auth.RequireScopes(auth.ScopeGoodrichAdmin))

	goodrichUser.POST("/api/v2/goodrich/orders", c.CreateOrder)
	goodrichUser.GET("/api/v2/goodrich/:user/orders", c.ListUserOrders)
	goodrichAdmin.GET("/api/v2/goodrich/orders", c.ListOrders)
	goodrichAdmin.GET("/api/v2/goodrich/orders/:orderID", c.GetOrder)
	goodrichAdmin.PATCH("/api/v2/goodrich/orders/:orderID", c.UpdateOrder)

}
