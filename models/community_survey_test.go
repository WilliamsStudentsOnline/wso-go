package models_test

import (
	"fmt"
	"testing"

	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	. "github.com/WilliamsStudentsOnline/wso-go/models"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestCommunitySurveyModel_GetAllCommunitySurvey(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewCommunitySurveyModel(db, zaptest.NewLogger(t).Sugar())

	communitySurveys := []CommunitySurvey{
		{

			Question: "what's your name?",
			User: &User{
				Type:   UserTypeStudent,
				UnixID: "s1",
				Name:   "Student 1",
			},
			NumOptions:    3,
			Options:       "john;johny;johnie",
			ResponseCount: 10,
		},
		{
			Question: "what year are you?",

			User: &User{
				Type:   UserTypeProfessor,
				UnixID: "p1",
				Name:   "Professor 1",
			},
			NumOptions:    4,
			Options:       "freshman;sophomore;junior;senior",
			ResponseCount: 15,
		},
	}

	for i := range communitySurveys {
		assert.NoError(db.Create(&communitySurveys[i]).Error)
	}

	var res []*CommunitySurvey
	preloads := []string{"user"}
	assert.NoError(m.GetAllCommunitySurveys(&res, &GetAllCommunitySurveyOptions{Preload: preloads}))
	fmt.Println("res contents:", res)
	n := len(communitySurveys)
	for i := range communitySurveys {
		assert.Equal(communitySurveys[i].Question, res[n-i-1].Question)
		assert.Equal(communitySurveys[i].User.UnixID, res[n-i-1].User.UnixID)
		assert.Equal(communitySurveys[i].User.Name, res[n-i-1].User.Name)
		assert.Equal(communitySurveys[i].ResponseCount, res[n-i-1].ResponseCount)
	}
}

func TestCommunitySurveyModel_CreateCommunitySurvey(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewCommunitySurveyModel(db, zaptest.NewLogger(t).Sugar())

	communitySurvey := CommunitySurvey{
		Question: "what's your name?",
		User: &User{
			Type:   UserTypeStudent,
			UnixID: "s1",
			Name:   "Student 1",
		},
		NumOptions:    3,
		Options:       "john;johny;johnie",
		ResponseCount: 10,
	}
	assert.NoError(m.CreateCommunitySurvey(&communitySurvey))

	var res CommunitySurvey
	assert.NoError(db.Find(&res).First(&res).Error)

	assert.Equal(communitySurvey.Question, res.Question)
	assert.Equal(communitySurvey.NumOptions, res.NumOptions)
	assert.Equal(communitySurvey.Options, res.Options)
	assert.Equal(communitySurvey.ResponseCount, res.ResponseCount)

}

func TestCommunitySurveyModel_GetCommunitySurveyByID(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewCommunitySurveyModel(db, zaptest.NewLogger(t).Sugar())

	survey := CommunitySurvey{
		Question: "what's your name?",
		User: &User{
			Type:   UserTypeStudent,
			UnixID: "s1",
			Name:   "Student 1",
		},
		NumOptions:    3,
		Options:       "john;johny;johnie",
		ResponseCount: 10,
	}

	assert.NoError(db.Create(&survey).Error)

	var res CommunitySurvey
	assert.NoError(m.GetCommunitySurveyByID(survey.ID, &res))

	assert.Equal(survey.ID, res.ID)
	assert.Equal(survey.NumOptions, res.NumOptions)
	assert.Equal(survey.ResponseCount, res.ResponseCount)

	//Does not preload
	assert.Nil(res.User)
}

func TestCommunitySurveyModel_UpdateCommunitySurvey(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewCommunitySurveyModel(db, zaptest.NewLogger(t).Sugar())

	survey := CommunitySurvey{
		Question: "what's your name?",
		User: &User{
			Type:   UserTypeStudent,
			UnixID: "s1",
			Name:   "Student 1",
		},
		NumOptions:    3,
		Options:       "john;johny;johnie",
		ResponseCount: 10,
	}

	assert.NoError(db.Create(&survey).Error)

	survey.NumOptions = 4
	survey.Options = "john;johny;johnie;jones"
	survey.ResponseCount = 0

	assert.NoError(m.UpdateCommunitySurvey(&survey))

	var res CommunitySurvey
	assert.NoError(db.First(&res).Error)

	assert.Equal(survey.ID, res.ID)
	assert.Equal(survey.NumOptions, res.NumOptions)
	assert.Equal(survey.ResponseCount, res.ResponseCount)
	assert.Equal(survey.Options, res.Options)

}

func TestCommunitySurveyModel_DeleteCommunitySurvey(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	m := NewCommunitySurveyModel(db, zaptest.NewLogger(t).Sugar())

	survey := CommunitySurvey{
		Question: "what's your name?",
		User: &User{
			Type:   UserTypeStudent,
			UnixID: "s1",
			Name:   "Student 1",
		},
		NumOptions:    3,
		Options:       "john;johny;johnie",
		ResponseCount: 10,
	}

	assert.NoError(db.Create(&survey).Error)

	assert.NoError(m.DeleteCommunitySurvey(&survey))

	var count int
	assert.NoError(db.Model(&CommunitySurveyResponse{}).Where("survey_id = ?", survey.ID).Count(&count).Error)
	assert.Equal(0, count)
}
