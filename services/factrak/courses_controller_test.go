package factrak_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
)

// Check that a slice of courses matches the expected slice, by comparing course IDs
func EqualCourseIDs(expected, resp []models.Course) error {
	if len(resp) != len(expected) {
		return errors.New(fmt.Sprintf("Expected length %d, got %d", len(expected), len(resp)))
	}

	for i, v := range expected {
		if v.ID != resp[i].ID {
			return errors.New(fmt.Sprintf("At index %d, expected ID %d, got %d", i, v.ID, resp[i].ID))
		}
	}
	return nil
}

// Unmarshal a slice of courses from an http response
func GetCoursesFromResp(assert *testify.Assertions, w *httptest.ResponseRecorder) []models.Course {
	respData := utils.GetGoodResp(assert, w)
	var resp []models.Course
	assert.NoError(json.Unmarshal(respData.Data, &resp))
	return resp
}

// Unmarshal a single course from an http response
func GetCourseFromResp(assert *testify.Assertions, w *httptest.ResponseRecorder) models.Course {
	respData := utils.GetGoodResp(assert, w)
	var resp models.Course
	assert.NoError(json.Unmarshal(respData.Data, &resp))
	return resp
}

func TestController_ListCourses(t *testing.T) {
	// Set up the test environment
	assert, db, router := SetupFactrakTest(t)

	// Need this to satisfy not null
	dept := models.Department{
		Name: "Computer Science",
	}
	area := models.AreaOfStudy{
		Name:         "Computer Science",
		Abbreviation: "CSCI",
		Department:   &dept,
	}
	assert.NoError(db.Create(&dept).Create(&area).Error)

	// Insert test user into db
	c1 := models.Course{
		Number:      "Course 1",
		AreaOfStudy: &area,
	}
	c2 := models.Course{
		Number:      "Course 2",
		AreaOfStudy: &area,
	}
	c3 := models.Course{
		Number:      "Course 3",
		AreaOfStudy: &area,
	}

	assert.NoError(db.Create(&c1).Create(&c2).Create(&c3).Error)

	// Get test courses
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/courses", nil)
	assert.NoError(err)

	// Check if response contains the correct courses
	resp := GetCoursesFromResp(assert, w)
	assert.NoError(EqualCourseIDs([]models.Course{c1, c2, c3}, resp))
}

func TestController_ListCoursesRanked(t *testing.T) {
	// Set up the test environment
	assert, db, router := SetupFactrakTest(t)

	// Insert test professors and students into db
	p1 := models.User{
		Type:       models.UserTypeProfessor,
		Name:       "Professor 1",
		UnixID:     "p1",
		AtWilliams: lib.BoolToPtr(true),
	}
	// Other prof
	p2 := models.User{
		Type:       models.UserTypeProfessor,
		Name:       "Professor 2",
		UnixID:     "p2",
		AtWilliams: lib.BoolToPtr(true),
	}
	err := db.Create(&p1).Create(&p2).Error
	assert.NoError(err)

	// Need this to satisfy not null
	dept := models.Department{
		Name: "Computer Science",
	}
	area1 := models.AreaOfStudy{
		Name:         "Computer Science",
		Abbreviation: "CSCI",
		Department:   &dept,
	}
	area2 := models.AreaOfStudy{
		Name:         "Human Computer Interaction",
		Abbreviation: "HCI",
		Department:   &dept,
	}
	assert.NoError(db.Create(&dept).Create(&area1).Create(&area2).Error)

	// Insert test course into db
	c1 := models.Course{
		Number:      "Course 1",
		AreaOfStudy: &area1,
	}
	c2 := models.Course{
		Number:      "Course 2",
		AreaOfStudy: &area2,
	}
	c3 := models.Course{
		Number:      "Course 3",
		AreaOfStudy: &area2,
	}
	assert.NoError(db.Create(&c1).Create(&c2).Create(&c3).Error)

	// Insert test students into db
	students := make([]*models.User, 10)
	for i := range students {
		name := fmt.Sprintf("Student %d", i)
		unix := fmt.Sprintf("s%d", i)
		students[i] = &models.User{Type: models.UserTypeStudent, Name: name, UnixID: unix}
		assert.NoError(db.Create(students[i]).Error)
	}

	// Insert test surveys into db (need 10)
	surveys := make([]*models.FactrakSurvey, 35)
	for i := range students {
		comment := fmt.Sprintf("Survey %d", i)
		surveys[i] = &models.FactrakSurvey{
			User:                 students[i],
			Professor:            &p1,
			Course:               &c1,
			Comment:              comment,
			WouldRecommendCourse: lib.BoolToPtr(true),
			CourseWorkload:       lib.IntToPtr(3),
			WouldTakeAnother:     lib.BoolToPtr(true),
			CourseStimulating:    lib.IntToPtr(5),
		}
		surveys[i+10] = &models.FactrakSurvey{
			User:              students[i],
			Professor:         &p1,
			Course:            &c2,
			Comment:           comment,
			CourseWorkload:    lib.IntToPtr(9),
			CourseStimulating: lib.IntToPtr(8),
		}
		surveys[i+20] = &models.FactrakSurvey{
			User:           students[i],
			Professor:      &p2,
			Course:         &c3,
			Comment:        comment,
			CourseWorkload: lib.IntToPtr(4),
		}
		assert.NoError(db.Create(surveys[i]).Create(surveys[i+10]).Create(surveys[i+20]).Error)
	}

	for i := 1; i < 5; i++ {
		comment := fmt.Sprintf("Survey %d", i)
		surveys[i+30] = &models.FactrakSurvey{
			User:                 students[i],
			Professor:            &p2,
			Course:               &c3,
			Comment:              comment,
			CourseWorkload:       lib.IntToPtr(0),
			WouldRecommendCourse: lib.BoolToPtr(false),
		}
		assert.NoError(db.Create(surveys[i+30]).Error)
	}

	// Test 1: Get courses ranked by invalid metric (should not work)
	apiErr := lib.ErrorInvalidRankingMetric
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/courses?metric=would_take_another&ascending=true", nil)
	assert.NoError(err)
	utils.CheckRespError(assert, w, apiErr)

	// Test 2: Get courses ranked by workload
	w, err = utils.DoHTTPReq(router, http.MethodGet, "/courses?metric=course_workload&ascending=true", nil)
	assert.NoError(err)
	resp := GetCoursesFromResp(assert, w)
	assert.NoError(EqualCourseIDs([]models.Course{c3, c1, c2}, resp))

	// Test 3: Get courses ranked by workload, for one professor
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/courses?metric=course_workload&ascending=true&professorID=%d", p1.ID), nil)
	assert.NoError(err)
	resp = GetCoursesFromResp(assert, w)
	assert.NoError(EqualCourseIDs([]models.Course{c1, c2}, resp))

	// Test 4: Get courses ranked by whether students would recommend them
	w, err = utils.DoHTTPReq(router, http.MethodGet, "/courses?metric=would_recommend_course", nil)
	assert.NoError(err)
	resp = GetCoursesFromResp(assert, w)
	assert.NoError(EqualCourseIDs([]models.Course{c1}, resp))

	// Test 5: Get courses ranked by how stimulating they are, limited to one area of study
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/courses?metric=course_stimulating&areaOfStudyID=%d", area2.ID), nil)
	assert.NoError(err)
	resp = GetCoursesFromResp(assert, w)
	assert.NoError(EqualCourseIDs([]models.Course{c2}, resp))

	// Test 6: Get courses ranked by workload, with pagination
	w, err = utils.DoHTTPReq(router, http.MethodGet, "/courses?metric=course_workload&ascending=true&limit=2", nil)
	assert.NoError(err)
	resp = GetCoursesFromResp(assert, w)
	assert.NoError(EqualCourseIDs([]models.Course{c3, c1}, resp))

	//Test 6, part 2 of pagination
	w, err = utils.DoHTTPReq(router, http.MethodGet, "/courses?metric=course_workload&ascending=true&limit=2&offset=2", nil)
	assert.NoError(err)
	resp = GetCoursesFromResp(assert, w)
	assert.NoError(EqualCourseIDs([]models.Course{c2}, resp))
}

func TestController_GetCourse(t *testing.T) {
	// Set up the test environment
	assert, db, router := SetupFactrakTest(t)

	p1 := models.User{
		Type:   models.UserTypeProfessor,
		Name:   "Professor 1",
		UnixID: "p1",
	}
	s1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "Student",
		UnixID: "s1",
	}
	// Need this to satisfy not null
	dept := models.Department{
		Name: "Computer Science",
	}
	area := models.AreaOfStudy{
		Name:         "Computer Science",
		Abbreviation: "CSCI",
		Department:   &dept,
	}
	assert.NoError(db.Create(&dept).Create(&area).Create(&s1).Create(&p1).Error)

	// Insert test user into db
	c1 := models.Course{
		Number:      "Course 1",
		AreaOfStudy: &area,
	}
	c2 := models.Course{
		Number:      "Course 2",
		AreaOfStudy: &area,
	}

	assert.NoError(db.Create(&c1).Create(&c2).Error)

	fs1 := models.FactrakSurvey{
		User:      &s1,
		Professor: &p1,
		Course:    &c1,
		Comment:   "Survey 1",
	}
	fs2 := models.FactrakSurvey{
		User:      &s1,
		Professor: &p1,
		Course:    &c1,
		Comment:   "Survey 2",
	}

	assert.NoError(db.Create(&fs1).Create(&fs2).Error)

	// Get test course
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/courses/%d", c1.ID), nil)
	assert.NoError(err)

	// Check if correct course
	resp := GetCourseFromResp(assert, w)
	assert.Equal(c1.Number, resp.Number)
	assert.Len(resp.FactrakSurveys, 2)

	// Check if we got surveys (in reverse order)
	assert.Equal(fs2.Comment, resp.FactrakSurveys[0].Comment)
	assert.Equal(fs1.Comment, resp.FactrakSurveys[1].Comment)

	// Check if we removed sensitive user data
	assert.Zero(resp.FactrakSurveys[0].UserID)
	assert.Nil(resp.FactrakSurveys[0].User)

	/* Get test bad course id (expect failure) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/courses/%d", 42), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusNotFound, w.Code)
}

func TestController_GetCourseWithProfessor(t *testing.T) {
	// Set up the test environment
	assert, db, router := SetupFactrakTest(t)

	// Insert test user into db
	p1 := models.User{
		Type:   models.UserTypeProfessor,
		Name:   "Professor 1",
		UnixID: "p1",
	}
	// Student
	s1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "Student 1",
		UnixID: "s1",
	}
	// Other prof
	p2 := models.User{
		Type:   models.UserTypeProfessor,
		Name:   "Professor 2",
		UnixID: "p2",
	}
	err := db.Create(&p1).Create(&s1).Create(&p2).Error
	assert.NoError(err)

	// Need this to satisfy not null
	dept := models.Department{
		Name: "Computer Science",
	}
	area := models.AreaOfStudy{
		Name:         "Computer Science",
		Abbreviation: "CSCI",
		Department:   &dept,
	}
	assert.NoError(db.Create(&dept).Create(&area).Error)

	// Insert test course into db
	c1 := models.Course{
		Number:      "Course 1",
		AreaOfStudy: &area,
	}
	c2 := models.Course{
		Number:      "Course 2",
		AreaOfStudy: &area,
	}

	assert.NoError(db.Create(&c1).Create(&c2).Error)

	fs1 := models.FactrakSurvey{
		User:      &s1,
		Professor: &p1,
		Course:    &c1,
		Comment:   "Survey 1",
	}
	fs2 := models.FactrakSurvey{
		User:      &s1,
		Professor: &p2,
		Course:    &c1,
		Comment:   "Survey 2",
	}
	fs3 := models.FactrakSurvey{
		User:      &s1,
		Professor: &p1,
		Course:    &c2,
		Comment:   "Survey 3",
	}
	fs4 := models.FactrakSurvey{
		User:      &s1,
		Professor: &p1,
		Course:    &c1,
		Comment:   "Survey 4",
	}

	assert.NoError(db.Create(&fs1).Create(&fs2).Create(&fs3).Create(&fs4).Error)

	/* Get course 1 with prof 1 (expect success) */
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/courses/%d?professorID=%d", c1.ID, p1.ID), nil)
	assert.NoError(err)

	// Check if correct course
	resp := GetCourseFromResp(assert, w)
	assert.Equal(c1.ID, resp.ID)
	assert.Len(resp.FactrakSurveys, 2)

	// Check if we got surveys (only courses) (in reverse order)
	assert.Equal(fs4.Comment, resp.FactrakSurveys[0].Comment)
	assert.Equal(fs1.Comment, resp.FactrakSurveys[1].Comment)

	// Check if we removed sensitive user data
	assert.Zero(resp.FactrakSurveys[0].UserID)
	assert.Nil(resp.FactrakSurveys[0].User)

	/* Get course 1 with a random prof (expect empty) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/courses/%d?professorID=%d", c1.ID, 42), nil)
	assert.NoError(err)

	// Check if correct course
	resp = GetCourseFromResp(assert, w)
	assert.Equal(c1.ID, resp.ID)
	assert.Len(resp.FactrakSurveys, 0)
}

func TestController_ListCourseSurveys(t *testing.T) {
	// Set up the test environment
	assert, db, router := SetupFactrakTest(t)

	p1 := models.User{
		Type:   models.UserTypeProfessor,
		Name:   "Professor 1",
		UnixID: "p1",
	}
	s1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "Student 1",
		UnixID: "s1",
	}
	// Need this to satisfy not null
	dept := models.Department{
		Name: "Computer Science",
	}
	area := models.AreaOfStudy{
		Name:         "Computer Science",
		Abbreviation: "CSCI",
		Department:   &dept,
	}
	assert.NoError(db.Create(&dept).Create(&area).Create(&s1).Create(&p1).Error)

	// Insert test user into db
	c1 := models.Course{
		Number:      "Course 1",
		AreaOfStudy: &area,
	}
	c2 := models.Course{
		Number:      "Course 2",
		AreaOfStudy: &area,
	}

	assert.NoError(db.Create(&c1).Create(&c2).Error)

	fs1 := models.FactrakSurvey{
		User:      &s1,
		Professor: &p1,
		Course:    &c1,
		Comment:   "Survey 1",
	}
	fs2 := models.FactrakSurvey{
		User:      &s1,
		Professor: &p1,
		Course:    &c2,
		Comment:   "Survey 2",
	}
	fs3 := models.FactrakSurvey{
		User:      &s1,
		Professor: &p1,
		Course:    &c1,
		Comment:   "Survey 3",
	}

	assert.NoError(db.Create(&fs1).Create(&fs2).Create(&fs3).Error)

	/* Get surveys for course 1 (expect success) */
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/courses/%d/surveys", c1.ID), nil)
	assert.NoError(err)

	// Check if response has correct surveys in reverse chronological order
	resp := GetSurveysFromResp(assert, w)
	assert.NoError(EqualSurveyIDs([]models.FactrakSurvey{fs3, fs1}, resp))

	// Assert that userID is not returned
	assert.Zero(resp[0].UserID)
	assert.Nil(resp[0].User)

	/* Get nonexistent course (expect failure) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/courses/%d/surveys", 42), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusNotFound, w.Code)

	/* Get surveys for course 2 (expect success) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/courses/%d/surveys", c2.ID), nil)
	assert.NoError(err)

	// Check if response has survey 2
	resp = GetSurveysFromResp(assert, w)
	assert.NoError(EqualSurveyIDs([]models.FactrakSurvey{fs2}, resp))
}

func TestController_ListCourseSurveysWithProfessor(t *testing.T) {
	// Set up the test environment
	assert, db, router := SetupFactrakTest(t)

	// Insert test user into db
	p1 := models.User{
		Type:   models.UserTypeProfessor,
		Name:   "Professor 1",
		UnixID: "p1",
	}
	// Student
	s1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "Student 1",
		UnixID: "s1",
	}
	// Other prof
	p2 := models.User{
		Type:   models.UserTypeProfessor,
		Name:   "Professor 2",
		UnixID: "p2",
	}
	err := db.Create(&p1).Create(&s1).Create(&p2).Error
	assert.NoError(err)

	// Need this to satisfy not null
	dept := models.Department{
		Name: "Computer Science",
	}
	area := models.AreaOfStudy{
		Name:         "Computer Science",
		Abbreviation: "CSCI",
		Department:   &dept,
	}
	assert.NoError(db.Create(&dept).Create(&area).Error)

	// Insert test course into db
	c1 := models.Course{
		Number:      "Course 1",
		AreaOfStudy: &area,
	}
	c2 := models.Course{
		Number:      "Course 2",
		AreaOfStudy: &area,
	}

	assert.NoError(db.Create(&c1).Create(&c2).Error)

	fs1 := models.FactrakSurvey{
		User:      &s1,
		Professor: &p1,
		Course:    &c1,
		Comment:   "Survey 1",
	}
	fs2 := models.FactrakSurvey{
		User:      &s1,
		Professor: &p2,
		Course:    &c1,
		Comment:   "Survey 2",
	}
	fs3 := models.FactrakSurvey{
		User:      &s1,
		Professor: &p1,
		Course:    &c2,
		Comment:   "Survey 3",
	}
	fs4 := models.FactrakSurvey{
		User:      &s1,
		Professor: &p1,
		Course:    &c1,
		Comment:   "Survey 4",
	}

	assert.NoError(db.Create(&fs1).Create(&fs2).Create(&fs3).Create(&fs4).Error)

	/* Get surveys for prof 1 and course 1 (expect success) */
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/professors/%d/surveys?courseID=%d", p1.ID, c1.ID), nil)
	assert.NoError(err)

	// Check if response is valid and has surveys, in reverse chronological order
	resp := GetSurveysFromResp(assert, w)
	assert.NoError(EqualSurveyIDs([]models.FactrakSurvey{fs4, fs1}, resp))

	// Check if we removed sensitive user data
	assert.Zero(resp[0].UserID)
	assert.Nil(resp[0].User)

	/* Get prof 1 with a nonexistent course (expect empty) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/professors/%d/surveys?courseID=%d", p1.ID, 42), nil)
	assert.NoError(err)

	// Check that the response is valid but has no surveys
	resp = GetSurveysFromResp(assert, w)
	assert.Len(resp, 0)
}

func TestController_ListCourseProfessors(t *testing.T) {
	// Set up the test environment
	assert, db, router := SetupFactrakTest(t)

	// Insert test user into db
	p1 := models.User{
		Type:   models.UserTypeProfessor,
		Name:   "Professor 1",
		UnixID: "p1",
	}
	s1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "Student",
		UnixID: "s1",
	}
	s2 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "Student",
		UnixID: "s2",
	}
	p2 := models.User{
		Type:   models.UserTypeProfessor,
		Name:   "Professor 2",
		UnixID: "p2",
	}
	assert.NoError(db.Create(&p1).Create(&s1).Create(&s2).Create(&p2).Error)

	// Need this to satisfy not null
	dept := models.Department{
		Name: "Computer Science",
	}
	area := models.AreaOfStudy{
		Name:         "Computer Science",
		Abbreviation: "CSCI",
		Department:   &dept,
	}
	assert.NoError(db.Create(&dept).Create(&area).Error)

	c1 := models.Course{
		Number:      "Course 1 by p1",
		AreaOfStudy: &area,
	}
	c2 := models.Course{
		Number:      "Course 2 by p2",
		AreaOfStudy: &area,
	}
	c3 := models.Course{
		Number:      "Course 3 by p1, p2",
		AreaOfStudy: &area,
	}
	assert.NoError(db.Create(&c1).Create(&c2).Create(&c3).Error)

	fs1 := models.FactrakSurvey{
		User:      &s1,
		Professor: &p1,
		Course:    &c1,
		Comment:   "Survey 1",
	}
	fs2 := models.FactrakSurvey{
		User:      &s2,
		Professor: &p1,
		Course:    &c1,
		Comment:   "Survey 2",
	}
	fs3 := models.FactrakSurvey{
		User:      &s1,
		Professor: &p2,
		Course:    &c2,
		Comment:   "Survey 3",
	}
	fs4 := models.FactrakSurvey{
		User:      &s2,
		Professor: &p1,
		Course:    &c3,
		Comment:   "Survey 4",
	}
	fs5 := models.FactrakSurvey{
		User:      &s1,
		Professor: &p2,
		Course:    &c3,
		Comment:   "Survey 5",
	}

	assert.NoError(db.Create(&fs1).Create(&fs2).Create(&fs3).Create(&fs4).Create(&fs5).Error)

	/* Get professors for test course 3 (expect success) */
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/courses/%d/professors", c3.ID), nil)
	assert.NoError(err)

	// Check if response contains profs 1 and 2, in order of surveys created first to last
	resp := GetUsersFromResp(assert, w)
	assert.Len(resp, 2)
	assert.NoError(EqualUserIDs([]models.User{p1, p2}, resp))

	/* Get bad course (expect failure) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/courses/%d/professors", 42), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusNotFound, w.Code)

	/* Get professors for test course 1 (expect success) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/courses/%d/professors", c1.ID), nil)
	assert.NoError(err)

	// Check if it is professor 1
	resp = GetUsersFromResp(assert, w)
	assert.NoError(EqualUserIDs([]models.User{p1}, resp))
}

func TestController_GetCourseRatings(t *testing.T) {
	// Set up the test environment
	assert, db, router := SetupFactrakTest(t)

	p1 := models.User{
		Type:   models.UserTypeProfessor,
		Name:   "Professor 1",
		UnixID: "p1",
	}
	s1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "Student 1",
		UnixID: "s1",
	}
	// Need this to satisfy not null
	dept := models.Department{
		Name: "Computer Science",
	}
	area := models.AreaOfStudy{
		Name:         "Computer Science",
		Abbreviation: "CSCI",
		Department:   &dept,
	}
	assert.NoError(db.Create(&dept).Create(&area).Create(&s1).Create(&p1).Error)

	// Insert test user into db
	c1 := models.Course{
		Number:      "Course 1",
		AreaOfStudy: &area,
	}
	c2 := models.Course{
		Number:      "Course 2",
		AreaOfStudy: &area,
	}

	assert.NoError(db.Create(&c1).Create(&c2).Error)

	fs1 := models.FactrakSurvey{
		User:                 &s1,
		Professor:            &p1,
		Course:               &c1,
		Comment:              "Survey 1",
		WouldRecommendCourse: lib.BoolToPtr(true),
		CourseWorkload:       lib.IntToPtr(0),
		CourseStimulating:    lib.IntToPtr(1),
		WouldTakeAnother:     lib.BoolToPtr(false),
		Approachability:      lib.IntToPtr(1),
		LeadLecture:          lib.IntToPtr(7),
		PromoteDiscussion:    lib.IntToPtr(4),
		OutsideHelpfulness:   nil,
	}
	fs2 := models.FactrakSurvey{
		User:                 &s1,
		Professor:            &p1,
		Course:               &c1,
		Comment:              "Survey 2",
		WouldRecommendCourse: lib.BoolToPtr(true),
		CourseWorkload:       lib.IntToPtr(0),
		CourseStimulating:    lib.IntToPtr(3),
		WouldTakeAnother:     lib.BoolToPtr(false),
		Approachability:      lib.IntToPtr(2),
		LeadLecture:          lib.IntToPtr(7),
		PromoteDiscussion:    lib.IntToPtr(6),
		OutsideHelpfulness:   lib.IntToPtr(1),
	}
	fs3 := models.FactrakSurvey{
		User:                 &s1,
		Professor:            &p1,
		Course:               &c1,
		Comment:              "Survey 3",
		WouldRecommendCourse: lib.BoolToPtr(true),
		CourseWorkload:       lib.IntToPtr(0),
		CourseStimulating:    lib.IntToPtr(5),
		WouldTakeAnother:     lib.BoolToPtr(false),
		Approachability:      lib.IntToPtr(3),
		LeadLecture:          lib.IntToPtr(7),
		PromoteDiscussion:    lib.IntToPtr(4),
		OutsideHelpfulness:   lib.IntToPtr(3),
	}
	fs4 := models.FactrakSurvey{
		User:                 &s1,
		Professor:            &p1,
		Course:               &c1,
		Comment:              "Survey 4",
		WouldRecommendCourse: lib.BoolToPtr(false),
		CourseWorkload:       lib.IntToPtr(0),
		CourseStimulating:    lib.IntToPtr(7),
		WouldTakeAnother:     lib.BoolToPtr(false),
		Approachability:      lib.IntToPtr(7),
		LeadLecture:          lib.IntToPtr(7),
		PromoteDiscussion:    lib.IntToPtr(6),
		OutsideHelpfulness:   lib.IntToPtr(4),
	}
	fs5 := models.FactrakSurvey{
		User:                 &s1,
		Professor:            &p1,
		Course:               &c2,
		Comment:              "Survey 5",
		WouldRecommendCourse: lib.BoolToPtr(true),
		CourseWorkload:       lib.IntToPtr(7),
		CourseStimulating:    lib.IntToPtr(7),
		WouldTakeAnother:     lib.BoolToPtr(true),
		Approachability:      lib.IntToPtr(7),
		LeadLecture:          lib.IntToPtr(7),
		PromoteDiscussion:    lib.IntToPtr(7),
		OutsideHelpfulness:   lib.IntToPtr(7),
	}

	assert.NoError(db.Create(&fs1).Create(&fs2).Create(&fs3).Create(&fs4).Create(&fs5).Error)

	/* Get test course 1 (expect success) */
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/courses/%d/ratings", c1.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.FactrakSurveyAvgRatings
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	assert.Equal(models.FactrakSurveyAvgRatings{
		AvgWouldRecommendCourse: 0.75,
		NumWouldRecommendCourse: 4,
		AvgCourseWorkload:       0,
		NumCourseWorkload:       4,
		AvgCourseStimulating:    4,
		NumCourseStimulating:    4,
		AvgWouldTakeAnother:     0,
		NumWouldTakeAnother:     4,
		AvgApproachability:      3.25,
		NumApproachability:      4,
		AvgLeadLecture:          7,
		NumLeadLecture:          4,
		AvgPromoteDiscussion:    5,
		NumPromoteDiscussion:    4,
		AvgOutsideHelpfulness:   2.6666666666666665,
		NumOutsideHelpfulness:   3,
	}, resp)
}

func TestController_GetCourseRatingsWithProfessor(t *testing.T) {
	// Set up the test environment
	assert, db, router := SetupFactrakTest(t)

	// Insert test user into db
	p1 := models.User{
		Type:   models.UserTypeProfessor,
		Name:   "Professor 1",
		UnixID: "p1",
	}
	// Student
	s1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "Student 1",
		UnixID: "s1",
	}
	// Other prof
	p2 := models.User{
		Type:   models.UserTypeProfessor,
		Name:   "Professor 2",
		UnixID: "p2",
	}
	err := db.Create(&p1).Create(&s1).Create(&p2).Error
	assert.NoError(err)

	// Need this to satisfy not null
	dept := models.Department{
		Name: "Computer Science",
	}
	area := models.AreaOfStudy{
		Name:         "Computer Science",
		Abbreviation: "CSCI",
		Department:   &dept,
	}
	assert.NoError(db.Create(&dept).Create(&area).Error)

	// Insert test course into db
	c1 := models.Course{
		Number:      "Course 1",
		AreaOfStudy: &area,
	}
	c2 := models.Course{
		Number:      "Course 2",
		AreaOfStudy: &area,
	}

	assert.NoError(db.Create(&c1).Create(&c2).Error)

	fs1 := models.FactrakSurvey{
		User:                 &s1,
		Professor:            &p1,
		Course:               &c1,
		Comment:              "Survey 1",
		WouldRecommendCourse: lib.BoolToPtr(true),
		CourseWorkload:       lib.IntToPtr(2),
	}
	fs2 := models.FactrakSurvey{
		User:                 &s1,
		Professor:            &p2,
		Course:               &c1,
		Comment:              "Survey 2",
		WouldRecommendCourse: lib.BoolToPtr(true),
		CourseWorkload:       lib.IntToPtr(7),
	}
	fs3 := models.FactrakSurvey{
		User:                 &s1,
		Professor:            &p1,
		Course:               &c2,
		Comment:              "Survey 3",
		WouldRecommendCourse: lib.BoolToPtr(true),
		CourseWorkload:       lib.IntToPtr(7),
	}
	fs4 := models.FactrakSurvey{
		User:                 &s1,
		Professor:            &p1,
		Course:               &c1,
		Comment:              "Survey 4",
		WouldRecommendCourse: lib.BoolToPtr(false),
		CourseWorkload:       lib.IntToPtr(5),
	}

	assert.NoError(db.Create(&fs1).Create(&fs2).Create(&fs3).Create(&fs4).Error)

	/* Get test prof 1's ratings for course 1 (expect success) */
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/professors/%d/ratings?courseID=%d", p1.ID, c1.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	resp := models.FactrakSurveyAvgRatings{}
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct length
	// Check if we got surveys (only courses) (in reverse order)
	assert.Equal(resp.AvgWouldRecommendCourse, 0.5)
	assert.Equal(resp.NumWouldRecommendCourse, 2)
	assert.Equal(resp.AvgCourseWorkload, 3.5)
	assert.Equal(resp.NumCourseWorkload, 2)

	/* Get prof 1's ratings for a random course (expect empty) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/professors/%d/ratings?courseID=%d", p1.ID, 42), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusOK, w.Code)

	respData = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	resp = models.FactrakSurveyAvgRatings{}
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	assert.Zero(resp.AvgWouldRecommendCourse)
	assert.Zero(resp.NumWouldRecommendCourse)
	assert.Zero(resp.AvgCourseWorkload)
	assert.Zero(resp.NumCourseWorkload)
}
