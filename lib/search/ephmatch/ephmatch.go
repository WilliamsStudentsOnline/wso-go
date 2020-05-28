package ephmatch

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/search"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type SearchEphmatch interface {
	SearchProfiles(query string, profiles *[]*models.EphmatchProfile, selfID uint, opts *models.GetAllProfilesOptions) (err error)
}

func NewSearchEphmatch(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) SearchEphmatch {
	switch cfg.SearchBackend {
	case search.SearchBackendSQL:
		return NewSearchEphmatchMySQL(db, log)
	default:
		return nil
	}
}
