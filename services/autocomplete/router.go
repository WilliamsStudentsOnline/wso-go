package autocomplete

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// SetupRouter sets up the router for Bulletins
func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config) {
	c := NewController(db, cfg)

	// Ensure factrak or factrak limited
	factrak := r.Group("")
	factrak.Use(auth.RequireScopes(auth.ScopeFactrakLimited, auth.ScopeFactrakFull))
	factrak.GET("/area-of-study", c.AreaOfStudy)
	factrak.GET("/course", c.Course)
	factrak.GET("/professor", c.Professor)
	factrak.GET("/factrak", c.Factrak)

	// Ensure user
	r.GET("/tag", auth.RequireScopes(auth.ScopeUsers), c.Tag)
}
