package factrak

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/search"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type SearchFactrak interface {
	SearchProfessors(query string, users *[]*models.User, opts SearchOptions) (err error)
	SearchCourses(query string, courses *[]*models.Course, opts SearchCoursesOptions) (err error)
	NewProfessorsOptions(offset *uint, limit *uint, preload []string) SearchOptions
	NewCoursesOptions(offset *uint, limit *uint, preload []string) SearchCoursesOptions
}

func NewSearchFactrak(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) SearchFactrak {
	switch cfg.SearchBackend {
	case search.SearchBackendSQL:
		return NewSearchFactrakMySQL(db)
	default:
		return nil
	}
}

type SearchOptions interface {
	models.Preloader
	models.Paginator
}

type SearchCoursesOptions interface {
	models.Options
	models.Preloader
	models.Paginator
	Post(courses []*models.Course)
}
