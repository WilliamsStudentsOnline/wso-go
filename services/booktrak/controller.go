package booktrak

import (
	"context"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
	"google.golang.org/api/books/v1"
	"google.golang.org/api/option"
)

type Controller struct {
	services.BaseController
	volumeService *books.VolumesService
	// Put models here:
	bookListingModel *models.BookListingModel
	bookModel        *models.BookModel
	courseModel      *models.CourseModel
	// userModel        *models.UserModel
}

func NewController(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) *Controller {
	return &Controller{
		BaseController:   services.BaseController{Log: log},
		bookModel:        models.NewBookModel(db, log),
		bookListingModel: models.NewBookListingModel(db, log),
		courseModel:      models.NewCourseModel(db, log),
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

	return nil
}
