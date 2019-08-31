package autocompletor

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/autocomplete"
	"github.com/WilliamsStudentsOnline/wso-go/lib/autocomplete/mysql"
	"github.com/WilliamsStudentsOnline/wso-go/lib/search"
	"github.com/jinzhu/gorm"
)

func NewAutocomplete(cfg *config.Config, db *gorm.DB) autocomplete.Autocomplete {
	if cfg.SearchBackend == search.SearchBackendSQL {
		return &mysql.Autocomplete{
			DB: db,
		}
	}

	return nil
}
