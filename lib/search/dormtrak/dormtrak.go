package dormtrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/search"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
)

type SearchDormtrak interface {
	SearchDorms(query string, dorms *[]*models.Dorm, opts models.Options) (err error)
	NewDormsOptions(offset *uint, limit *uint, preload *[]string) models.Options
}

func NewSearchDormtrak(db *gorm.DB, cfg *config.Config) SearchDormtrak {
	switch cfg.SearchBackend {
	case search.SearchBackendSQL:
		return NewSearchDormtrakMySQL(db)
	default:
		return nil
	}
}
