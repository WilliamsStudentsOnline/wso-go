package ephmatch

import (
	"strings"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type SearchEphmatchMySQL struct {
	DB  *gorm.DB
	Log *zap.SugaredLogger
}

func NewSearchEphmatchMySQL(db *gorm.DB, log *zap.SugaredLogger) *SearchEphmatchMySQL {
	return &SearchEphmatchMySQL{
		DB:  db,
		Log: log,
	}
}

func (s *SearchEphmatchMySQL) SearchProfiles(query string, profiles *[]*models.EphmatchProfile, selfID uint, opts *models.GetAllProfilesOptions) (err error) {
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
	tx := s.DB.Model(&models.EphmatchProfile{}).
		Joins("INNER JOIN users ON users.id = ephmatch_profiles.user_id").
		Where(sqlSB.String(), sqlParams...)
	tx = tx.Where("users.type = ?", models.UserTypeStudent).
		Preload("User")
	// Run options
	if opts != nil {
		tx = opts.Run(tx)
	}

	tx = tx.Not(models.EphmatchProfile{UserID: selfID})

	err = tx.Find(profiles).Error
	if err != nil {
		return
	}

	epM := models.NewEphmatchProfileModel(s.DB, s.Log)

	if opts != nil {
		// Do recommended sorting algorithm
		if opts.Sort != nil && *opts.Sort == "recommended" {
			err = epM.SortProfilesByLiked(profiles, selfID, opts)
			if err != nil {
				return
			}
		}

		// Populate liked field
		if lib.StringsContains(opts.Preload, "liked") {
			err = epM.PopulateLiked(profiles, selfID)
			if err != nil {
				return
			}
		}

		// Populate matched field
		if lib.StringsContains(opts.Preload, "matched") {
			err = epM.PopulateMatched(profiles, selfID)
			if err != nil {
				return
			}
		}
	}

	return
}
