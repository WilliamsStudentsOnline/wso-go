package catalog_update

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// ArchiveRawCatalog writes the raw catalog JSON for a year into archiveDir as raw-{year}.json.
// year 0 is archived as raw-current.json.
func ArchiveRawCatalog(archiveDir string, year int, data []byte) error {
	if archiveDir == "" {
		return fmt.Errorf("archive dir is empty")
	}
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return err
	}
	name := "raw-current.json"
	if year != 0 {
		name = fmt.Sprintf("raw-%d.json", year)
	}
	path := filepath.Join(archiveDir, name)
	return os.WriteFile(path, data, 0644)
}

// EarliestCatalogYear is the oldest spring calendar year the Williams catalog feed serves.
const EarliestCatalogYear = 2023

// ReadCatalogYearFile loads raw catalog JSON for a year from catalogDir.
// Tries raw-{year}.json, then year_{year}.json.
func ReadCatalogYearFile(catalogDir string, year int) ([]byte, error) {
	if catalogDir == "" {
		return nil, fmt.Errorf("catalog dir is empty")
	}
	candidates := []string{
		filepath.Join(catalogDir, fmt.Sprintf("raw-%d.json", year)),
		filepath.Join(catalogDir, fmt.Sprintf("year_%d.json", year)),
	}
	var errs []string
	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err == nil {
			return data, nil
		}
		errs = append(errs, fmt.Sprintf("%s: %v", path, err))
	}
	return nil, fmt.Errorf("catalog file for year %d not found in %s (%s)", year, catalogDir, strings.Join(errs, "; "))
}

// IngestCatalog upserts raw catalog rows into the canonical course tables for one feed year.
// Re-running on the same data is idempotent (same offering/listing counts).
func IngestCatalog(db *gorm.DB, rawCourses []RawCourse, log *zap.SugaredLogger) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}
	if log == nil {
		log = zap.NewNop().Sugar()
	}

	// Cache unix_id -> user id lookups for this run (name matching comes in a later change).
	userByUnix := make(map[string]*uint)

	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	for i := range rawCourses {
		if err := upsertRawCourse(tx, &rawCourses[i], userByUnix); err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}
	log.Infof("Ingested %d catalog rows into DB", len(rawCourses))
	return nil
}

func upsertRawCourse(db *gorm.DB, raw *RawCourse, userByUnix map[string]*uint) error {
	crseID := strings.TrimSpace(raw.CourseID)
	if crseID == "" {
		return nil
	}

	title := strings.TrimSpace(raw.CourseTitleLong)
	if title == "" {
		title = strings.TrimSpace(raw.Description)
	}
	description := strings.TrimSpace(raw.DescriptionSearch)
	if description == "" {
		description = strings.TrimSpace(raw.Description)
	}

	course, err := upsertCanonicalCourse(db, crseID, title, description, raw.AcademicYear)
	if err != nil {
		return err
	}

	subject := strings.TrimSpace(raw.Subject)
	letter := strings.TrimSpace(raw.CourseLetter)
	if err := upsertListing(db, course.ID, subject, raw.CatalogNumber, letter, raw.AcademicYear); err != nil {
		return err
	}

	if raw.ClassNumber == 0 {
		// Catalog entries without a class number are not schedulable offerings.
		return nil
	}

	offering, err := upsertOffering(db, course.ID, raw, title)
	if err != nil {
		return err
	}

	if err := replaceMeetings(db, offering.ID, raw); err != nil {
		return err
	}
	return replaceInstructors(db, offering.ID, raw, userByUnix)
}

func upsertCanonicalCourse(db *gorm.DB, crseID, title, description string, year int) (*models.CourseCanonical, error) {
	var course models.CourseCanonical
	err := db.Where("crse_id = ?", crseID).First(&course).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	if gorm.IsRecordNotFoundError(err) {
		course = models.CourseCanonical{
			CrseID:      crseID,
			Title:       title,
			Description: description,
		}
		if err := db.Create(&course).Error; err != nil {
			return nil, err
		}
		return &course, nil
	}

	// Keep courses_canonical title/description as the newest academic year seen.
	var maxYear sql.NullInt64
	row := db.Model(&models.Offering{}).
		Where("course_canonical_id = ?", course.ID).
		Select("MAX(year)").
		Row()
	if err := row.Scan(&maxYear); err != nil {
		return nil, err
	}
	if !maxYear.Valid || year >= int(maxYear.Int64) {
		course.Title = title
		course.Description = description
		if err := db.Save(&course).Error; err != nil {
			return nil, err
		}
	}
	return &course, nil
}

func upsertListing(db *gorm.DB, courseID uint, subject string, number int, letter string, year int) error {
	if subject == "" {
		return nil
	}

	var listing models.CourseListing
	err := db.Where(
		"course_canonical_id = ? AND subject = ? AND number = ? AND letter = ?",
		courseID, subject, number, letter,
	).First(&listing).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return err
	}

	if gorm.IsRecordNotFoundError(err) {
		listing = models.CourseListing{
			CourseCanonicalID: courseID,
			Subject:           subject,
			Number:            number,
			Letter:            letter,
			FirstValidYear:    year,
			LastValidYear:     year,
		}
		return db.Create(&listing).Error
	}

	changed := false
	if year < listing.FirstValidYear {
		listing.FirstValidYear = year
		changed = true
	}
	if year > listing.LastValidYear {
		listing.LastValidYear = year
		changed = true
	}
	if changed {
		return db.Save(&listing).Error
	}
	return nil
}

func upsertOffering(db *gorm.DB, courseID uint, raw *RawCourse, title string) (*models.Offering, error) {
	status := models.OfferingStatusOffered
	if strings.TrimSpace(raw.Offered) != "Y" {
		status = models.OfferingStatusNotOffered
	}
	if strings.EqualFold(strings.TrimSpace(raw.Facility1), "Cancelled") {
		status = models.OfferingStatusCancelled
	}

	component := strings.TrimSpace(raw.SSRComponent)
	section := strings.TrimSpace(raw.ClassSection)
	term := TermFromStrm(raw.Semester)

	var offering models.Offering
	err := db.Where("strm = ? AND class_nbr = ?", raw.Semester, raw.ClassNumber).First(&offering).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	if gorm.IsRecordNotFoundError(err) {
		offering = models.Offering{
			CourseCanonicalID: courseID,
			Strm:              raw.Semester,
			Year:              raw.AcademicYear,
			Term:              term,
			Section:           section,
			ClassNbr:          raw.ClassNumber,
			Title:             title,
			Component:         component,
			Status:            status,
		}
		if err := db.Create(&offering).Error; err != nil {
			return nil, err
		}
		return &offering, nil
	}

	offering.CourseCanonicalID = courseID
	offering.Year = raw.AcademicYear
	offering.Term = term
	offering.Section = section
	offering.Title = title
	offering.Component = component
	offering.Status = status
	if err := db.Save(&offering).Error; err != nil {
		return nil, err
	}
	return &offering, nil
}

func replaceMeetings(db *gorm.DB, offeringID uint, raw *RawCourse) error {
	if err := db.Unscoped().Where("offering_id = ?", offeringID).Delete(&models.OfferingMeeting{}).Error; err != nil {
		return err
	}

	slots := []struct {
		days, start, end, facility string
	}{
		{raw.StandardMeeting1, raw.StartTime1, raw.EndTime1, raw.Facility1},
		{raw.StandardMeeting2, raw.StartTime2, raw.EndTime2, raw.Facility2},
		{raw.StandardMeeting3, raw.StartTime3, raw.EndTime3, raw.Facility3},
	}

	for _, slot := range slots {
		days := strings.TrimSpace(slot.days)
		if days == "" {
			continue
		}
		start := strings.TrimSpace(slot.start)
		end := strings.TrimSpace(slot.end)
		facility := strings.TrimSpace(slot.facility)
		if start != "" {
			if t, err := time.Parse(hourFormat, start); err == nil {
				start = t.Format(hourFormat)
			}
		}
		if end != "" {
			if t, err := time.Parse(hourFormat, end); err == nil {
				end = t.Format(hourFormat)
			}
		}
		meeting := models.OfferingMeeting{
			OfferingID: offeringID,
			Days:       days,
			Start:      start,
			End:        end,
			Facility:   facility,
		}
		if err := db.Create(&meeting).Error; err != nil {
			return err
		}
	}
	return nil
}

func replaceInstructors(db *gorm.DB, offeringID uint, raw *RawCourse, userByUnix map[string]*uint) error {
	if err := db.Unscoped().Where("offering_id = ?", offeringID).Delete(&models.OfferingInstructor{}).Error; err != nil {
		return err
	}

	slots := []struct {
		first, middle, last, unix string
	}{
		{raw.FirstName1, raw.MiddleName1, raw.LastName1, raw.UnixID1},
		{raw.FirstName2, raw.MiddleName2, raw.LastName2, raw.UnixID2},
		{raw.FirstName3, raw.MiddleName3, raw.LastName3, raw.UnixID3},
		{raw.FirstName4, raw.MiddleName4, raw.LastName4, raw.UnixID4},
		{raw.FirstName5, raw.MiddleName5, raw.LastName5, raw.UnixID5},
		{raw.FirstName6, raw.MiddleName6, raw.LastName6, raw.UnixID6},
	}

	for _, slot := range slots {
		fn := strings.TrimSpace(slot.first)
		mn := strings.TrimSpace(slot.middle)
		ln := strings.TrimSpace(slot.last)
		unixID := strings.TrimSpace(slot.unix)
		if fn == "" && unixID == "" {
			break
		}

		name := fn
		if mn != "" {
			if name != "" {
				name += " "
			}
			name += mn
		}
		if ln != "" {
			if name != "" {
				name += " "
			}
			name += ln
		}
		if name == "" {
			name = unixID
		}

		var userID *uint
		if unixID != "" {
			if cached, ok := userByUnix[unixID]; ok {
				userID = cached
			} else {
				var user models.User
				err := db.Where("unix_id = ?", unixID).First(&user).Error
				if err != nil && !gorm.IsRecordNotFoundError(err) {
					return err
				}
				if err == nil {
					id := user.ID
					userID = &id
					userByUnix[unixID] = userID
				} else {
					userByUnix[unixID] = nil
				}
			}
		}

		instr := models.OfferingInstructor{
			OfferingID:   offeringID,
			UnixID:       unixID,
			UserID:       userID,
			NameAsListed: name,
		}
		if err := db.Create(&instr).Error; err != nil {
			return err
		}
	}
	return nil
}
