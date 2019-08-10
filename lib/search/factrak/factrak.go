package factrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/search"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
)

type SearchFactrak interface {
	SearchProfessors(query string, users *[]*models.User, opts models.Options) (err error)
	SearchCourses(query string, courses *[]*models.Course, opts models.Options) (err error)
	NewProfessorsOptions(offset *uint, limit *uint, preload *[]string) models.Options
	NewCoursesOptions(offset *uint, limit *uint, preload *[]string) models.Options
}

func NewSearchFactrak(db *gorm.DB, cfg *config.Config) SearchFactrak {
	switch cfg.SearchBackend {
	case search.SearchBackendSQL:
		return NewSearchFactrakMySQL(db)
	default:
		return nil
	}
}
