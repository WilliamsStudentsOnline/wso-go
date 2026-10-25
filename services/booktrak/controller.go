package booktrak

import (
	"context"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
	books "google.golang.org/api/books/v1"
	"google.golang.org/api/option"
)

type Controller struct {
	services.BaseController
	volumeService *books.VolumesService
	// Put models here:
	bookListingModel *models.BookListingModel
	bookModel        *models.BookModel
	courseModel      *models.CourseModel
	apiAvailable     bool
}

func NewController(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) *Controller {
	return &Controller{
		BaseController:   services.BaseController{Log: log},
		bookModel:        models.NewBookModel(db, log),
		bookListingModel: models.NewBookListingModel(db, log),
		courseModel:      models.NewCourseModel(db, log),
		apiAvailable:     false,
	}
}

func (t *Controller) SetupSearch() error {
	ctx := context.Background()
	booksService, err := books.NewService(ctx, option.WithoutAuthentication())
	if err != nil {
		return err
	}
	volumeService := books.NewVolumesService(booksService)
	t.volumeService = volumeService

	// Try a search
	_, err = t.searchVolumes("Programming", lib.IntToPtr(1))
	if err != nil {
		return err
	}
	t.apiAvailable = true
	return nil
}

// Can't use server health check to avoid import cycle
type BooktrakHealthCheckResponse struct {
	OK bool `json:"ok"`
}

func (t *Controller) HealthCheck(c *gin.Context) {
	t.RespondOK(c, BooktrakHealthCheckResponse{OK: t.apiAvailable})
}

func (t *Controller) RespondInternalServerError(c *gin.Context) {
	t.RespondAPIError(c, lib.ErrorInternalServerError)
}

func (t *Controller) RespondGone(c *gin.Context) {
	t.RespondAPIError(c, lib.ErrorBooktrakDeprecated)
}
