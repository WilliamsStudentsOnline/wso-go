package models_test

import (
	"fmt"
	"testing"

	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	. "github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestCommunitySurveyResponseModel_GetAllCommunitySurveyResponse(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewCommunitySurveyResponseModel(db, zaptest.NewLogger(t).Sugar())

	user1 := User{
		Type:   UserTypeStudent,
		UnixID: "s1",
		Name:   "Student 1",
	}

	user2 := User{
		Type:   UserTypeStudent,
		UnixID: "s2",
		Name:   "Student 2",
	}

	communitySurveyResponses := []CommunitySurveyResponse{
		{
			User: &user1,
			Survey: &CommunitySurvey{
				Question: "what's your name?",
				User: &User{
					Type:   UserTypeStudent,
					UnixID: "s3",
					Name:   "Student 3",
				},
			},
			Response: "madaliday",
		},
		{
			User: &user2,
			Survey: &CommunitySurvey{
				Question: "where do you live?",
				User: &User{
					Type:   UserTypeStudent,
					UnixID: "s4",
					Name:   "Student 4",
				},
			},
			Response: "mission",
		},
	}

	for i := range communitySurveyResponses {
		assert.NoError(db.Create(&communitySurveyResponses[i]).Error)
	}

	var res []*CommunitySurveyResponse
	preloads := []string{"user"}
	assert.NoError(m.GetAllCommunitySurveyResponses(&res, &GetAllCommunitySurveyResponseOptions{Preload: preloads}))

	n := len(res)
	fmt.Println(n)
	for i := range communitySurveyResponses {
		assert.Equal(communitySurveyResponses[i].ID, res[n-i-1].ID)

		assert.Equal(communitySurveyResponses[i].User.UnixID, res[n-i-1].User.UnixID)
		assert.Equal(communitySurveyResponses[i].User.Name, res[n-i-1].User.Name)

		assert.Equal(communitySurveyResponses[i].Response, res[n-i-1].Response)
		if res[n-i-1].Survey != nil && res[n-i-1].Survey.User != nil {
			assert.Equal(communitySurveyResponses[i].Survey.User.Name, res[n-i-1].Survey.User.Name)
		}
	}

}

func TestCommunitySurveyResponseModel_CreateCommunitySurveyResponse(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewCommunitySurveyResponseModel(db, zaptest.NewLogger(t).Sugar())

	surveyResponse := CommunitySurveyResponse{
		User: &User{
			Name: "Student 1",
		},
		Survey: &CommunitySurvey{
			Question: "what's your name?",
			User: &User{
				Type:   UserTypeStudent,
				UnixID: "s3",
				Name:   "Student 3",
			},
		},
		Response: "madaliday",
	}

	assert.NoError(m.CreateCommunitySurveyResponse(&surveyResponse))

	var res CommunitySurveyResponse
	assert.NoError(db.Find(&res).First(&res).Error)
	assert.Equal(surveyResponse.ID, res.ID)
	assert.Equal(surveyResponse.Response, res.Response)

}

func TestCommunitySurveyResponseModel_GetCommunitySurveyResponseByID(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewCommunitySurveyResponseModel(db, zaptest.NewLogger(t).Sugar())

	user1 := User{
		Type:   UserTypeStudent,
		UnixID: "s1",
		Name:   "Student 1",
	}

	surveyResponse := CommunitySurveyResponse{
		User: &user1,
		Survey: &CommunitySurvey{
			Question: "what's your name?",
			User: &User{
				Type:   UserTypeStudent,
				UnixID: "s3",
				Name:   "Student 3",
			},
		},
		Response: "madaliday",
	}

	assert.NoError(db.Create(&surveyResponse).Error)

	var res CommunitySurveyResponse
	assert.NoError(m.GetCommunitySurveyResponseByID(surveyResponse.ID, &res))

	assert.Equal(surveyResponse.ID, res.ID)
	assert.Equal(surveyResponse.Response, res.Response)

	//Does not preload
	assert.Nil(res.Survey)
}

func TestCommunitySurveyResponseModel_DeleteCommunitySurveyResponse(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewCommunitySurveyResponseModel(db, zaptest.NewLogger(t).Sugar())

	user1 := User{
		Type:   UserTypeStudent,
		UnixID: "s1",
		Name:   "Student 1",
	}

	surveyResponse := CommunitySurveyResponse{
		User:     &user1,
		SurveyID: 1234,
		Survey: &CommunitySurvey{
			Question: "what's your name?",
			User: &User{
				Type:   UserTypeStudent,
				UnixID: "s3",
				Name:   "Student 3",
			},
		},
		Response: "madaliday",
	}

	assert.NoError(db.Create(&surveyResponse).Error)

	assert.NoError(m.DeleteCommunitySurveyResponse(&surveyResponse))

	var count int
	assert.NoError(db.Model(&CommunitySurveyResponse{}).Where("survey_id = ?", surveyResponse.SurveyID).Count(&count).Error)
	assert.Equal(0, count)
}
