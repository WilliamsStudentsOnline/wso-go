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
	gone := c.RespondGone

	writer := r.Group("")
	writer.Use(auth.RequireScopes(auth.ScopeBooktrakWrite))

	r.GET("/health-check", gone)

	r.GET("/books/web", gone)
	r.GET("/books", gone)
	writer.POST("/books", gone)
	r.GET("/books/:bookID", gone)
	writer.PATCH("/books/:bookID", gone)

	writer.POST("/listings", gone)
	r.GET("/listings", gone)
	r.GET("/listings/:bookListingID", gone)
	writer.PUT("/listings/:bookListingID", gone)
	writer.DELETE("/listings/:bookListingID", gone)
}
