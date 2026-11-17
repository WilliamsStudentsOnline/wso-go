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
			AcademicYear:      2025,
			Offered:           "Y",
			Semester:          1251,
			CourseID:          "020209",
			EffectiveDate:     "01-SEP-24",
			Subject:           "AFR",
			CatalogNumber:     105,
			CourseLetter:      "",
			ClassSection:      "01",
			ClassNumber:       1089,
			SSRComponent:      "LEC",
			Description:       "African Art Survey",
			CourseTitleLong:   "Materials 2025",
			DescriptionSearch: "Desc 2025",
			FirstName1:        "Michelle",
			MiddleName1:       "M.",
			LastName1:         "Apotsos",
			UnixID1:           "ma11",
			StandardMeeting1:  "MW",
			StartTime1:        "11:00",
			EndTime1:          "12:15",
			Facility1:         "Hopkins",
			Facility3:         "ShouldNotBeWrongTag",
			StandardMeeting3:  "F",
			StartTime3:        "13:00",
			EndTime3:          "14:00",
		},
		{
			AcademicYear:      2024,
			Offered:           "Y",
			Semester:          1241,
			CourseID:          "020209",
			Subject:           "AFR",
			CatalogNumber:     105,
			ClassSection:      "01",
			ClassNumber:       2001,
			SSRComponent:      "LEC",
			Description:       "African Art Survey",
			CourseTitleLong:   "Materials 2024 OLDER",
			DescriptionSearch: "Desc 2024",
			FirstName1:        "Michelle",
			LastName1:         "Apotsos",
			UnixID1:           "ma11",
			StandardMeeting1:  "TR",
			StartTime1:        "10:00",
			EndTime1:          "11:15",
			Facility1:         "Hopkins",
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
	assert.NoError(db.Model(&models.CourseOffering{}).Count(&offeringCount).Error)
	assert.NoError(db.Model(&models.CourseOfferingMeeting{}).Count(&meetingCount).Error)
	assert.NoError(db.Model(&models.CourseOfferingInstructor{}).Count(&instructorCount).Error)

	assert.Equal(2, courseCount)
	assert.Equal(2, listingCount)
	assert.Equal(3, offeringCount)
	var cisCount int
	assert.NoError(db.Model(&models.CourseInSemester{}).Count(&cisCount).Error)
	assert.Equal(3, cisCount)
	assert.Equal(3, meetingCount)
	assert.Equal(3, instructorCount)

	var course models.CourseCanonical
	assert.NoError(db.Where("crse_id = ?", "020209").First(&course).Error)
	assert.Equal("Materials 2025", course.Title)
	assert.Equal("Desc 2025", course.Description)

	var listing models.CourseListing
	assert.NoError(db.Where("course_canonical_id = ?", course.ID).First(&listing).Error)
	assert.Equal(2024, listing.FirstValidYear)
	assert.Equal(2025, listing.LastValidYear)

	var cancelled models.CourseOffering
	assert.NoError(db.Where("class_nbr = ?", 3001).First(&cancelled).Error)
	assert.Equal(models.CourseOfferingStatusCancelled, cancelled.Status)
	assert.Equal("Fall", cancelled.Term)
	var cancelledCIS models.CourseInSemester
	assert.NoError(db.Where("id = ?", *cancelled.CourseInSemesterID).First(&cancelledCIS).Error)
	assert.Equal(models.CourseOfferingStatusCancelled, cancelledCIS.Status)

	var matched models.CourseOfferingInstructor
	assert.NoError(db.Where("unix_id = ?", "ma11").First(&matched).Error)
	assert.NotNil(matched.UserID)
	assert.Equal(prof.ID, *matched.UserID)

	var unmatched models.CourseOfferingInstructor
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
	assert.NoError(db.Model(&models.CourseOffering{}).Count(&offeringCount).Error)
	assert.Equal(2, offeringCount)
}

// Not-offered rows reuse the strm and class number of the last real section (seen with AFR 104)
func TestIngestNotOfferedKeepsRealSection(t *testing.T) {
	assert := testify.New(t)
	db := test_utils.SetupServiceTest(assert)

	offered := []RawCourse{{
		AcademicYear: 2023, Offered: "Y", Semester: 1233, CourseID: "010001",
		Subject: "AFR", CatalogNumber: 104, ClassSection: "01", ClassNumber: 3377,
		SSRComponent: "LEC", CourseTitleLong: "Real Title", FirstName1: "Real", LastName1: "Prof",
		StandardMeeting1: "MWF", StartTime1: "09:00", EndTime1: "09:50",
	}}
	notOffered := []RawCourse{{
		AcademicYear: 2026, Offered: "N", Semester: 1233, CourseID: "010001",
		Subject: "AFR", CatalogNumber: 104, ClassSection: "NA", ClassNumber: 3377,
		SSRComponent: "LEC", CourseTitleLong: "Real Title", FirstName1: "Real", LastName1: "Prof",
	}}
	assert.NoError(IngestCatalog(db, offered, zap.NewNop().Sugar()))
	assert.NoError(IngestCatalog(db, notOffered, zap.NewNop().Sugar()))

	var offerings []models.CourseOffering
	assert.NoError(db.Preload("Meetings").Preload("Instructors").Find(&offerings).Error)
	if assert.Len(offerings, 1) {
		o := offerings[0]
		assert.Equal(models.CourseOfferingStatusOffered, o.Status)
		assert.Equal(2023, o.Year)
		assert.Equal("Spring", o.Term)
		assert.Equal("01", o.Section)
		assert.Len(o.Meetings, 1)
		assert.Len(o.Instructors, 1)
	}

	var listing models.CourseListing
	assert.NoError(db.Where("subject = ? AND number = ?", "AFR", 104).First(&listing).Error)
	assert.Equal(2023, listing.FirstValidYear)
	assert.Equal(2026, listing.LastValidYear)
}

func TestYearFromStrm(t *testing.T) {
	assert := testify.New(t)
	y, ok := YearFromStrm(1271)
	assert.True(ok)
	assert.Equal(2027, y)
	y, ok = YearFromStrm(1201)
	assert.True(ok)
	assert.Equal(2020, y)
	_, ok = YearFromStrm(0)
	assert.False(ok)
}

func TestIngestCourseInSemesterFields(t *testing.T) {
	assert := testify.New(t)
	db := test_utils.SetupServiceTest(assert)

	rows := []RawCourse{
		{
			AcademicYear: 2027, Offered: "Y", Semester: 1271, CourseID: "030001",
			Subject: "CHEM", CatalogNumber: 101, ClassSection: "01", ClassNumber: 4001,
			SSRComponent: "LEC", CourseTitleLong: "Concepts of Chemistry", DescriptionSearch: "Atoms.",
			AttributesSearch: "DIV_D3,QFR_QFR,AFR_AFRCORE", GradingBasis: "OPT", ComponentList: "lecture/laboratory",
			PreReqs: "None", EnrollmentLimit: "60", ExpectedEnrollment: "NA", Consent: "N", UnitsMinimum: "1",
			DistributionNote3: "Weekly problem sets.",
		},
		{
			AcademicYear: 2027, Offered: "Y", Semester: 1271, CourseID: "030001",
			Subject: "CHEM", CatalogNumber: 101, ClassSection: "L1", ClassNumber: 4002,
			SSRComponent: "LAB", CourseTitleLong: "Concepts of Chemistry", Facility1: "Cancelled",
			AttributesSearch: "", EnrollmentLimit: "16",
		},
	}
	assert.NoError(IngestCatalog(db, rows, zap.NewNop().Sugar()))
	assert.NoError(IngestCatalog(db, rows, zap.NewNop().Sugar()))

	var all []models.CourseInSemester
	assert.NoError(db.Preload("Offerings").Find(&all).Error)
	if !assert.Len(all, 1) {
		return
	}
	cis := all[0]
	assert.Equal(models.CourseOfferingStatusOffered, cis.Status)
	assert.Equal(models.CatalogSourceCatalog, cis.Source)
	assert.Equal(2027, cis.Year)
	assert.Equal("Fall", cis.Term)
	assert.True(cis.Div3)
	assert.True(cis.QFR)
	assert.False(cis.Div1)
	assert.Equal("DIV_D3,QFR_QFR,AFR_AFRCORE", cis.Attributes)
	assert.Equal("OPT", cis.GradingBasis)
	assert.Equal("None", cis.Prereqs)
	assert.Equal("Weekly problem sets.", cis.QFRNotes)
	assert.Len(cis.Offerings, 2)

	var lecture models.CourseOffering
	assert.NoError(db.Where("class_nbr = ?", 4001).First(&lecture).Error)
	if assert.NotNil(lecture.EnrollmentLimit) {
		assert.Equal(60, *lecture.EnrollmentLimit)
	}
	assert.Nil(lecture.ExpectedEnrollment)
	assert.Equal("N", lecture.Consent)
}

// For terms before the feed, the not-offered pointer is the only record we have
func TestIngestRecoversSectionsBeforeFeed(t *testing.T) {
	assert := testify.New(t)
	db := test_utils.SetupServiceTest(assert)

	notOfferedOld := RawCourse{
		AcademicYear: 2023, Offered: "N", Semester: 1201, CourseID: "040001",
		Subject: "AFR", CatalogNumber: 105, ClassSection: "NA", ClassNumber: 1089,
		SSRComponent: "SEM", CourseTitleLong: "Arts of Africa", FirstName1: "Michelle", LastName1: "Apotsos",
		StandardMeeting1: "MW", StartTime1: "11:00", EndTime1: "12:15",
	}
	notOfferedInFeed := RawCourse{
		AcademicYear: 2025, Offered: "N", Semester: 1241, CourseID: "040002",
		Subject: "AFR", CatalogNumber: 106, ClassNumber: 2002, CourseTitleLong: "In-feed term",
		FirstName1: "Some", LastName1: "Prof",
	}
	assert.NoError(IngestCatalog(db, []RawCourse{notOfferedOld, notOfferedInFeed}, zap.NewNop().Sugar()))
	notOfferedOld.AcademicYear = 2024
	assert.NoError(IngestCatalog(db, []RawCourse{notOfferedOld}, zap.NewNop().Sugar()))

	var offerings []models.CourseOffering
	assert.NoError(db.Preload("Meetings").Preload("Instructors").Find(&offerings).Error)
	if assert.Len(offerings, 1) {
		o := offerings[0]
		assert.Equal(models.CatalogSourceLastOffered, o.Source)
		assert.Equal(2020, o.Year)
		assert.Equal("Fall", o.Term)
		assert.Equal(models.CourseOfferingStatusOffered, o.Status)
		assert.Equal("", o.Section)
		assert.Len(o.Meetings, 0)
		if assert.Len(o.Instructors, 1) {
			assert.Equal("Michelle Apotsos", o.Instructors[0].NameAsListed)
		}
	}
	var cis []models.CourseInSemester
	assert.NoError(db.Find(&cis).Error)
	if assert.Len(cis, 1) {
		assert.Equal(models.CatalogSourceLastOffered, cis[0].Source)
		assert.Equal(1201, cis[0].Strm)
	}

	assert.NoError(db.Model(&models.CourseOffering{}).Where("class_nbr = ?", 1089).
		Updates(map[string]interface{}{"source": models.CatalogSourceCatalog, "course_in_semester_id": nil}).Error)
	assert.NoError(IngestCatalog(db, []RawCourse{notOfferedOld}, zap.NewNop().Sugar()))
	var o models.CourseOffering
	assert.NoError(db.Where("class_nbr = ?", 1089).First(&o).Error)
	assert.Equal(models.CatalogSourceLastOffered, o.Source)
	assert.NotNil(o.CourseInSemesterID)
}

func TestIngestNotOfferedInFeedNeverOverwrites(t *testing.T) {
	assert := testify.New(t)
	db := test_utils.SetupServiceTest(assert)
	real := RawCourse{
		AcademicYear: 2024, Offered: "Y", Semester: 1241, CourseID: "050001",
		Subject: "AFR", CatalogNumber: 110, ClassSection: "01", ClassNumber: 3003,
		CourseTitleLong: "Real", FirstName1: "Real", LastName1: "Prof",
	}
	pointer := real
	pointer.AcademicYear, pointer.Offered, pointer.ClassSection = 2025, "N", "NA"
	assert.NoError(IngestCatalog(db, []RawCourse{real}, zap.NewNop().Sugar()))
	assert.NoError(IngestCatalog(db, []RawCourse{pointer}, zap.NewNop().Sugar()))
	var o models.CourseOffering
	assert.NoError(db.Where("class_nbr = ?", 3003).First(&o).Error)
	assert.Equal(models.CatalogSourceCatalog, o.Source)
	assert.Equal("01", o.Section)
}
