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

	writer := r.Group("")
	writer.Use(auth.RequireScopes(auth.ScopeBooktrakWrite))

	log.Info("Starting up Google books API for Booktrak search (may fail if there is no network connection to Google)...")

	// Health check endpoint returns false if SetupSearch fails
	r.GET("/health-check", c.HealthCheck)

	err := c.SetupSearch()
	// TODO: In the future, fall back to using only DB for searches. For now, just disable all routes if the API is down.
	if err != nil {
		log.Errorf("Failed to start up Booktrak: %w", err)

		r.GET("/books/web", c.RespondInternalServerError)
		r.GET("/books", c.RespondInternalServerError)
		writer.POST("/books", c.RespondInternalServerError)
		r.GET("/books/:bookID", c.RespondInternalServerError)
		writer.PATCH("/books/:bookID", c.RespondInternalServerError)

		writer.POST("/listings", c.RespondInternalServerError)
		r.GET("/listings", c.RespondInternalServerError)
		r.GET("/listings/:bookListingID", c.RespondInternalServerError)
		writer.PUT("/listings/:bookListingID", c.RespondInternalServerError)
		writer.DELETE("/listings/:bookListingID", c.RespondInternalServerError)
		return
	}

	log.Info("Google books API for Booktrak started successfully")
	r.GET("/books/web", c.SearchBooks)
	r.GET("/books", c.ListBooks)
	writer.POST("/books", c.CreateBook)
	r.GET("/books/:bookID", c.GetBook)
	writer.PATCH("/books/:bookID", c.UpdateBookCourses)

	writer.POST("/listings", c.CreateBookListing)
	r.GET("/listings", c.ListBookListings)
	r.GET("/listings/:bookListingID", c.GetBookListing)
	writer.PUT("/listings/:bookListingID", c.UpdateBookListing)
	writer.DELETE("/listings/:bookListingID", c.DeleteBookListing)
}
