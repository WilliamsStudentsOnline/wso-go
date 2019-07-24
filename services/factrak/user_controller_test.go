package factrak_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/gin-gonic/gin"
	testify "github.com/stretchr/testify/assert"

	. "github.com/WilliamsStudentsOnline/wso-go/services/factrak"
)

func TestController_ListUserSurveys(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := gin.Default()
	SetupRouter(router, db)

	// Insert test user into db
	p1 := models.User{
		Type:       models.UserTypeProfessor,
		Name:       "Professor 1",
		UnixID:     "p1",
		Visible:    true,
		AtWilliams: true,
	}
	s1 := models.User{
		Type:       models.UserTypeStudent,
		Name:       "Student",
		UnixID:     "s1",
		Visible:    true,
		AtWilliams: true,
	}
	s2 := models.User{
		Type:       models.UserTypeStudent,
		Name:       "Student",
		UnixID:     "s2",
		Visible:    true,
		AtWilliams: true,
	}
	p2 := models.User{
		Type:       models.UserTypeProfessor,
		Name:       "Professor 2",
		UnixID:     "p2",
		Visible:    true,
		AtWilliams: true,
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

	/* Get test student 1 (expect success) */
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/users/%d/surveys", s1.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp []models.FactrakSurvey
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if is survey 1 and 3
	assert.Len(resp, 2)

	// It should be in order of created first to created last
	assert.Equal(fs1.Comment, resp[1].Comment)
	assert.Equal(fs3.Comment, resp[0].Comment)

	// Assert that userID is not returned
	assert.Zero(resp[0].UserID)
	assert.Nil(resp[0].User)

	/* Get test prof 1 (expect empty success) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/users/%d/surveys", p1.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	resp = []models.FactrakSurvey{}
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if is survey 1 and 3
	assert.Len(resp, 0)

	/* Get random fake user (expect failure) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/users/%d/surveys", 42), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusNotFound, w.Code)

	/* Get test student 2 (expect success) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/users/%d/surveys", s2.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	resp = []models.FactrakSurvey{}
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if is survey 2
	assert.Len(resp, 1)
	assert.Equal(fs2.Comment, resp[0].Comment)
}
