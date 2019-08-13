package dormtrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/search"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
)

type SearchDormtrak interface {
	SearchDorms(query string, dorms *[]*models.Dorm, opts SearchOptions) (err error)
	NewDormsOptions(offset *uint, limit *uint, preload *[]string) SearchOptions
}

func NewSearchDormtrak(db *gorm.DB, cfg *config.Config) SearchDormtrak {
	switch cfg.SearchBackend {
	case search.SearchBackendSQL:
		return NewSearchDormtrakMySQL(db)
	default:
		return nil
	}
}

type SearchOptions interface {
	models.Preloader
	models.Paginator
}
