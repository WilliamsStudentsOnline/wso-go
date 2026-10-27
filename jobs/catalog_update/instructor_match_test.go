package catalog_update_test

import (
	"testing"

	. "github.com/WilliamsStudentsOnline/wso-go/jobs/catalog_update"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNormalizePersonName(t *testing.T) {
	assert := testify.New(t)
	assert.Equal("brittany meche", NormalizePersonName("Brittany  Meché"))
	assert.Equal("michelle m apotsos", NormalizePersonName("Michelle M. Apotsos"))
	assert.Equal("obrien", NormalizePersonName("O'Brien"))
	assert.Equal("jose garcia", NormalizePersonName("José  García"))
}

func TestInstructorMatcher(t *testing.T) {
	assert := testify.New(t)
	db := test_utils.SetupServiceTest(assert)

	atWilliams := false // matching must ignore at_williams
	profUID := models.User{
		Type: models.UserTypeProfessor, Name: "Michelle M. Apotsos", UnixID: "ma11",
		AtWilliams: &atWilliams, Visible: lib.BoolToPtr(true),
	}
	profAccent := models.User{
		Type: models.UserTypeProfessor, Name: "Brittany Meché", UnixID: "bm18",
		Visible: lib.BoolToPtr(true),
	}
	profNick := models.User{
		Type: models.UserTypeProfessor, Name: "William Lenhart", UnixID: "wjl1",
		Nickname: lib.StrToPtr("Bill"), Visible: lib.BoolToPtr(true),
	}
	profNoUID := models.User{
		Type: models.UserTypeProfessor, Name: "Jane Q. Doe", UnixID: "jqd1",
		Visible: lib.BoolToPtr(true),
	}
	studentSameName := models.User{
		Type: models.UserTypeStudent, Name: "Jane Q. Doe", UnixID: "jqd2",
		Visible: lib.BoolToPtr(true),
	}
	assert.NoError(db.Create(&profUID).Error)
	assert.NoError(db.Create(&profAccent).Error)
	assert.NoError(db.Create(&profNick).Error)
	assert.NoError(db.Create(&profNoUID).Error)
	assert.NoError(db.Create(&studentSameName).Error)

	m, err := NewInstructorMatcher(db, zap.NewNop().Sugar())
	assert.NoError(err)

	t.Run("unix id wins even when not at williams", func(t *testing.T) {
		id, via := m.Match("ma11", "Michelle", "M.", "Apotsos")
		assert.Equal("unix_id", via)
		assert.NotNil(id)
		assert.Equal(profUID.ID, *id)
	})

	t.Run("accent folded name match when uid missing", func(t *testing.T) {
		id, via := m.Match("", "Brittany", "", "Meche")
		assert.Equal("name", via)
		assert.NotNil(id)
		assert.Equal(profAccent.ID, *id)
	})

	t.Run("middle name dropped", func(t *testing.T) {
		id, via := m.Match("wrong-uid", "Michelle", "X", "Apotsos")
		assert.Equal("name", via)
		assert.NotNil(id)
		assert.Equal(profUID.ID, *id)
	})

	t.Run("nickname match", func(t *testing.T) {
		id, via := m.Match("", "Bill", "", "Lenhart")
		assert.Equal("name", via)
		assert.NotNil(id)
		assert.Equal(profNick.ID, *id)
	})

	t.Run("prefers professor when student shares name", func(t *testing.T) {
		id, via := m.Match("", "Jane", "Q.", "Doe")
		assert.Equal("name", via)
		assert.NotNil(id)
		assert.Equal(profNoUID.ID, *id)
	})

	t.Run("unmatched leaves nil", func(t *testing.T) {
		id, via := m.Match("ghost", "Nobody", "", "Here")
		assert.Equal("", via)
		assert.Nil(id)
	})
}

func TestIngestUsesNameFallback(t *testing.T) {
	assert := testify.New(t)
	db := test_utils.SetupServiceTest(assert)

	prof := models.User{
		Type: models.UserTypeProfessor, Name: "Brittany Meché", UnixID: "bm18",
		Visible: lib.BoolToPtr(true),
	}
	assert.NoError(db.Create(&prof).Error)

	raw := []RawCourse{{
		AcademicYear: 2027, Offered: "Y", Semester: 1271, CourseID: "020001",
		Subject: "ENGL", CatalogNumber: 100, ClassSection: "01", ClassNumber: 5001,
		SSRComponent: "SEM", CourseTitleLong: "Poetry",
		FirstName1: "Brittany", LastName1: "Meche", // no UID; accent differs
	}}
	assert.NoError(IngestCatalog(db, raw, zap.NewNop().Sugar()))

	var instr models.OfferingInstructor
	assert.NoError(db.First(&instr).Error)
	assert.NotNil(instr.UserID)
	assert.Equal(prof.ID, *instr.UserID)
	assert.Equal("", instr.UnixID)
}

func TestRematchOfferingInstructors(t *testing.T) {
	assert := testify.New(t)
	db := test_utils.SetupServiceTest(assert)

	// Ingest with no users → unmatched
	raw := []RawCourse{{
		AcademicYear: 2027, Offered: "Y", Semester: 1271, CourseID: "020002",
		Subject: "CSCI", CatalogNumber: 134, ClassSection: "01", ClassNumber: 5002,
		SSRComponent: "LEC", CourseTitleLong: "Intro",
		FirstName1: "Bill", LastName1: "Lenhart", UnixID1: "wjl1",
	}}
	assert.NoError(IngestCatalog(db, raw, zap.NewNop().Sugar()))

	var before models.OfferingInstructor
	assert.NoError(db.First(&before).Error)
	assert.Nil(before.UserID)

	prof := models.User{
		Type: models.UserTypeProfessor, Name: "William Lenhart", UnixID: "wjl1",
		Nickname: lib.StrToPtr("Bill"), Visible: lib.BoolToPtr(true),
	}
	assert.NoError(db.Create(&prof).Error)

	stats, err := RematchOfferingInstructors(db, zap.NewNop().Sugar())
	assert.NoError(err)
	assert.Equal(1, stats.UIDHits)

	var after models.OfferingInstructor
	assert.NoError(db.First(&after).Error)
	assert.NotNil(after.UserID)
	assert.Equal(prof.ID, *after.UserID)
}
