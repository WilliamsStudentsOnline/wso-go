package factrak_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/factrak"
	"github.com/gin-gonic/gin"
	testify "github.com/stretchr/testify/assert"
)

func TestController_ListSurveys(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := gin.Default()
	SetupRouter(router, db)

	c1 := models.Course{
		Number: "c1",
		AreaOfStudy: &models.AreaOfStudy{
			Name: "Computer Science",
			Abbreviation: "CSCI",
			Department: &models.Department{
				Name: "Computer Science",
			},
		},
	}

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
	assert.NoError(db.Create(&p1).Create(&s1).Error)

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

	// Get test surveys
	w, err := utils.DoHTTPReq(router, http.MethodGet, "/surveys", nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp []models.FactrakSurvey
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct surveys
	assert.Len(resp, 2)
	assert.Equal(fs1.Comment, resp[0].Comment)
	assert.Equal(fs2.Comment, resp[1].Comment)

	// Make sure anonymous
	assert.Zero(resp[0].UserID)
	assert.Nil(resp[0].User)
}

func TestController_GetSurvey(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := gin.Default()
	SetupRouter(router, db)

	c1 := models.Course{
		Number: "c1",
		AreaOfStudy: &models.AreaOfStudy{
			Name: "Computer Science",
			Abbreviation: "CSCI",
			Department: &models.Department{
				Name: "Computer Science",
			},
		},
	}

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
	assert.NoError(db.Create(&p1).Create(&s1).Error)

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

	// Get test survey 1
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/surveys/%d", fs1.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.FactrakSurvey
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct surveys
	assert.Equal(fs1.Comment, resp.Comment)

	// Make sure anonymous
	assert.Zero(resp.UserID)
	assert.Nil(resp.User)

	// Get bad survey (expect failure
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/surveys/%d", 42), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusNotFound, w.Code)
}