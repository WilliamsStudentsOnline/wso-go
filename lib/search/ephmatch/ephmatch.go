package ephmatch

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/search"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type SearchEphmatch interface {
	SearchProfiles(query string, profiles *[]*models.EphmatchProfile, opts SearchOptions) (err error)
	NewProfilesOptions(offset *uint, limit *uint, preload []string) SearchOptions
}

func NewSearchEphmatch(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) SearchEphmatch {
	switch cfg.SearchBackend {
	case search.SearchBackendSQL:
		return NewSearchEphmatchMySQL(db)
	default:
		return nil
	}
}

type SearchOptions interface {
	models.Preloader
	models.Paginator
}
