package factrak_backfill_test

import (
	"testing"

	. "github.com/WilliamsStudentsOnline/wso-go/jobs/factrak_backfill"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestParseCourseCode(t *testing.T) {
	assert := testify.New(t)
	n, l, err := ParseCourseCode("136")
	assert.NoError(err)
	assert.Equal(136, n)
	assert.Equal("", l)

	n, l, err = ParseCourseCode("301B")
	assert.NoError(err)
	assert.Equal(301, n)
	assert.Equal("B", l)

	_, _, err = ParseCourseCode("B301")
	assert.Error(err)
}

func TestCatalogYearTerm(t *testing.T) {
	assert := testify.New(t)
	y, term, ok := CatalogYearTerm("fall", 2026)
	assert.True(ok)
	assert.Equal(2027, y)
	assert.Equal("Fall", term)

	y, term, ok = CatalogYearTerm("winter-study", 2027)
	assert.True(ok)
	assert.Equal(2027, y)
	assert.Equal("Winter", term)

	y, term, ok = CatalogYearTerm("spring", 2027)
	assert.True(ok)
	assert.Equal(2027, y)
	assert.Equal("Spring", term)
}

func TestBackfillExactAndAliasAndDisambiguate(t *testing.T) {
	assert := testify.New(t)
	db := test_utils.SetupServiceTest(assert)
	log := zap.NewNop().Sugar()

	// Alias WGST → WGSS (may already be seeded by migration)
	assert.NoError(db.Where(models.CourseSubjectAlias{FactrakAbbrev: "WGST"}).
		Attrs(models.CourseSubjectAlias{CatalogSubject: "WGSS"}).
		FirstOrCreate(&models.CourseSubjectAlias{}).Error)

	dept := models.Department{Name: "Test Dept"}
	assert.NoError(db.Create(&dept).Error)
	aosCSCI := models.AreaOfStudy{Name: "Computer Science", Abbreviation: "CSCI", DepartmentID: &dept.ID}
	aosWGST := models.AreaOfStudy{Name: "Women's Studies", Abbreviation: "WGST", DepartmentID: &dept.ID}
	assert.NoError(db.Create(&aosCSCI).Error)
	assert.NoError(db.Create(&aosWGST).Error)

	prof := models.User{Type: models.UserTypeProfessor, Name: "Ada Lovelace", UnixID: "al1", Visible: lib.BoolToPtr(true)}
	prof2 := models.User{Type: models.UserTypeProfessor, Name: "Other Prof", UnixID: "op1", Visible: lib.BoolToPtr(true)}
	student := models.User{Type: models.UserTypeStudent, Name: "Stu Dent", UnixID: "sd1", Visible: lib.BoolToPtr(true)}
	assert.NoError(db.Create(&prof).Error)
	assert.NoError(db.Create(&prof2).Error)
	assert.NoError(db.Create(&student).Error)

	// Exact listing: CSCI 136
	cc1 := models.CourseCanonical{CrseID: "100001", Title: "Data Structures"}
	assert.NoError(db.Create(&cc1).Error)
	assert.NoError(db.Create(&models.CourseListing{
		CourseCanonicalID: cc1.ID, Subject: "CSCI", Number: 136, Letter: "",
		FirstValidYear: 2023, LastValidYear: 2027,
	}).Error)
	off1 := models.Offering{
		CourseCanonicalID: cc1.ID, Strm: 1271, Year: 2027, Term: "Fall",
		Section: "01", ClassNbr: 1001, Title: "Data Structures", Component: "LEC",
		Status: models.OfferingStatusOffered,
	}
	assert.NoError(db.Create(&off1).Error)
	assert.NoError(db.Create(&models.OfferingInstructor{
		OfferingID: off1.ID, UnixID: "al1", UserID: &prof.ID, NameAsListed: "Ada Lovelace",
	}).Error)

	// Alias: WGSS 101 in catalog, Factrak WGST 101
	cc2 := models.CourseCanonical{CrseID: "100002", Title: "Intro WGSS"}
	assert.NoError(db.Create(&cc2).Error)
	assert.NoError(db.Create(&models.CourseListing{
		CourseCanonicalID: cc2.ID, Subject: "WGSS", Number: 101, Letter: "",
		FirstValidYear: 2023, LastValidYear: 2027,
	}).Error)

	// Multi-CRSE same code ENGL 304 — two canonicals; only cc3A taught by prof
	cc3A := models.CourseCanonical{CrseID: "100003", Title: "ENGL 304 A"}
	cc3B := models.CourseCanonical{CrseID: "100004", Title: "ENGL 304 B"}
	assert.NoError(db.Create(&cc3A).Error)
	assert.NoError(db.Create(&cc3B).Error)
	assert.NoError(db.Create(&models.CourseListing{
		CourseCanonicalID: cc3A.ID, Subject: "ENGL", Number: 304, Letter: "",
		FirstValidYear: 2023, LastValidYear: 2027,
	}).Error)
	assert.NoError(db.Create(&models.CourseListing{
		CourseCanonicalID: cc3B.ID, Subject: "ENGL", Number: 304, Letter: "",
		FirstValidYear: 2023, LastValidYear: 2027,
	}).Error)
	off3 := models.Offering{
		CourseCanonicalID: cc3A.ID, Strm: 1273, Year: 2027, Term: "Spring",
		Section: "01", ClassNbr: 1003, Title: "ENGL 304 A", Component: "SEM",
		Status: models.OfferingStatusOffered,
	}
	assert.NoError(db.Create(&off3).Error)
	assert.NoError(db.Create(&models.OfferingInstructor{
		OfferingID: off3.ID, UnixID: "al1", UserID: &prof.ID, NameAsListed: "Ada Lovelace",
	}).Error)

	aosENGL := models.AreaOfStudy{Name: "English", Abbreviation: "ENGL", DepartmentID: &dept.ID}
	assert.NoError(db.Create(&aosENGL).Error)

	fc1 := models.Course{Number: "136", AreaOfStudyID: &aosCSCI.ID}
	fc2 := models.Course{Number: "101", AreaOfStudyID: &aosWGST.ID}
	fc3 := models.Course{Number: "304", AreaOfStudyID: &aosENGL.ID}
	assert.NoError(db.Create(&fc1).Error)
	assert.NoError(db.Create(&fc2).Error)
	assert.NoError(db.Create(&fc3).Error)

	seasonFall := models.FactrakSurveySemesterSeasonFall
	seasonSpring := models.FactrakSurveySemesterSeasonSpring
	year2026 := 2026 // Fall 2026 → catalog 2027 Fall
	year2027 := 2027

	s1 := models.FactrakSurvey{
		UserID: student.ID, ProfessorID: prof.ID, CourseID: fc1.ID,
		Comment: "exact", SemesterSeason: &seasonFall, SemesterYear: &year2026,
	}
	s2 := models.FactrakSurvey{
		UserID: student.ID, ProfessorID: prof.ID, CourseID: fc2.ID,
		Comment: "alias",
	}
	s3 := models.FactrakSurvey{
		UserID: student.ID, ProfessorID: prof.ID, CourseID: fc3.ID,
		Comment: "disambig", SemesterSeason: &seasonSpring, SemesterYear: &year2027,
	}
	s4 := models.FactrakSurvey{
		UserID: student.ID, ProfessorID: prof2.ID, CourseID: fc3.ID,
		Comment: "ambiguous unmatched", SemesterSeason: &seasonSpring, SemesterYear: &year2027,
	}
	assert.NoError(db.Create(&s1).Error)
	assert.NoError(db.Create(&s2).Error)
	assert.NoError(db.Create(&s3).Error)
	assert.NoError(db.Create(&s4).Error)

	// Dry run
	report, err := RunBackfill(db, log, Options{DryRun: true, IncludeResults: true})
	assert.NoError(err)
	t.Logf("dry-run report: exact=%d disambig=%d ambig=%d unmatched=%d",
		report.ExactListing, report.Disambiguated, report.Ambiguous, report.Unmatched)
	for _, r := range report.Results {
		t.Logf("  survey=%d status=%s subject=%s note=%q candidates=%d",
			r.SurveyID, r.Status, r.CatalogSubject, r.Note, r.CandidateCount)
	}
	assert.Equal(4, report.Total)
	assert.Equal(2, report.ExactListing) // s1 exact + s2 alias (single listing)
	assert.Equal(1, report.Disambiguated)
	assert.Equal(1, report.Unmatched)

	// Apply
	report, err = RunBackfill(db, log, Options{DryRun: false, SkipLinked: true, IncludeResults: true})
	assert.NoError(err)
	assert.Equal(3, report.Applied)

	var linked models.FactrakSurvey
	assert.NoError(db.Where("id = ?", s1.ID).First(&linked).Error)
	assert.NotNil(linked.CanonicalCourseID)
	assert.Equal(cc1.ID, *linked.CanonicalCourseID)
	assert.NotNil(linked.OfferingID)
	assert.Equal(off1.ID, *linked.OfferingID)

	linked = models.FactrakSurvey{}
	assert.NoError(db.Where("id = ?", s2.ID).First(&linked).Error)
	assert.NotNil(linked.CanonicalCourseID)
	assert.Equal(cc2.ID, *linked.CanonicalCourseID)

	linked = models.FactrakSurvey{}
	assert.NoError(db.Where("id = ?", s3.ID).First(&linked).Error)
	assert.NotNil(linked.CanonicalCourseID)
	assert.Equal(cc3A.ID, *linked.CanonicalCourseID)

	linked = models.FactrakSurvey{}
	assert.NoError(db.Where("id = ?", s4.ID).First(&linked).Error)
	assert.Nil(linked.CanonicalCourseID)
}
