package autocomplete

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/autocomplete/mysql"
	"github.com/WilliamsStudentsOnline/wso-go/lib/search"
	"github.com/jinzhu/gorm"
)

type Autocomplete interface {
	AreaOfStudy(q string) ([]string, error)
	Course(q string) ([]string, error)
	Professor(q string) ([]string, error)
	Tag(q string) ([]string, error)
}

func NewAutocomplete(cfg *config.Config, db *gorm.DB) Autocomplete {
	if cfg.SearchBackend == search.SearchBackendSQL {
		return &mysql.Autocomplete{
			DB: db,
		}
	}

	return nil
}
