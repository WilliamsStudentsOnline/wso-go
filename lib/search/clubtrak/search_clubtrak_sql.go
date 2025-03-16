package clubtrak

import (
	"strings"

	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
)

type SearchClubtrakMySQL struct {
	DB *gorm.DB
}

func NewSearchClubtrakMySQL(db *gorm.DB) *SearchClubtrakMySQL {
	return &SearchClubtrakMySQL{
		DB: db,
	}
}

func (s *SearchClubtrakMySQL) SearchClubs(query string, clubs *[]*models.Club, opts SearchOptions) (err error) {
	// Do SQL
	tx := s.DB.Model(&models.Club{})
	tx = tx.Where("lower(clubs.name) LIKE ?", "%"+strings.ToLower(query)+"%")
	// Run options
	if opts != nil {
		tx = opts.Paginate(tx)

		//Not using a preloader for now
		//tx = opts.Preloader(tx)
	}

	err = tx.Find(clubs).Error
	return
}

func (*SearchClubtrakMySQL) NewClubsOptions(offset *uint, limit *uint, preload []string) SearchOptions {
	return &models.GetAllClubsOptions{
		Offset: offset,
		Limit:  limit,
	}
}
