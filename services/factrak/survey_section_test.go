package factrak_test

import (
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/factrak"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

// Course, term and (with one instructor) professor come from the section
func TestController_CreateSurveyFromSection(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	student := models.User{Type: models.UserTypeStudent, Name: "Student", UnixID: "s1"}
	prof := models.User{Type: models.UserTypeProfessor, Name: "Prof", UnixID: "p1", AtWilliams: lib.BoolToPtr(true)}
	visiting := models.User{Type: models.UserTypeProfessor, Name: "Visiting", UnixID: "v1", AtWilliams: lib.BoolToPtr(false)}
	area := models.AreaOfStudy{Name: "Computer Science", Abbreviation: "CSCI", Department: &models.Department{Name: "CS"}}
	assert.NoError(db.Create(&student).Create(&prof).Create(&visiting).Create(&area).Error)

	canonical := models.CourseCanonical{CrseID: "100136", Title: "Data Structures"}
	assert.NoError(db.Create(&canonical).Error)
	assert.NoError(db.Create(&models.CourseListing{CourseCanonicalID: canonical.ID, Subject: "CSCI", Number: 136,
		FirstValidYear: 2020, LastValidYear: 2030}).Error)
	cis := models.CourseInSemester{CourseCanonicalID: canonical.ID, Strm: 1253, Year: 2025, Term: "Spring",
		Status: models.CourseOfferingStatusOffered, Title: canonical.Title}
	assert.NoError(db.Create(&cis).Error)
	section := func(classNbr int, teachers ...models.User) {
		o := models.CourseOffering{CourseCanonicalID: canonical.ID, CourseInSemesterID: &cis.ID, Strm: 1253, Year: 2025,
			Term: "Spring", Section: "01", ClassNbr: classNbr, Title: canonical.Title, Component: "LEC",
			Status: models.CourseOfferingStatusOffered}
		assert.NoError(db.Create(&o).Error)
		for i := range teachers {
			assert.NoError(db.Create(&models.CourseOfferingInstructor{CourseOfferingID: o.ID, UserID: &teachers[i].ID,
				NameAsListed: teachers[i].Name}).Error)
		}
	}
	section(3001, prof)
	section(3002, prof, visiting)

	router := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf)
	utils.AddUserContexts(router, student.ID)
	SetupRouter(router, db, utils.SetupConfig(), zaptest.NewLogger(t).Sugar())

	strm, one, two, missing := 1253, 3001, 3002, 9999
	createSurveyExpectError(assert, router, SurveyCreateParams{Strm: &strm, ClassNbr: &missing,
		Comment: generateSurveyTestComment()}, lib.ErrorSurveySectionNotFound)
	createSurveyExpectError(assert, router, SurveyCreateParams{Strm: &strm, ClassNbr: &two,
		Comment: generateSurveyTestComment()}, lib.ErrorSurveyProfessorNotFound)
	other := models.User{Type: models.UserTypeProfessor, Name: "Other", UnixID: "o1", AtWilliams: lib.BoolToPtr(true)}
	assert.NoError(db.Create(&other).Error)
	createSurveyExpectError(assert, router, SurveyCreateParams{Strm: &strm, ClassNbr: &one, ProfessorID: &other.ID,
		Comment: generateSurveyTestComment()}, lib.ErrorSurveySectionProfessor)

	survey := createSurveyExpectSuccess(assert, router, SurveyCreateParams{Strm: &strm, ClassNbr: &one,
		Comment: generateSurveyTestComment()})
	var saved models.FactrakSurvey
	assert.NoError(db.Preload("Course").First(&saved, survey.ID).Error)
	assert.Equal(prof.ID, saved.ProfessorID)
	assert.Equal("136", saved.Course.Number)
	assert.Equal(area.ID, *saved.Course.AreaOfStudyID)
	assert.Equal(canonical.ID, *saved.CourseCanonicalID)
	assert.NotNil(saved.CourseOfferingID)
	assert.Equal(models.CourseLinkVerificationSourceAuthorConfirmed, saved.CourseLinkVerificationSource)
	assert.Equal(models.FactrakSurveySemesterSeasonSpring, *saved.SemesterSeason)
	assert.Equal(2025, *saved.SemesterYear)

	var mine []models.UserCourse
	assert.NoError(db.Preload("Offerings").Where("user_id = ?", student.ID).Find(&mine).Error)
	if assert.Len(mine, 1) {
		assert.Equal(cis.ID, *mine[0].CourseInSemesterID)
		assert.Len(mine[0].Offerings, 1)
		var linked models.FactrakSurvey
		assert.NoError(db.First(&linked, saved.ID).Error)
		assert.Equal(&mine[0].ID, linked.UserCourseID)
	}

	// Legacy rules turn away a professor not at Williams and not teaching this year
	createSurveyExpectError(assert, router, SurveyCreateParams{Strm: &strm, ClassNbr: &two, ProfessorID: &visiting.ID,
		Comment: generateSurveyTestComment()}, lib.ErrorSurveyProfessorNotFound)

	// A lecturer typed "staff" can be reviewed for a section they taught
	lecturer := models.User{Type: models.UserTypeStaff, Name: "Lecturer", UnixID: "l1", AtWilliams: lib.BoolToPtr(true)}
	assert.NoError(db.Create(&lecturer).Error)
	section(3003, lecturer)
	three := 3003
	survey = createSurveyExpectSuccess(assert, router, SurveyCreateParams{Strm: &strm, ClassNbr: &three,
		Comment: generateSurveyTestComment()})
	var lectured models.FactrakSurvey
	assert.NoError(db.First(&lectured, survey.ID).Error)
	assert.Equal(lecturer.ID, lectured.ProfessorID)
}
