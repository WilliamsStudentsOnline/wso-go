package user_home_info

import (
	"fmt"

	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// Stats summarizes an ingest run.
type Stats struct {
	Updated   int
	NotFound  int
	OffCampus int
}

// IngestRecords updates existing users from registrar WSO info rows.
//
// Updates home address fields always (blank clears), SU box / entry when
// non-empty (does not wipe existing values with blanks), and sets
// off_campus when DORM is "Off-Campus".
//
// Does not assign dorm rooms: the dump has building/entry but not room
// numbers for typical housing.
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

		updates := map[string]interface{}{
			"home_town":    NullableString(rec.HomeTown),
			"home_state":   NullableString(rec.HomeState),
			"home_country": NullableString(rec.HomeCountry),
			"home_zip":     NullableString(rec.HomeZip),
		}

		if rec.SUBox != "" {
			updates["su_box"] = rec.SUBox
		}
		if rec.Entry != "" {
			updates["entry"] = rec.Entry
		}
		if rec.Dorm == "Off-Campus" {
			updates["off_campus"] = true
			stats.OffCampus++
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
