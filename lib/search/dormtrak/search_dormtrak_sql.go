package dormtrak

import (
	"strings"

	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
)

type SearchDormtrakMySQL struct {
	DB *gorm.DB
}

func NewSearchDormtrakMySQL(db *gorm.DB) *SearchDormtrakMySQL {
	return &SearchDormtrakMySQL{
		DB: db,
	}
}

func (s *SearchDormtrakMySQL) SearchDorms(query string, dorms *[]*models.Dorm, opts SearchOptions) (err error) {
	// Do SQL
	tx := s.DB.Model(&models.Dorm{})
	tx = tx.Where("lower(dorms.name) LIKE ?", "%"+strings.ToLower(query)+"%")
	tx = models.NewDormModel(nil, nil).ScopeTrakked(tx)
	// Run options
	if opts != nil {
		tx = opts.Paginate(tx)
		tx = opts.Preloader(tx)
	}

	err = tx.Find(dorms).Error
	return
}

func (*SearchDormtrakMySQL) NewDormsOptions(offset *uint, limit *uint, preload []string) SearchOptions {
	return &models.GetAllDormsOptions{
		Offset:  offset,
		Limit:   limit,
		Preload: preload,
	}
}
