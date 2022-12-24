package booktrak

import (
	"context"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
	"google.golang.org/api/books/v1"
	"google.golang.org/api/option"
)

type Controller struct {
	services.BaseController
	volumeService *books.VolumesService
}

func NewController(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) *Controller {
	return &Controller{
		BaseController: services.BaseController{Log: log},
	}
}

func (t *Controller) SetupSearch() error {
	ctx := context.Background()
	booksService, err := books.NewService(ctx, option.WithoutAuthentication())
	if err != nil {
		return err
	}
	VolumeService := books.NewVolumesService(booksService)
	t.volumeService = VolumeService

	return nil
}

type SearchBooksRequest struct {
	Query string `form:"q"`
}

func (t *Controller) SearchBooks(c *gin.Context) {
	var err error
	request := SearchBooksRequest{}
	if err = c.ShouldBindQuery(&request); err != nil {
		t.RespondBadBind(c, err)
		return
	}

	volumes, err := t.volumeService.
		List(request.Query).
		MaxResults(10).
		Fields("items/volumeInfo/authors",
			"items/volumeInfo/imageLinks",
			"items/volumeInfo/industryIdentifiers",
			"items/volumeInfo/infoLink",
			"items/volumeInfo/title",
			"items/volumeInfo/subtitle",
			"items/volumeInfo/publisher").
		Do()

	if err != nil {
		t.RespondAPIError(c, lib.ErrorInternalServerError)
		return
	}

	t.RespondOK(c, volumes)
}
