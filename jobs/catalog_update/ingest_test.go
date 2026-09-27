package catalog_update_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/WilliamsStudentsOnline/wso-go/jobs/catalog_update"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func sampleRawCourses() []RawCourse {
	return []RawCourse{
		{
			AcademicYear:     2025,
			Offered:          "Y",
			Semester:         1251,
			CourseID:         "020209",
			EffectiveDate:    "01-SEP-24",
			Subject:          "AFR",
			CatalogNumber:    105,
			CourseLetter:     "",
			ClassSection:     "01",
			ClassNumber:      1089,
			SSRComponent:     "LEC",
			Description:      "African Art Survey",
			CourseTitleLong:  "Materials 2025",
			DescriptionSearch: "Desc 2025",
			FirstName1:       "Michelle",
			MiddleName1:      "M.",
			LastName1:        "Apotsos",
			UnixID1:          "ma11",
			StandardMeeting1: "MW",
			StartTime1:       "11:00",
			EndTime1:         "12:15",
			Facility1:        "Hopkins",
			Facility3:        "ShouldNotBeWrongTag",
			StandardMeeting3: "F",
			StartTime3:       "13:00",
			EndTime3:         "14:00",
		},
		{
			AcademicYear:     2024,
			Offered:          "Y",
			Semester:         1241,
			CourseID:         "020209",
			Subject:          "AFR",
			CatalogNumber:    105,
			ClassSection:     "01",
			ClassNumber:      2001,
			SSRComponent:     "LEC",
			Description:      "African Art Survey",
			CourseTitleLong:  "Materials 2024 OLDER",
			DescriptionSearch: "Desc 2024",
			FirstName1:       "Michelle",
			LastName1:        "Apotsos",
			UnixID1:          "ma11",
			StandardMeeting1: "TR",
			StartTime1:       "10:00",
			EndTime1:         "11:15",
			Facility1:        "Hopkins",
		},
		{
			AcademicYear:    2025,
			Offered:         "Y",
			Semester:        1251,
			CourseID:        "099999",
			Subject:         "CSCI",
			CatalogNumber:   136,
			ClassSection:    "01",
			ClassNumber:     3001,
			SSRComponent:    "LEC",
			CourseTitleLong: "Data Structures",
			Facility1:       "Cancelled",
			FirstName1:      "Bill",
			LastName1:       "Lenhart",
			UnixID1:         "missinguid",
		},
	}
}

func TestIngestCatalogIdempotent(t *testing.T) {
	assert := testify.New(t)
	db := test_utils.SetupServiceTest(assert)

	atWilliams := true
	prof := models.User{
		Type:       models.UserTypeProfessor,
		Name:       "Michelle M. Apotsos",
		UnixID:     "ma11",
		AtWilliams: &atWilliams,
		Visible:    lib.BoolToPtr(true),
	}
	assert.NoError(db.Create(&prof).Error)

	raw := sampleRawCourses()
	assert.NoError(IngestCatalog(db, raw, zap.NewNop().Sugar()))
	assert.NoError(IngestCatalog(db, raw, zap.NewNop().Sugar()))

	var courseCount, listingCount, offeringCount, meetingCount, instructorCount int
	assert.NoError(db.Model(&models.CourseCanonical{}).Count(&courseCount).Error)
	assert.NoError(db.Model(&models.CourseListing{}).Count(&listingCount).Error)
	assert.NoError(db.Model(&models.Offering{}).Count(&offeringCount).Error)
	assert.NoError(db.Model(&models.OfferingMeeting{}).Count(&meetingCount).Error)
	assert.NoError(db.Model(&models.OfferingInstructor{}).Count(&instructorCount).Error)

	assert.Equal(2, courseCount)
	assert.Equal(2, listingCount)
	assert.Equal(3, offeringCount)
	assert.Equal(3, meetingCount) // two for AFR 2025 (MW+F), one for AFR 2024
	assert.Equal(3, instructorCount)

	var course models.CourseCanonical
	assert.NoError(db.Where("crse_id = ?", "020209").First(&course).Error)
	assert.Equal("Materials 2025", course.Title)
	assert.Equal("Desc 2025", course.Description)

	var listing models.CourseListing
	assert.NoError(db.Where("course_canonical_id = ?", course.ID).First(&listing).Error)
	assert.Equal(2024, listing.FirstValidYear)
	assert.Equal(2025, listing.LastValidYear)

	var cancelled models.Offering
	assert.NoError(db.Where("class_nbr = ?", 3001).First(&cancelled).Error)
	assert.Equal(models.OfferingStatusCancelled, cancelled.Status)
	assert.Equal("Fall", cancelled.Term)

	var matched models.OfferingInstructor
	assert.NoError(db.Where("unix_id = ?", "ma11").First(&matched).Error)
	assert.NotNil(matched.UserID)
	assert.Equal(prof.ID, *matched.UserID)

	var unmatched models.OfferingInstructor
	assert.NoError(db.Where("unix_id = ?", "missinguid").First(&unmatched).Error)
	assert.Nil(unmatched.UserID)
}

func TestArchiveRawCatalog(t *testing.T) {
	assert := testify.New(t)
	dir := t.TempDir()
	data := []byte(`[{"CRSE_ID":"1"}]`)
	assert.NoError(ArchiveRawCatalog(dir, 2027, data))
	b, err := os.ReadFile(filepath.Join(dir, "raw-2027.json"))
	assert.NoError(err)
	assert.Equal(data, b)

	assert.NoError(ArchiveRawCatalog(dir, 0, data))
	_, err = os.Stat(filepath.Join(dir, "raw-current.json"))
	assert.NoError(err)
}

func TestReadCatalogYearFile(t *testing.T) {
	assert := testify.New(t)
	dir := t.TempDir()
	payload := []byte(`[{"CRSE_ID":"020209","OFFERED":"Y","STRM":"1251","CLASS_NBR":"1","SUBJECT":"AFR","CATALOG_NBR":"105","WMS_ACAD_YEAR":"2025"}]`)
	assert.NoError(os.WriteFile(filepath.Join(dir, "year_2025.json"), payload, 0644))

	data, err := ReadCatalogYearFile(dir, 2025)
	assert.NoError(err)
	assert.Equal(payload, data)

	assert.NoError(os.WriteFile(filepath.Join(dir, "raw-2026.json"), payload, 0644))
	data, err = ReadCatalogYearFile(dir, 2026)
	assert.NoError(err)
	assert.Equal(payload, data)

	_, err = ReadCatalogYearFile(dir, 2023)
	assert.Error(err)
}

func TestIngestHistoricalYearRange(t *testing.T) {
	assert := testify.New(t)
	db := test_utils.SetupServiceTest(assert)

	// Two years, same listing: range should expand; newest title wins.
	y2023 := []RawCourse{{
		AcademicYear: 2023, Offered: "Y", Semester: 1231, CourseID: "020209",
		Subject: "AFR", CatalogNumber: 105, ClassSection: "01", ClassNumber: 1001,
		SSRComponent: "LEC", CourseTitleLong: "Title 2023", DescriptionSearch: "D23",
		FirstName1: "A", LastName1: "Prof",
	}}
	y2024 := []RawCourse{{
		AcademicYear: 2024, Offered: "Y", Semester: 1241, CourseID: "020209",
		Subject: "AFR", CatalogNumber: 105, ClassSection: "01", ClassNumber: 1002,
		SSRComponent: "LEC", CourseTitleLong: "Title 2024", DescriptionSearch: "D24",
		FirstName1: "A", LastName1: "Prof",
	}}

	assert.NoError(IngestCatalog(db, y2023, zap.NewNop().Sugar()))
	assert.NoError(IngestCatalog(db, y2024, zap.NewNop().Sugar()))

	var course models.CourseCanonical
	assert.NoError(db.Where("crse_id = ?", "020209").First(&course).Error)
	assert.Equal("Title 2024", course.Title)

	var listing models.CourseListing
	assert.NoError(db.Where("course_canonical_id = ?", course.ID).First(&listing).Error)
	assert.Equal(2023, listing.FirstValidYear)
	assert.Equal(2024, listing.LastValidYear)

	var offeringCount int
	assert.NoError(db.Model(&models.Offering{}).Count(&offeringCount).Error)
	assert.Equal(2, offeringCount)
}
