package ephmatch

import (
	"strings"

	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
)

type SearchEphmatchMySQL struct {
	DB *gorm.DB
}

func NewSearchEphmatchMySQL(db *gorm.DB) *SearchEphmatchMySQL {
	return &SearchEphmatchMySQL{
		DB: db,
	}
}

func (s *SearchEphmatchMySQL) SearchProfiles(query string, profiles *[]*models.EphmatchProfile, opts SearchOptions) (err error) {
	words := strings.Split(strings.ToLower(query), " ")

	// Generate an SQL query with each word being different
	sqlSB := new(strings.Builder)
	sqlParams := make([]interface{}, len(words))

	for i := range words {
		sqlSB.WriteString("lower(users.name) LIKE ?")
		sqlParams[i] = "%" + words[i] + "%"
		if i+1 < len(words) {
			sqlSB.WriteString(" OR ")
		}
	}

	// Do SQL
	tx := s.DB.Model(&models.EphmatchProfile{}). //.Model(&models.User{})
							Joins("JOIN users ON users.id = ephmatch_profiles.user_id").
							Where(sqlSB.String(), sqlParams...)
	tx = tx.Where("users.at_williams = ? AND users.type = ?", true, models.UserTypeProfessor)
	// Run options
	if opts != nil {
		tx = opts.Paginate(tx)
		tx = opts.Preloader(tx)
	}

	err = tx.Find(users).Error
	return
}
