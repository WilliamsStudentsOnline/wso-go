package order

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	c := NewController(db, cfg, log)

	// route endpoints to controller functions
	r.POST("/api/v2/goodrich/orders", c.CreateOrder) // VERIFY
	r.GET("/api/v2/goodrich/:user_id/orders", c.ListOrders)
	r.GET("/api/v2/goodrich/orders/:order_id", c.GetOrder)
}
