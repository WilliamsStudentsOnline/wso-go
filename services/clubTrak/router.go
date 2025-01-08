package clubTrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// SetupRouter sets up the router for clubs
func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	c := NewController(db, cfg, log)

	// Clubtrak CRUD endpoints
	r.POST("/addClub", c.AddClub)
	r.GET("/testing", c.TestingClub)

}
