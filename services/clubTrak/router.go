package clubTrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// SetupRouter sets up the router for Bulletins
func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	c := NewController(db, cfg, log)

	writer := r.Group("")
	// TODO: Deprecate auth.ScopeWriteSelf from here
	writer.Use(auth.RequireScopes(auth.ScopeWriteSelf, auth.ScopeBulletinWrite))

	// Clubtrak CRUD endpoints
	r.POST("/clubTrak", c.AddClub)

}
