package factrak

import (
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestSyncReviewedUserCourses(t *testing.T) {
	assert := testify.New(t)
	db := test_utils.SetupServiceTest(assert)

	dept := models.Department{Name: "Test Dept"}
	assert.NoError(db.Create(&dept).Error)
	area := models.AreaOfStudy{Name: "CSCI", Abbreviation: "CSCI", DepartmentID: &dept.ID}
	assert.NoError(db.Create(&area).Error)
	fc := models.Course{Number: "136", AreaOfStudyID: &area.ID}
	assert.NoError(db.Create(&fc).Error)
	prof := models.User{Type: models.UserTypeProfessor, Name: "Prof", UnixID: "prof", Visible: lib.BoolToPtr(true)}
	assert.NoError(db.Create(&prof).Error)
	alice := models.User{Type: models.UserTypeStudent, Name: "Alice", UnixID: "alice", Visible: lib.BoolToPtr(true)}
	assert.NoError(db.Create(&alice).Error)
	bob := models.User{Type: models.UserTypeStudent, Name: "Bob", UnixID: "bob", Visible: lib.BoolToPtr(true)}
	assert.NoError(db.Create(&bob).Error)

	cc := models.CourseCanonical{CrseID: "100136", Title: "Data Structures"}
	assert.NoError(db.Create(&cc).Error)
	cis := models.CourseInSemester{CourseCanonicalID: cc.ID, Strm: 1251, Year: 2025, Term: "Fall",
		Status: models.CourseOfferingStatusOffered, Title: cc.Title}
	assert.NoError(db.Create(&cis).Error)
	section := func(classNbr int, component string) models.CourseOffering {
		o := models.CourseOffering{CourseCanonicalID: cc.ID, CourseInSemesterID: &cis.ID, Strm: 1251, Year: 2025, Term: "Fall",
			Section: "01", ClassNbr: classNbr, Title: cc.Title, Component: component, Status: models.CourseOfferingStatusOffered}
		assert.NoError(db.Create(&o).Error)
		return o
	}
	lec, lab := section(1001, "LEC"), section(1002, "LAB")

	review := func(author models.User, offering *uint, source string) {
		assert.NoError(db.Create(&models.FactrakSurvey{UserID: author.ID, ProfessorID: prof.ID, CourseID: fc.ID,
			CourseCanonicalID: &cc.ID, CourseOfferingID: offering, CourseLinkVerificationSource: source}).Error)
	}
	review(alice, &lec.ID, models.CourseLinkVerificationSourceVerified)
	review(alice, &lab.ID, models.CourseLinkVerificationSourceAuthorConfirmed)
	review(bob, nil, models.CourseLinkVerificationSourceConfirmed)

	courses, sections, err := SyncReviewedUserCourses(db, zap.NewNop().Sugar())
	assert.NoError(err)
	assert.EqualValues(1, courses)
	assert.EqualValues(2, sections)

	var entries []models.UserCourse
	assert.NoError(db.Preload("Offerings").Find(&entries).Error)
	if assert.Len(entries, 1) {
		assert.Equal(alice.ID, entries[0].UserID)
		assert.Equal(cis.ID, *entries[0].CourseInSemesterID)
		assert.Len(entries[0].Offerings, 2)
	}
	linkedTo := func(author models.User) []*uint {
		var reviews []models.FactrakSurvey
		assert.NoError(db.Where("user_id = ?", author.ID).Order("id").Find(&reviews).Error)
		var out []*uint
		for _, r := range reviews {
			out = append(out, r.UserCourseID)
		}
		return out
	}
	assert.Equal([]*uint{&entries[0].ID, &entries[0].ID}, linkedTo(alice))
	assert.Equal([]*uint{nil}, linkedTo(bob))

	cis2 := models.CourseInSemester{CourseCanonicalID: cc.ID, Strm: 1271, Year: 2027, Term: "Fall",
		Status: models.CourseOfferingStatusOffered, Title: cc.Title}
	assert.NoError(db.Create(&cis2).Error)
	bobOld := models.UserCourse{UserID: bob.ID, CourseCanonicalID: cc.ID, CourseInSemesterID: &cis.ID}
	bobNew := models.UserCourse{UserID: bob.ID, CourseCanonicalID: cc.ID, CourseInSemesterID: &cis2.ID}
	assert.NoError(db.Create(&bobOld).Error)
	assert.NoError(db.Create(&bobNew).Error)
	fall, year := models.FactrakSurveySemesterSeasonFall, 2024
	assert.NoError(db.Model(&models.FactrakSurvey{}).Where("user_id = ?", bob.ID).
		Updates(map[string]interface{}{"semester_season": fall, "semester_year": year}).Error)
	_, _, err = SyncReviewedUserCourses(db, nil)
	assert.NoError(err)
	assert.Equal([]*uint{&bobOld.ID}, linkedTo(bob))

	assert.NoError(db.Model(&entries[0]).Update("hidden", true).Error)
	courses, sections, err = SyncReviewedUserCourses(db, nil)
	assert.NoError(err)
	assert.Zero(courses)
	assert.Zero(sections)
	var again models.UserCourse
	assert.NoError(db.First(&again, entries[0].ID).Error)
	assert.True(again.Hidden)

	assert.NoError(db.Unscoped().Where("user_course_id = ? AND course_offering_id = ?", entries[0].ID, lab.ID).
		Delete(&models.UserCourseOffering{}).Error)
	courses, sections, err = SyncReviewedUserCourses(db, nil)
	assert.NoError(err)
	assert.Zero(courses)
	assert.Zero(sections)
	var kept []models.UserCourseOffering
	assert.NoError(db.Where("user_course_id = ?", entries[0].ID).Find(&kept).Error)
	if assert.Len(kept, 1) {
		assert.Equal(lec.ID, kept[0].CourseOfferingID)
	}
}
