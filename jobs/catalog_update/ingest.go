package catalog_update

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

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

// Oldest spring calendar year the Williams catalog feed serves
const EarliestCatalogYear = 2023

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

func IngestCatalog(db *gorm.DB, rawCourses []RawCourse, log *zap.SugaredLogger) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}
	if log == nil {
		log = zap.NewNop().Sugar()
	}

	matcher, err := NewInstructorMatcher(db)
	if err != nil {
		return fmt.Errorf("build instructor matcher: %w", err)
	}

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

	strms := map[int]struct{}{}
	for i := range rawCourses {
		if err := upsertRawCourse(tx, &rawCourses[i], matcher); err != nil {
			tx.Rollback()
			return err
		}
		strms[rawCourses[i].Semester] = struct{}{}
	}
	if err := unionCourseInSemesterAttributes(tx, strms); err != nil {
		return err
	}
	if err := refreshCourseInSemesterStatus(tx, strms); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}
	log.Infof("Ingested %d catalog rows into DB", len(rawCourses))
	return nil
}

func upsertRawCourse(db *gorm.DB, raw *RawCourse, matcher *InstructorMatcher) error {
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
		// No class number: not a schedulable offering
		return nil
	}
	if strings.TrimSpace(raw.Offered) != "Y" {
		// Not-offered rows reuse the strm and class number of the last real section, so never overwrite it
		return recoverLastOffered(db, course.ID, raw, title, matcher)
	}

	cis, err := upsertCourseInSemester(db, course.ID, raw, title, description)
	if err != nil {
		return err
	}
	offering, err := upsertOffering(db, course.ID, cis.ID, raw, title)
	if err != nil {
		return err
	}

	if err := replaceMeetings(db, offering.ID, raw); err != nil {
		return err
	}
	return replaceInstructors(db, offering.ID, raw, matcher)
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

	var maxYear sql.NullInt64
	row := db.Model(&models.CourseOffering{}).
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

// 1YYT, e.g. 1271 is Fall of catalog year 2027 (Fall 2026)
func YearFromStrm(strm int) (int, bool) {
	if strm < 1000 || strm > 1999 {
		return 0, false
	}
	return 2000 + (strm/10)%100, true
}

func upsertOffering(db *gorm.DB, courseID, courseInSemesterID uint, raw *RawCourse, title string) (*models.CourseOffering, error) {
	status := models.CourseOfferingStatusOffered
	if strings.EqualFold(strings.TrimSpace(raw.Facility1), "Cancelled") {
		status = models.CourseOfferingStatusCancelled
	}

	component := strings.TrimSpace(raw.SSRComponent)
	section := strings.TrimSpace(raw.ClassSection)
	term := TermFromStrm(raw.Semester)
	year, ok := YearFromStrm(raw.Semester)
	if !ok {
		year = raw.AcademicYear
	}

	var offering models.CourseOffering
	err := db.Where("strm = ? AND class_nbr = ?", raw.Semester, raw.ClassNumber).First(&offering).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	if gorm.IsRecordNotFoundError(err) {
		offering = models.CourseOffering{
			CourseCanonicalID: courseID,
			Strm:              raw.Semester,
			Year:              year,
			Term:              term,
			Section:           section,
			ClassNbr:          raw.ClassNumber,
			Title:             title,
			Component:         component,
			Status:            status,
		}
		setOfferingCatalogFields(&offering, courseInSemesterID, raw)
		if err := db.Create(&offering).Error; err != nil {
			return nil, err
		}
		return &offering, nil
	}

	setOfferingCatalogFields(&offering, courseInSemesterID, raw)
	offering.CourseCanonicalID = courseID
	offering.Year = year
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

func setOfferingCatalogFields(o *models.CourseOffering, courseInSemesterID uint, raw *RawCourse) {
	o.CourseInSemesterID = &courseInSemesterID
	o.Source = models.CatalogSourceCatalog
	o.Subject = strings.TrimSpace(raw.Subject)
	o.CatalogNumber = raw.CatalogNumber
	o.CatalogLetter = strings.TrimSpace(raw.CourseLetter)
	o.Attributes = strings.TrimSpace(raw.AttributesSearch)
	o.Div1 = strings.Contains(o.Attributes, "DIV_D1")
	o.Div2 = strings.Contains(o.Attributes, "DIV_D2")
	o.Div3 = strings.Contains(o.Attributes, "DIV_D3")
	o.DPE = strings.Contains(o.Attributes, "DPE_DPE")
	o.QFR = strings.Contains(o.Attributes, "QFR_QFR")
	o.WAC = strings.Contains(o.Attributes, "WAC_WAC")
	o.Consent = strings.TrimSpace(raw.Consent)
	o.EnrollmentLimit = parseOptionalInt(raw.EnrollmentLimit)
	o.ExpectedEnrollment = parseOptionalInt(raw.ExpectedEnrollment)
}

func parseOptionalInt(v string) *int {
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil {
		return nil
	}
	return &n
}

func upsertCourseInSemester(db *gorm.DB, courseID uint, raw *RawCourse, title, description string) (*models.CourseInSemester, error) {
	year, ok := YearFromStrm(raw.Semester)
	if !ok {
		year = raw.AcademicYear
	}
	cancelled := strings.EqualFold(strings.TrimSpace(raw.Facility1), "Cancelled")

	var cis models.CourseInSemester
	err := db.Where("course_canonical_id = ? AND strm = ?", courseID, raw.Semester).First(&cis).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}
	isNew := gorm.IsRecordNotFoundError(err)
	// A cancelled section's details don't replace an offered one's
	if !isNew && cancelled && cis.Source == models.CatalogSourceCatalog {
		return &cis, nil
	}

	attrs := strings.TrimSpace(raw.AttributesSearch)
	cis.CourseCanonicalID = courseID
	cis.Strm = raw.Semester
	cis.Year = year
	cis.Term = TermFromStrm(raw.Semester)
	cis.Source = models.CatalogSourceCatalog
	if isNew {
		cis.Status = models.CourseOfferingStatusOffered
		if cancelled {
			cis.Status = models.CourseOfferingStatusCancelled
		}
	}
	cis.Title = title
	cis.Description = description
	cis.ShortDescription = strings.TrimSpace(raw.ShortDescription)
	cis.Components = strings.TrimSpace(raw.ComponentList)
	cis.ClassFormat = strings.TrimSpace(raw.ClassFormat)
	cis.GradingBasis = strings.TrimSpace(raw.GradingBasis)
	cis.Units = strings.TrimSpace(raw.UnitsMinimum)
	cis.Campus = strings.TrimSpace(raw.Campus)
	cis.Attributes = attrs
	cis.Div1 = strings.Contains(attrs, "DIV_D1")
	cis.Div2 = strings.Contains(attrs, "DIV_D2")
	cis.Div3 = strings.Contains(attrs, "DIV_D3")
	cis.DPE = strings.Contains(attrs, "DPE_DPE")
	cis.QFR = strings.Contains(attrs, "QFR_QFR")
	cis.WAC = strings.Contains(attrs, "WAC_WAC")
	cis.Prereqs = strings.TrimSpace(raw.PreReqs)
	cis.EnrollmentPreferences = strings.TrimSpace(raw.EnrollmentPreference)
	cis.RequirementsEvaluation = strings.TrimSpace(raw.Evaluation)
	cis.DepartmentNotes = strings.TrimSpace(raw.DepartmentNotes)
	cis.WACNotes = strings.TrimSpace(raw.DistributionNote1)
	cis.DPENotes = strings.TrimSpace(raw.DistributionNote2)
	cis.QFRNotes = strings.TrimSpace(raw.DistributionNote3)
	cis.ExtraInfo = strings.TrimSpace(strings.TrimSpace(raw.ExtraInfo) + "\n" + strings.TrimSpace(raw.ExtraInfo2))
	cis.MaterialFee = strings.TrimSpace(raw.MaterialFee)

	if isNew {
		err = db.Create(&cis).Error
	} else {
		err = db.Save(&cis).Error
	}
	if err != nil {
		return nil, err
	}
	return &cis, nil
}

// Only for terms before the feed starts, which never get real rows
func recoverLastOffered(db *gorm.DB, courseID uint, raw *RawCourse, title string, matcher *InstructorMatcher) error {
	year, ok := YearFromStrm(raw.Semester)
	if !ok || year >= EarliestCatalogYear {
		return nil
	}

	var existing models.CourseOffering
	err := db.Where("strm = ? AND class_nbr = ?", raw.Semester, raw.ClassNumber).First(&existing).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return err
	}

	var cis models.CourseInSemester
	err = db.Where("course_canonical_id = ? AND strm = ?", courseID, raw.Semester).First(&cis).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return err
	}
	if gorm.IsRecordNotFoundError(err) {
		cis = models.CourseInSemester{
			CourseCanonicalID: courseID,
			Strm:              raw.Semester,
			Year:              year,
			Term:              TermFromStrm(raw.Semester),
			Status:            models.CourseOfferingStatusOffered,
			Source:            models.CatalogSourceLastOffered,
			Title:             title,
		}
		if err := db.Create(&cis).Error; err != nil {
			return err
		}
	} else if cis.Source != models.CatalogSourceLastOffered {
		cis.Source = models.CatalogSourceLastOffered
		if err := db.Save(&cis).Error; err != nil {
			return err
		}
	}

	offering := existing
	offering.CourseCanonicalID = courseID
	offering.CourseInSemesterID = &cis.ID
	offering.Source = models.CatalogSourceLastOffered
	offering.Strm = raw.Semester
	offering.Year = year
	offering.Term = TermFromStrm(raw.Semester)
	offering.ClassNbr = raw.ClassNumber
	offering.Title = title
	offering.Component = strings.TrimSpace(raw.SSRComponent)
	offering.Status = models.CourseOfferingStatusOffered
	// The section code and meetings on these rows don't describe the old section
	offering.Section = ""
	if offering.ID == 0 {
		err = db.Create(&offering).Error
	} else {
		err = db.Save(&offering).Error
	}
	if err != nil {
		return err
	}
	return replaceInstructors(db, offering.ID, raw, matcher)
}

// Each cross-listed row carries only its own code's attributes (Div I under one, Div II under another)
func unionCourseInSemesterAttributes(db *gorm.DB, strms map[int]struct{}) error {
	list := make([]int, 0, len(strms))
	for strm := range strms {
		list = append(list, strm)
	}
	if len(list) == 0 {
		return nil
	}
	var offerings []models.CourseOffering
	err := db.Select("course_in_semester_id, attributes").
		Where("strm IN (?) AND course_in_semester_id IS NOT NULL AND source = ? AND deleted_at IS NULL", list, models.CatalogSourceCatalog).
		Find(&offerings).Error
	if err != nil {
		return err
	}
	union := map[uint]map[string]bool{}
	order := map[uint][]string{} // first-seen order, so a single listing's string is unchanged
	for _, o := range offerings {
		id := *o.CourseInSemesterID
		set := union[id]
		if set == nil {
			set = map[string]bool{}
			union[id] = set
		}
		for _, a := range strings.Split(o.Attributes, ",") {
			if a = strings.TrimSpace(a); a != "" && !set[a] {
				set[a] = true
				order[id] = append(order[id], a)
			}
		}
	}
	for id, set := range union {
		codes := order[id]
		err := db.Model(&models.CourseInSemester{}).Where("id = ?", id).Updates(map[string]interface{}{
			"attributes": strings.Join(codes, ","),
			"div1":       set["DIV_D1"], "div2": set["DIV_D2"], "div3": set["DIV_D3"],
			"dpe": set["DPE_DPE"], "qfr": set["QFR_QFR"], "wac": set["WAC_WAC"],
		}).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func refreshCourseInSemesterStatus(db *gorm.DB, strms map[int]struct{}) error {
	list := make([]int, 0, len(strms))
	for strm := range strms {
		list = append(list, strm)
	}
	if len(list) == 0 {
		return nil
	}
	return db.Exec(`
UPDATE course_in_semesters cis
SET status = CASE WHEN EXISTS (
	SELECT 1 FROM course_offerings o
	WHERE o.course_in_semester_id = cis.id AND o.deleted_at IS NULL AND o.status = ?
) THEN ? ELSE ? END
WHERE cis.strm IN (?) AND cis.deleted_at IS NULL`,
		models.CourseOfferingStatusOffered, models.CourseOfferingStatusOffered, models.CourseOfferingStatusCancelled, list).Error
}

func replaceMeetings(db *gorm.DB, offeringID uint, raw *RawCourse) error {
	if err := db.Unscoped().Where("course_offering_id = ?", offeringID).Delete(&models.CourseOfferingMeeting{}).Error; err != nil {
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
		meeting := models.CourseOfferingMeeting{
			CourseOfferingID: offeringID,
			Days:             days,
			Start:            start,
			End:              end,
			Facility:         facility,
		}
		if err := db.Create(&meeting).Error; err != nil {
			return err
		}
	}
	return nil
}

func replaceInstructors(db *gorm.DB, offeringID uint, raw *RawCourse, matcher *InstructorMatcher) error {
	if err := db.Unscoped().Where("course_offering_id = ?", offeringID).Delete(&models.CourseOfferingInstructor{}).Error; err != nil {
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

		userID, _ := matcher.Match(unixID, fn, mn, ln)

		instr := models.CourseOfferingInstructor{
			CourseOfferingID: offeringID,
			UnixID:           unixID,
			UserID:           userID,
			NameAsListed:     name,
		}
		if err := db.Create(&instr).Error; err != nil {
			return err
		}
	}
	return nil
}
