package booktrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

func SetupRouter(r gin.IRouter, db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) {
	c := NewController(db, cfg, log)
	// Scoping
	full := r.Group("")
	// TODO: Change this scope to a specific booktrak scope
	full.Use(auth.RequireScopes(auth.ScopeFactrakFull))

	err := c.SetupSearch()
	if err != nil {
		log.Errorf("Failed to start up booktrak search: %w", err)
	} else {
		r.GET("/books/web", c.SearchBooks)
	}
	r.POST("/books", c.CreateOrUpdateBook)
	r.GET("/books", c.ListBooks)

	r.POST("/listings", c.CreateBookListing)
	r.GET("/listings", c.ListBookListings)
	r.DELETE("/listings/:bookListingID", c.DeleteBookListing)
}
