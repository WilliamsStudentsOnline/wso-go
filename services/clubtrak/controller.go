package clubtrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	search "github.com/WilliamsStudentsOnline/wso-go/lib/search/clubtrak"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type Controller struct {
	services.BaseController
	// Put models here:
	clubModel      *models.ClubModel
	userModel      *models.UserModel
	clubtrakSearch search.SearchClubtrak
}

func NewController(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) *Controller {
	return &Controller{
		BaseController: services.BaseController{Log: log},
		clubModel:      models.NewClubModel(db, log),
		userModel:      models.NewUserModel(db, log),
		clubtrakSearch: search.NewSearchClubtrak(db, cfg, log),
	}
}
