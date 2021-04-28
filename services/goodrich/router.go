package goodrich

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	c := NewController(db, cfg, log)

	manager := r.Group("")
	manager.Use(auth.RequireScopes(auth.ScopeGoodrichManager))

	r.GET("/timeslots", c.ListTimeSlots)

	// Menu Service
	r.GET("/menu", c.ListMenu)
	//TODO[low]: r.GET("/menu/:itemID", c.GetMenuItem)

	// Manager Menu Service
	//manager.POST("/menu", c.CreateMenuItem)
	//manager.PATCH("/menu/:itemID", c.UpdateMenuItem)

	// Order Service
	r.GET("/user/orders", c.ListUserOrders)
	r.GET("/user/orders/:orderID", c.GetUserOrder)
	r.POST("/orders", c.CreateOrder)

	// Manager Order Service
	manager.GET("/orders", c.ListOrders)
	manager.GET("/orders/:orderID", c.GetOrder)
	manager.PATCH("/orders/:orderID", c.UpdateOrder)
}
