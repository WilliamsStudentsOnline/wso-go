package user_home_info

import (
	"fmt"

	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type Stats struct {
	Updated  int
	NotFound int
	Skipped  int
}

// IngestRecords updates existing users from registrar rows; only non-empty fields; DORM is not mapped to off_campus
func IngestRecords(db *gorm.DB, userModel *models.UserModel, log *zap.SugaredLogger, records []Record) (Stats, error) {
	var stats Stats

	for _, rec := range records {
		var user models.User
		err := db.Where("unix_id = ?", rec.UnixID).First(&user).Error
		if gorm.IsRecordNotFoundError(err) {
			stats.NotFound++
			log.Warnf("user %s not found, skipping", rec.UnixID)
			continue
		}
		if err != nil {
			return stats, fmt.Errorf("lookup %s: %w", rec.UnixID, err)
		}

		updates := map[string]interface{}{}
		if rec.HomeTown != "" {
			updates["home_town"] = rec.HomeTown
		}
		if rec.HomeState != "" {
			updates["home_state"] = rec.HomeState
		}
		if rec.HomeCountry != "" {
			updates["home_country"] = rec.HomeCountry
		}
		if rec.HomeZip != "" {
			updates["home_zip"] = rec.HomeZip
		}
		if rec.SUBox != "" {
			updates["su_box"] = rec.SUBox
		}
		if rec.Entry != "" {
			updates["entry"] = rec.Entry
		}

		if len(updates) == 0 {
			stats.Skipped++
			continue
		}

		if err := db.Model(&user).Updates(updates).Error; err != nil {
			return stats, fmt.Errorf("update %s: %w", rec.UnixID, err)
		}
		if err := userModel.PopulateSearchFields(user.ID); err != nil {
			return stats, fmt.Errorf("search fields %s: %w", rec.UnixID, err)
		}

		stats.Updated++
		log.Infof("updated %s", rec.UnixID)
	}

	return stats, nil
}
