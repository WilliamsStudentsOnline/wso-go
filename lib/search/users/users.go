package users

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/search"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
)

type SearchUsers interface {
	Search(query string, users *[]*models.User, opts SearchOptions) (err error)
	NewOptions(offset *uint, limit *uint, preload *[]string) SearchOptions
}

func NewSearchUsersMySQL(db *gorm.DB) *SearchUsersMySQL {
	return &SearchUsersMySQL{
		DB: db,
	}
}

func NewSearchUsers(db *gorm.DB, cfg *config.Config) SearchUsers {
	switch cfg.SearchBackend {
	case search.SearchBackendSQL:
		return NewSearchUsersMySQL(db)
	default:
		return nil
	}
}

type SearchOptions interface {
	models.Preloader
	models.Paginator
}
