package factrak_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/factrak"
	testify "github.com/stretchr/testify/assert"
)

func TestController_GetAgreement(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	// Populate the database
	s1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "Student 1",
		UnixID: "s1",
	}
	s2 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "Student 2",
		UnixID: "s2",
	}
	survey := models.FactrakSurvey{
		User: &s1,
		Professor: &models.User{
			Type:   models.UserTypeProfessor,
			Name:   "Professor 1",
			UnixID: "p1",
		},
		Course: &models.Course{
			Number: "c1",
			AreaOfStudy: &models.AreaOfStudy{
				Name:         "Computer Science",
				Abbreviation: "CSCI",
				Department: &models.Department{
					Name: "Computer Science",
				},
			},
		},
		Comment: generateSurveyTestComment(),
	}
	agreement := models.FactrakAgreement{
		Agrees:        true,
		FactrakSurvey: &survey,
		User:          &s1,
	}
	assert.NoError(db.Create(&s1).Create(&s2).Create(&survey).Create(&agreement).Error)

	// Setup router
	router := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf)
	utils.AddUserContexts(router, s1.ID)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg)

	// First, we run tests on validations

	// Test 1: error on bad survey
	apiErr := lib.ErrorRecordNotFound
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/surveys/%d/agreement", 42), nil)
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	// Assert correct error
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	// Test 2: error on missing agreement (via other user)
	apiErr = lib.ErrorSurveyAgreementNotFound
	r1 := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf)
	utils.AddUserContexts(r1, s2.ID)
	SetupRouter(r1, db, cfg)
	w, err = utils.DoHTTPReq(r1, http.MethodGet, fmt.Sprintf("/surveys/%d/agreement", survey.ID), nil)
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	// Assert correct error
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	// Test 3: actually get agreement
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/surveys/%d/agreement", survey.ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)

	// Check to make sure we got the right thing
	var resp models.FactrakAgreement
	assert.NoError(json.Unmarshal(respData.Data, &resp))
	// Assertions
	assert.Equal(agreement.ID, resp.ID)
	assert.Equal(agreement.FactrakSurveyID, resp.FactrakSurveyID)
	assert.Equal(agreement.UserID, resp.UserID)
	assert.Equal(true, resp.Agrees)
}

func TestController_CreateAgreement(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	// Populate the database
	s1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "Student 1",
		UnixID: "s1",
	}
	s2 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "Student 2",
		UnixID: "s2",
	}
	s3 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "Student 3",
		UnixID: "s3",
	}
	survey := models.FactrakSurvey{
		User: &s1,
		Professor: &models.User{
			Type:   models.UserTypeProfessor,
			Name:   "Professor 1",
			UnixID: "p1",
		},
		Course: &models.Course{
			Number: "c1",
			AreaOfStudy: &models.AreaOfStudy{
				Name:         "Computer Science",
				Abbreviation: "CSCI",
				Department: &models.Department{
					Name: "Computer Science",
				},
			},
		},
		Comment: generateSurveyTestComment(),
	}
	a1 := models.FactrakAgreement{
		Agrees:        true,
		FactrakSurvey: &survey,
		User:          &s3,
	}
	assert.NoError(db.Create(&s1).Create(&s2).Create(&s3).Create(&survey).Create(&a1).Error)

	// Setup router
	router := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf)
	utils.AddUserContexts(router, s2.ID)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg)

	// First, we run tests on validations

	// Test 1: error on missing data
	apiErr := lib.ErrorMalformedRequestData
	w, err := utils.DoHTTPReq(router, http.MethodPost, fmt.Sprintf("/surveys/%d/agreement", survey.ID),
		bytes.NewBufferString(`{}`))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	// Assert correct error
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	// Test 2: error on bad survey
	apiErr = lib.ErrorRecordNotFound
	params := AgreementCreateParams{Agree: lib.BoolToPtr(false)}
	paramsData, err := json.Marshal(params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(router, http.MethodPost, fmt.Sprintf("/surveys/%d/agreement", 42), bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	// Assert correct error
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	// Test 3: error on existing agreement (via other user)
	apiErr = lib.ErrorSurveyAgreementAlreadyExists
	r1 := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf)
	utils.AddUserContexts(r1, s3.ID)
	SetupRouter(r1, db, cfg)
	w, err = utils.DoHTTPReq(r1, http.MethodPost, fmt.Sprintf("/surveys/%d/agreement", survey.ID), bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	// Assert correct error
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	// Test 3.5: error on self survey
	apiErr = lib.ErrorSurveyAgreementNoSelf
	r2 := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf)
	utils.AddUserContexts(r2, s1.ID)
	SetupRouter(r2, db, cfg)
	w, err = utils.DoHTTPReq(r2, http.MethodPost, fmt.Sprintf("/surveys/%d/agreement", survey.ID), bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	// Assert correct error
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	// Test 4: actually create agreement
	w, err = utils.DoHTTPReq(router, http.MethodPost, fmt.Sprintf("/surveys/%d/agreement", survey.ID), bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(http.StatusCreated, w.Code)
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)

	// Check to make sure we got the right thing
	var resp models.FactrakAgreement
	assert.NoError(json.Unmarshal(respData.Data, &resp))
	// Assertions
	assert.Equal(survey.ID, resp.FactrakSurveyID)
	assert.Equal(s2.ID, resp.UserID)
	assert.Equal(false, resp.Agrees)

	// Make sure it's in the db
	var agrDB models.FactrakAgreement
	assert.NoError(db.First(&agrDB, resp.ID).Error)
	assert.Equal(survey.ID, agrDB.FactrakSurveyID)
	assert.Equal(s2.ID, agrDB.UserID)
	assert.Equal(false, agrDB.Agrees)
}

func TestController_UpdateAgreement(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	// Populate the database
	s1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "Student 1",
		UnixID: "s1",
	}
	s2 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "Student 2",
		UnixID: "s2",
	}
	survey := models.FactrakSurvey{
		User: &s1,
		Professor: &models.User{
			Type:   models.UserTypeProfessor,
			Name:   "Professor 1",
			UnixID: "p1",
		},
		Course: &models.Course{
			Number: "c1",
			AreaOfStudy: &models.AreaOfStudy{
				Name:         "Computer Science",
				Abbreviation: "CSCI",
				Department: &models.Department{
					Name: "Computer Science",
				},
			},
		},
		Comment: generateSurveyTestComment(),
	}
	agreement := models.FactrakAgreement{
		Agrees:        true,
		FactrakSurvey: &survey,
		User:          &s1,
	}
	assert.NoError(db.Create(&s1).Create(&s2).Create(&survey).Create(&agreement).Error)

	// Setup router
	router := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf)
	utils.AddUserContexts(router, s1.ID)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg)

	// First, we run tests on validations

	// Test 1: error on missing data
	apiErr := lib.ErrorMalformedRequestData
	w, err := utils.DoHTTPReq(router, http.MethodPatch, fmt.Sprintf("/surveys/%d/agreement", survey.ID),
		bytes.NewBufferString(`{}`))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	// Assert correct error
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	// Test 2: error on bad survey
	apiErr = lib.ErrorRecordNotFound

	params := AgreementUpdateParams{Agree: lib.BoolToPtr(false)}
	paramsData, err := json.Marshal(params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(router, http.MethodPatch, fmt.Sprintf("/surveys/%d/agreement", 42), bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	// Assert correct error
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	// Test 3: error on non-existent agreement (via other user)
	apiErr = lib.ErrorSurveyAgreementNotFound
	r1 := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf)
	utils.AddUserContexts(r1, s2.ID)
	SetupRouter(r1, db, cfg)
	w, err = utils.DoHTTPReq(r1, http.MethodPatch, fmt.Sprintf("/surveys/%d/agreement", survey.ID), bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	// Assert correct error
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	// Test 4: actually update agreement
	w, err = utils.DoHTTPReq(router, http.MethodPatch, fmt.Sprintf("/surveys/%d/agreement", survey.ID), bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)

	// Check to make sure we got the right thing
	var resp models.FactrakAgreement
	assert.NoError(json.Unmarshal(respData.Data, &resp))
	// Assertions
	assert.Equal(survey.ID, resp.FactrakSurveyID)
	assert.Equal(s1.ID, resp.UserID)
	assert.Equal(false, resp.Agrees)

	// Make sure it's in the db
	var agrDB models.FactrakAgreement
	assert.NoError(db.First(&agrDB, resp.ID).Error)
	assert.Equal(survey.ID, agrDB.FactrakSurveyID)
	assert.Equal(s1.ID, agrDB.UserID)
	assert.Equal(false, agrDB.Agrees)
}

func TestController_DeleteAgreement(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)

	// Populate the database
	s1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "Student 1",
		UnixID: "s1",
	}
	s2 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "Student 2",
		UnixID: "s2",
	}
	survey := models.FactrakSurvey{
		User: &s1,
		Professor: &models.User{
			Type:   models.UserTypeProfessor,
			Name:   "Professor 1",
			UnixID: "p1",
		},
		Course: &models.Course{
			Number: "c1",
			AreaOfStudy: &models.AreaOfStudy{
				Name:         "Computer Science",
				Abbreviation: "CSCI",
				Department: &models.Department{
					Name: "Computer Science",
				},
			},
		},
		Comment: generateSurveyTestComment(),
	}
	agreement := models.FactrakAgreement{
		Agrees:        true,
		FactrakSurvey: &survey,
		User:          &s1,
	}
	assert.NoError(db.Create(&s1).Create(&s2).Create(&survey).Create(&agreement).Error)

	// Setup router
	router := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf)
	utils.AddUserContexts(router, s1.ID)
	cfg := utils.SetupConfig()
	SetupRouter(router, db, cfg)

	// First, we run tests on validations

	// Test 1: error on bad survey
	apiErr := lib.ErrorRecordNotFound
	w, err := utils.DoHTTPReq(router, http.MethodDelete, fmt.Sprintf("/surveys/%d/agreement", 42), nil)
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	// Assert correct error
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	// Test 3: error on non-existent agreement (via other user)
	apiErr = lib.ErrorSurveyAgreementNotFound
	r1 := utils.SetupRouter(auth.ScopeFactrakFull, auth.ScopeWriteSelf)
	utils.AddUserContexts(r1, s2.ID)
	SetupRouter(r1, db, cfg)
	w, err = utils.DoHTTPReq(r1, http.MethodDelete, fmt.Sprintf("/surveys/%d/agreement", survey.ID), nil)
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	// Assert correct error
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	// Test 4: actually delete agreement
	w, err = utils.DoHTTPReq(router, http.MethodDelete, fmt.Sprintf("/surveys/%d/agreement", survey.ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)

	// Check to make sure we got the right thing
	var resp models.FactrakAgreement
	assert.NoError(json.Unmarshal(respData.Data, &resp))
	// Assertions
	assert.Equal(survey.ID, resp.FactrakSurveyID)
	assert.Equal(s1.ID, resp.UserID)
	assert.Equal(true, resp.Agrees)

	// Make sure it's not in the db
	var count int
	assert.NoError(db.Unscoped().Table("factrak_agreements").Where("id = ?", resp.ID).Count(&count).Error)
	assert.Zero(count)
}
