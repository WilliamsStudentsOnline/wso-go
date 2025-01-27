package clubtrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/search"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type SearchClubtrak interface {
	SearchClubs(query string, clubs *[]*models.Club, opts SearchOptions) (err error)
	NewClubsOptions(offset *uint, limit *uint, preload []string) SearchOptions
}

func NewSearchClubtrak(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) SearchClubtrak {
	switch cfg.SearchBackend {
	case search.SearchBackendSQL:
		return NewSearchClubtrakMySQL(db)
	default:
		return nil
	}
}

type SearchOptions interface {
	models.Preloader
	models.Paginator
}
