package factrak_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/factrak"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestController_ListProfessors(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	// Insert test user into db
	p1 := models.User{
		Type:   models.UserTypeProfessor,
		Name:   "Prof1",
		UnixID: "p1",
	}
	p2 := models.User{
		Type:   models.UserTypeProfessor,
		Name:   "Prof2",
		UnixID: "p2",
	}
	// Should not show up
	s1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "Student1",
		UnixID: "s1",
	}
	// Not at williams
	p3 := models.User{
		Type:       models.UserTypeProfessor,
		Name:       "Prof3",
		UnixID:     "p3",
		AtWilliams: lib.BoolToPtr(false),
	}
	err := db.Create(&p1).Create(&p2).Create(&s1).Create(&p3).Error
	assert.NoError(err)

	// Have to do this because at_williams is not a pointer. TODO: Change at_williams to a pointer
	err = db.Model(&p3).Update("at_williams", false).Error
	assert.NoError(err)

	// Get test user
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/professors", nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp []models.User
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct users
	assert.Len(resp, 2)
	assert.Equal(p1.ID, resp[0].ID)
	assert.Equal(p1.UnixID, resp[0].UnixID)
	assert.Equal(p2.ID, resp[1].ID)
	assert.Equal(p2.UnixID, resp[1].UnixID)
}

// Check that a slice of users matches an expected slice, by comparing IDs
func EqualUserIDs(expected, resp []models.User) error {
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

func TestController_ListProfessorsRanked(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	// Create department and areas of study
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
		Department: &dept,
	}
	// Third prof
	p3 := models.User{
		Type:       models.UserTypeProfessor,
		Name:       "Professor 3",
		UnixID:     "p3",
		AtWilliams: lib.BoolToPtr(true),
	}
	// Student
	s1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "Student 1",
		UnixID: "s1",
	}
	// Student
	s2 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "Student 2",
		UnixID: "s2",
	}
	err := db.Create(&p1).Create(&p2).Create(&p3).Create(&s1).Create(&s2).Error
	assert.NoError(err)

	// Insert test courses into db
	courses := make([]*models.Course, 10)
	for i := range courses {
		number := fmt.Sprintf("Course %d", i)
		courses[i] = &models.Course{Number: number, AreaOfStudy: &area1}
		assert.NoError(db.Create(courses[i]).Error)
	}

	// Insert test surveys into db (need 10 to count for rankings)
	surveys := make([]*models.FactrakSurvey, 35)
	for i := range courses {
		comment := fmt.Sprintf("Base Survey %d", i)
		surveys[i] = &models.FactrakSurvey{
			User:                 &s1,
			Professor:            &p1,
			Course:               courses[i],
			Comment:              comment,
			WouldRecommendCourse: lib.BoolToPtr(true),
			CourseWorkload:       lib.IntToPtr(3),
			WouldTakeAnother:     lib.BoolToPtr(true),
			Approachability:      lib.IntToPtr(5),
		}
		surveys[i+10] = &models.FactrakSurvey{
			User:            &s1,
			Professor:       &p2,
			Course:          courses[i],
			Comment:         comment,
			CourseWorkload:  lib.IntToPtr(9),
			Approachability: lib.IntToPtr(2),
		}
		surveys[i+20] = &models.FactrakSurvey{
			User:            &s1,
			Professor:       &p3,
			Course:          courses[i],
			Comment:         comment,
			CourseWorkload:  lib.IntToPtr(4),
			Approachability: lib.IntToPtr(9),
		}
		assert.NoError(db.Create(surveys[i]).Create(surveys[i+10]).Create(surveys[i+20]).Error)
	}
	for i := 1; i < 5; i++ {
		comment := fmt.Sprintf("Small-set Survey %d", i)
		surveys[i+30] = &models.FactrakSurvey{
			User:             &s2,
			Professor:        &p3,
			Course:           courses[i],
			Comment:          comment,
			CourseWorkload:   lib.IntToPtr(0),
			WouldTakeAnother: lib.BoolToPtr(false),
		}
		assert.NoError(db.Create(surveys[i+30]).Error)
	}

	// Get professors ranked by invalid metric (should not work)
	apiErr := lib.ErrorInvalidRankingMetric
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/professors?metric=would_recommend_course&direction=true", nil)
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	// Get users ranked by workload
	w, err = utils.DoHTTPReq(router, http.MethodGet, "/professors?metric=course_workload&direction=true", nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Decode and check response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp []models.User
	assert.NoError(json.Unmarshal(respData.Data, &resp))
	assert.NoError(EqualUserIDs([]models.User{p3, p1, p2}, resp))

	// Get users ranked by whether students would take another of their classes
	w, err = utils.DoHTTPReq(router, http.MethodGet, "/professors?metric=would_take_another", nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Decode and check response
	respData = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	assert.NoError(json.Unmarshal(respData.Data, &resp))
	assert.NoError(EqualUserIDs([]models.User{p1}, resp))

	// Get users ranked by approachability, limited to an area of study
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/professors?metric=approachability&areaOfStudyID=%d", area2.ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Decode and check response
	respData = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	assert.NoError(json.Unmarshal(respData.Data, &resp))
	assert.NoError(EqualUserIDs([]models.User{p2}, resp))
}

func TestController_GetProfessor(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	// Insert test user into db
	p1 := models.User{
		Type:   models.UserTypeProfessor,
		Name:   "Professor",
		UnixID: "p1",
	}
	// Should not show up
	s1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "Student",
		UnixID: "s1",
	}
	// Not at williams
	p2 := models.User{
		Type:       models.UserTypeProfessor,
		Name:       "Professor Not At Williams",
		UnixID:     "p2",
		AtWilliams: lib.BoolToPtr(false),
	}
	err := db.Create(&p1).Create(&s1).Create(&p2).Error
	assert.NoError(err)

	// Have to do this because at_williams is not a pointer. TODO: Change at_williams to a pointer
	err = db.Model(&p2).Update("at_williams", false).Error
	assert.NoError(err)

	fs1 := models.FactrakSurvey{
		User:      &s1,
		Professor: &p1,
		Comment:   "Survey 1",
	}
	fs2 := models.FactrakSurvey{
		User:      &s1,
		Professor: &p1,
		Comment:   "Survey 2",
	}
	assert.NoError(db.Create(&fs1).Create(&fs2).Error)

	/* Get test prof 1 (expect success) */
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/professors/%d", p1.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	resp := models.User{}
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct prof
	assert.Equal(p1.ID, resp.ID)
	assert.Len(resp.FactrakSurveys, 2)
	// Assert that we don't include ProfessorFactrakSurveys
	assert.Nil(resp.ProfessorFactrakSurveys)

	// Check if we got surveys (in reverse order)
	assert.Equal(resp.FactrakSurveys[0].Comment, fs2.Comment)
	assert.Equal(resp.FactrakSurveys[1].Comment, fs1.Comment)

	// Check if we removed sensitive user data
	assert.Zero(resp.FactrakSurveys[0].UserID)
	assert.Nil(resp.FactrakSurveys[0].User)

	/* Get test student 1 (expect failure) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/professors/%d", s1.ID), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusNotFound, w.Code)

	/* Get test prof 2 (expect failure) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/professors/%d", p2.ID), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusNotFound, w.Code)
}

func TestController_GetProfessorWithCourse(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

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

	/* Get test prof 1 (expect success) */
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/professors/%d?courseID=%d", p1.ID, c1.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	resp := models.User{}
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct prof
	assert.Equal(p1.ID, resp.ID)
	assert.Len(resp.FactrakSurveys, 2)
	// Assert that we don't include ProfessorFactrakSurveys
	assert.Nil(resp.ProfessorFactrakSurveys)

	// Check if we got surveys (only courses) (in reverse order)
	assert.Equal(resp.FactrakSurveys[0].Comment, fs4.Comment)
	assert.Equal(resp.FactrakSurveys[1].Comment, fs1.Comment)

	// Check if we removed sensitive user data
	assert.Zero(resp.FactrakSurveys[0].UserID)
	assert.Nil(resp.FactrakSurveys[0].User)

	/* Get prof 1 with a random course (expect empty) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/professors/%d?courseID=%d", p1.ID, 42), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusOK, w.Code)

	respData = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	resp = models.User{}
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct prof
	assert.Equal(p1.ID, resp.ID)
	assert.Len(resp.FactrakSurveys, 0)
}

func TestController_ListProfessorSurveys(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

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

	fs1 := models.FactrakSurvey{
		User:      &s1,
		Professor: &p1,
		Comment:   "Survey 1",
	}
	fs2 := models.FactrakSurvey{
		User:      &s2,
		Professor: &p1,
		Comment:   "Survey 2",
	}
	fs3 := models.FactrakSurvey{
		User:      &s1,
		Professor: &p2,
		Comment:   "Survey 3",
	}

	assert.NoError(db.Create(&fs1).Create(&fs2).Create(&fs3).Error)

	/* Get test prof 1 (expect success) */
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/professors/%d/surveys", p1.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp []models.FactrakSurvey
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if is survey 1 and 2
	assert.Len(resp, 2)

	// It should be in order of created first to created last
	assert.Equal(fs1.Comment, resp[1].Comment)
	assert.Equal(fs2.Comment, resp[0].Comment)

	// Assert that userID is not returned
	assert.Zero(resp[0].UserID)
	assert.Nil(resp[0].User)

	/* Get test student 1 (expect failure) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/professors/%d/surveys", s1.ID), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusNotFound, w.Code)

	/* Get test prof 2 (expect success) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/professors/%d/surveys", p2.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	resp = []models.FactrakSurvey{}
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if is survey 3
	assert.Len(resp, 1)
	assert.Equal(fs3.Comment, resp[0].Comment)
}

func TestController_ListProfessorSurveysWithCourse(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

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

	/* Get test prof 1 (expect success) */
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/professors/%d/surveys?courseID=%d", p1.ID, c1.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	resp := []models.FactrakSurvey{}
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct length
	assert.Len(resp, 2)

	// Check if we got surveys (only courses) (in reverse order)
	assert.Equal(resp[0].Comment, fs4.Comment)
	assert.Equal(resp[1].Comment, fs1.Comment)

	// Check if we removed sensitive user data
	assert.Zero(resp[0].UserID)
	assert.Nil(resp[0].User)

	/* Get prof 1 with a random course (expect empty) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/professors/%d/surveys?courseID=%d", p1.ID, 42), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusOK, w.Code)

	respData = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	resp = []models.FactrakSurvey{}
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct prof
	assert.Len(resp, 0)
}

func TestController_ListProfessorCourses(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

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

	/* Get test prof 1 (expect success) */
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/professors/%d/courses", p1.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp []models.Course
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if is survey 1 and 2
	assert.Len(resp, 2)

	// It should be in order of created first to created last
	assert.Equal(c1.Number, resp[0].Number)
	assert.Equal(c3.Number, resp[1].Number)

	/* Get test student 1 (expect failure) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/professors/%d/courses", s1.ID), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusNotFound, w.Code)

	/* Get test prof 2 (expect success) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/professors/%d/courses", p2.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	resp = []models.Course{}
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if is survey 3
	assert.Len(resp, 2)
	assert.Equal(c2.Number, resp[0].Number)
	assert.Equal(c3.Number, resp[1].Number)
}

func TestController_GetProfessorRatings(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

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
	p2 := models.User{
		Type:   models.UserTypeProfessor,
		Name:   "Professor 2",
		UnixID: "p2",
	}
	assert.NoError(db.Create(&p1).Create(&s1).Create(&p2).Error)

	fs1 := models.FactrakSurvey{
		User:                 &s1,
		Professor:            &p1,
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
		Professor:            &p2,
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

	/* Get test prof 1 (expect success) */
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/professors/%d/ratings", p1.ID), nil)
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

func TestController_GetProfessorRatingsWithCourse(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

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

	/* Get test prof 1 (expect success) */
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

	/* Get prof 1 with a random course (expect empty) */
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
