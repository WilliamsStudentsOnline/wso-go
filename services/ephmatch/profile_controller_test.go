package ephmatch_test

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
	. "github.com/WilliamsStudentsOnline/wso-go/services/ephmatch"
	"github.com/gin-gonic/gin"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestController_GetSelfProfile(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	cfg := utils.SetupConfig()

	srYear := (&models.StudentModel{}).SeniorYear()

	s := []*models.User{
		// Case: active profile
		{

			EphmatchProfile: &models.EphmatchProfile{
				Gender:      lib.StrToPtr("she/her/hers"),
				Description: lib.StrToPtr("test123"),
			},
		},
		// Case: deleted profile
		{
			EphmatchProfile: &models.EphmatchProfile{
				Gender:      lib.StrToPtr("they/them/theirs"),
				Description: lib.StrToPtr("hello world"),
			},
		},
		// Case: not student
		{
			Type:   models.UserTypeProfessor,
			Name:   "Professor 1",
			UnixID: "p1",
			EphmatchProfile: &models.EphmatchProfile{
				Gender:      lib.StrToPtr("bteevev"),
				Description: lib.StrToPtr("42"),
			},
		},
		// Case: not visible
		{
			Visible: lib.BoolToPtr(false),
			EphmatchProfile: &models.EphmatchProfile{
				Gender:      lib.StrToPtr("1"),
				Description: lib.StrToPtr("2"),
			},
		},
		// Case: no profile
		{},
	}
	for i, val := range s {
		if val.Name == "" {
			val.Name = fmt.Sprintf("Student %d", i)
		}
		if val.UnixID == "" {
			val.UnixID = fmt.Sprintf("s%d", i)
		}
		if val.Type == "" {
			val.Type = models.UserTypeStudent
		}
		if val.ClassYear == nil {
			val.ClassYear = &srYear
		}
		assert.NoError(db.Create(val).Error)
	}

	// Delete the deleted profile
	assert.NoError(db.Delete(s[1].EphmatchProfile).Error)

	testCases := []struct {
		name     string
		expected *models.EphmatchProfile
		user     *models.User
		err      *lib.APIError
	}{
		{
			"active profile",
			s[0].EphmatchProfile,
			s[0],
			nil,
		},
		{
			"deleted profile",
			s[1].EphmatchProfile,
			s[1],
			nil,
		}, {
			"not student",
			s[2].EphmatchProfile,
			s[2],
			nil,
		}, {
			"not visible",
			s[3].EphmatchProfile,
			s[3],
			nil,
		}, {
			"no profile",
			nil,
			s[4],
			lib.ErrorRecordNotFound,
		},
	}

	// Test for db query
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := testify.New(t)
			router := utils.SetupRouter(auth.ScopeEphmatch)
			utils.AddUserContexts(router, tc.user.ID)
			SetupRouter(router, db, cfg, zap.S())

			w, err := utils.DoHTTPReq(router, http.MethodGet, "/profile", nil)
			a.NoError(err)

			resp := utils.GetHTTPDataResp(a, w.Body.Bytes())

			if tc.err != nil {
				a.Equal(tc.err.HTTPCode, w.Code)
				a.NotNil(resp.Error)
				a.Equal(tc.err.Code, resp.Error.ErrorCode)
			} else {
				a.Equal(http.StatusOK, w.Code)
				a.Nil(resp.Error)

				var res models.EphmatchProfile
				err = json.Unmarshal(resp.Data, &res)
				a.NoError(err)

				a.Equal(tc.expected.Description, res.Description)
				a.Equal(tc.expected.Gender, res.Gender)
				a.Equal(tc.user.ID, res.UserID)
			}
		})
	}
}

func TestController_CreateProfile(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	cfg := utils.SetupConfig()

	srYear := (&models.StudentModel{}).SeniorYear()

	s := []*models.User{
		// Case: no profile
		{},
		// Case: active profile (overwrite)
		{

			EphmatchProfile: &models.EphmatchProfile{
				Gender:      lib.StrToPtr("she/her/hers"),
				Description: lib.StrToPtr("test123"),
			},
		},
		// Case: deleted profile
		{
			EphmatchProfile: &models.EphmatchProfile{
				Gender:      lib.StrToPtr("they/them/theirs"),
				Description: lib.StrToPtr("hello world"),
			},
		},
	}
	var routers []*gin.Engine
	for i, val := range s {
		val.Name = fmt.Sprintf("Student %d", i)
		val.UnixID = fmt.Sprintf("s%d", i)
		val.Type = models.UserTypeStudent
		val.ClassYear = &srYear
		assert.NoError(db.Create(val).Error)

		r := utils.SetupRouter(auth.ScopeEphmatch)
		utils.AddUserContexts(r, val.ID)
		SetupRouter(r, db, cfg, zap.S())
		routers = append(routers, r)
	}

	// Delete the deleted profile
	assert.NoError(db.Delete(s[2].EphmatchProfile).Error)

	/* Create ephmatch with bad gender (expect failure) */
	apiErr := lib.ErrorEphmatchGenderUnknown
	params := ProfileCreateParams{Description: lib.StrToPtr("foobar"), Gender: lib.StrToPtr("custom gender")}
	paramsData, err := json.Marshal(&params)
	assert.NoError(err)
	w, err := utils.DoHTTPReq(routers[0], http.MethodPost, "/profile", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Create ephmatch with good custom gender (expect success on user 1) */
	params = ProfileCreateParams{Description: lib.StrToPtr("foobar"), Gender: lib.StrToPtr("custom gender"), OtherGender: true}
	paramsData, err = json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(routers[0], http.MethodPost, "/profile", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(http.StatusCreated, w.Code)
	// Get from DB
	resDB := models.EphmatchProfile{}
	assert.NoError(db.Where(models.EphmatchProfile{UserID: s[0].ID}).First(&resDB).Error)
	assert.Equal(*params.Description, *resDB.Description)
	assert.Equal(*params.Gender, *resDB.Gender)

	/* Create profile where it already exists (expect success on user 2) */
	params = ProfileCreateParams{Description: lib.StrToPtr("description here 123"), Gender: lib.StrToPtr("he/him/his")}
	paramsData, err = json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(routers[1], http.MethodPost, "/profile", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(http.StatusCreated, w.Code)
	// Get from DB
	resDB = models.EphmatchProfile{}
	assert.NoError(db.Where(models.EphmatchProfile{UserID: s[1].ID}).First(&resDB).Error)
	assert.Equal(*params.Description, *resDB.Description)
	assert.Equal(*params.Gender, *resDB.Gender)

	/* Create profile where it was deleted (expect success on user 3) */
	params = ProfileCreateParams{Description: lib.StrToPtr("abc 123 hello world"), Gender: lib.StrToPtr("she/her/hers")}
	paramsData, err = json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(routers[2], http.MethodPost, "/profile", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(http.StatusCreated, w.Code)
	// Get from DB
	resDB = models.EphmatchProfile{}
	assert.NoError(db.Where(models.EphmatchProfile{UserID: s[2].ID}).First(&resDB).Error)
	assert.Equal(*params.Description, *resDB.Description)
	assert.Equal(*params.Gender, *resDB.Gender)
}

func TestController_CreateProfile_DeleteUndelete(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	cfg := utils.SetupConfig()

	srYear := (&models.StudentModel{}).SeniorYear()

	u := &models.User{
		Name:      "Student 1",
		UnixID:    "s1",
		Type:      models.UserTypeStudent,
		ClassYear: &srYear,
		EphmatchProfile: &models.EphmatchProfile{
			Gender:      lib.StrToPtr("she/her/hers"),
			Description: lib.StrToPtr("test123"),
		},
	}
	assert.NoError(db.Create(&u).Error)

	r := utils.SetupRouter(auth.ScopeEphmatch)
	utils.AddUserContexts(r, u.ID)
	SetupRouter(r, db, cfg, zap.S())

	/* Delete profile */
	w, err := utils.DoHTTPReq(r, http.MethodDelete, "/profile", nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	/* Create profile where it was deleted (expect success) */
	params := ProfileCreateParams{Description: lib.StrToPtr("test123"), Gender: lib.StrToPtr("she/her/hers")}
	paramsData, err := json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(r, http.MethodPost, "/profile", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(http.StatusCreated, w.Code)

	// Get from DB
	resDB := models.EphmatchProfile{}
	assert.NoError(db.Where(models.EphmatchProfile{UserID: u.ID}).First(&resDB).Error)
	assert.Equal(*params.Description, *resDB.Description)
	assert.Equal(*params.Gender, *resDB.Gender)
	assert.Nil(resDB.DeletedAt)

	var count int
	assert.NoError(db.Model(models.EphmatchProfile{}).Where(models.EphmatchProfile{UserID: u.ID}).Count(&count).Error)
	assert.Equal(1, count)

	w, err = utils.DoHTTPReq(r, http.MethodGet, "/profile", nil)
	assert.NoError(err)
	resp := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Equal(http.StatusOK, w.Code)
	assert.Nil(resp.Error)
	var res models.EphmatchProfile
	err = json.Unmarshal(resp.Data, &res)
	assert.NoError(err)
	assert.Equal(*params.Description, *res.Description)
	assert.Equal(*params.Gender, *res.Gender)
	assert.False(res.Deleted)
}

func TestController_UpdateProfile(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	cfg := utils.SetupConfig()

	srYear := (&models.StudentModel{}).SeniorYear()

	s := []*models.User{
		// Case: active profile (overwrite)
		{

			EphmatchProfile: &models.EphmatchProfile{
				Gender:      lib.StrToPtr("she/her/hers"),
				Description: lib.StrToPtr("test123"),
			},
		},
		// Case: deleted profile
		{
			EphmatchProfile: &models.EphmatchProfile{
				Gender:      lib.StrToPtr("they/them/theirs"),
				Description: lib.StrToPtr("hello world"),
			},
		},
		// Case: no profile
		{},
	}
	var routers []*gin.Engine
	for i, val := range s {
		val.Name = fmt.Sprintf("Student %d", i)
		val.UnixID = fmt.Sprintf("s%d", i)
		val.Type = models.UserTypeStudent
		val.ClassYear = &srYear
		assert.NoError(db.Create(val).Error)

		r := utils.SetupRouter(auth.ScopeEphmatch)
		utils.AddUserContexts(r, val.ID)
		SetupRouter(r, db, cfg, zap.S())
		routers = append(routers, r)
	}

	// Delete the deleted profile
	assert.NoError(db.Delete(s[1].EphmatchProfile).Error)

	/* Update profile with bad gender (expect failure) */
	apiErr := lib.ErrorEphmatchGenderUnknown
	params := ProfileUpdateParams{Gender: lib.StrToPtr("custom gender")}
	paramsData, err := json.Marshal(&params)
	assert.NoError(err)
	w, err := utils.DoHTTPReq(routers[0], http.MethodPatch, "/profile", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Update profile with missing profile (expect failure) */
	apiErr = lib.ErrorRecordNotFound
	params = ProfileUpdateParams{Description: lib.StrToPtr("what's up")}
	paramsData, err = json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(routers[2], http.MethodPatch, "/profile", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Update profile with deleted profile (expect failure) */
	apiErr = lib.ErrorRecordNotFound
	params = ProfileUpdateParams{Description: lib.StrToPtr("foobar123 hello world 54231")}
	paramsData, err = json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(routers[1], http.MethodPatch, "/profile", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Update profile with good custom gender (expect success on user 1) */
	params = ProfileUpdateParams{Gender: lib.StrToPtr("abcdefg fewvc"), OtherGender: true}
	paramsData, err = json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(routers[0], http.MethodPatch, "/profile", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)
	// Get from DB
	resDB := models.EphmatchProfile{}
	assert.NoError(db.Where(models.EphmatchProfile{UserID: s[0].ID}).First(&resDB).Error)
	assert.Equal(*s[0].EphmatchProfile.Description, *resDB.Description)
	assert.Equal(*params.Gender, *resDB.Gender)
}

func TestController_UpdateProfile_Deleted(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	cfg := utils.SetupConfig()

	srYear := (&models.StudentModel{}).SeniorYear()

	u := &models.User{
		Name:      "Student 1",
		UnixID:    "s1",
		Type:      models.UserTypeStudent,
		ClassYear: &srYear,
		EphmatchProfile: &models.EphmatchProfile{
			Gender:      lib.StrToPtr("she/her/hers"),
			Description: lib.StrToPtr("test123"),
		},
	}
	assert.NoError(db.Create(&u).Error)

	r := utils.SetupRouter(auth.ScopeEphmatch)
	utils.AddUserContexts(r, u.ID)
	SetupRouter(r, db, cfg, zap.S())

	/* Delete profile */
	w, err := utils.DoHTTPReq(r, http.MethodDelete, "/profile", nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	/* Updated profile where it was deleted (expect failure) */
	apiErr := lib.ErrorRecordNotFound
	params := ProfileUpdateParams{Description: lib.StrToPtr("4gwe"), Gender: lib.StrToPtr(models.EphmatchProfileGenderNB)}
	paramsData, err := json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(r, http.MethodPatch, "/profile", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)
}

func TestController_UpdateProfile_Again(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	cfg := utils.SetupConfig()

	srYear := (&models.StudentModel{}).SeniorYear()

	u := &models.User{
		Name:      "Student 1",
		UnixID:    "s1",
		Type:      models.UserTypeStudent,
		ClassYear: &srYear,
		EphmatchProfile: &models.EphmatchProfile{
			Gender:      lib.StrToPtr("she/her/hers"),
			Description: lib.StrToPtr("test123"),
		},
	}
	assert.NoError(db.Create(&u).Error)

	r := utils.SetupRouter(auth.ScopeEphmatch)
	utils.AddUserContexts(r, u.ID)
	SetupRouter(r, db, cfg, zap.S())

	/* Update profile */
	params := ProfileUpdateParams{Description: lib.StrToPtr("4gwe"), Gender: lib.StrToPtr(models.EphmatchProfileGenderNB)}
	paramsData, err := json.Marshal(&params)
	assert.NoError(err)
	w, err := utils.DoHTTPReq(r, http.MethodPatch, "/profile", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	var count int
	assert.NoError(db.Model(models.EphmatchProfile{}).Where(models.EphmatchProfile{UserID: u.ID}).Count(&count).Error)
	assert.Equal(1, count)
}

func TestController_DeleteProfile(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	cfg := utils.SetupConfig()

	srYear := (&models.StudentModel{}).SeniorYear()

	s := []*models.User{
		// Case: active profile
		{

			EphmatchProfile: &models.EphmatchProfile{
				Gender:      lib.StrToPtr("she/her/hers"),
				Description: lib.StrToPtr("test123"),
			},
		},
		// Case: deleted profile
		{
			EphmatchProfile: &models.EphmatchProfile{
				Gender:      lib.StrToPtr("they/them/theirs"),
				Description: lib.StrToPtr("hello world"),
			},
		},
		// Case: no profile
		{},
	}
	var routers []*gin.Engine
	for i, val := range s {
		val.Name = fmt.Sprintf("Student %d", i)
		val.UnixID = fmt.Sprintf("s%d", i)
		val.Type = models.UserTypeStudent
		val.ClassYear = &srYear
		assert.NoError(db.Create(val).Error)

		r := utils.SetupRouter(auth.ScopeEphmatch)
		utils.AddUserContexts(r, val.ID)
		SetupRouter(r, db, cfg, zap.S())
		routers = append(routers, r)
	}

	// Delete the deleted profile
	assert.NoError(db.Delete(s[1].EphmatchProfile).Error)

	/* Update profile with good custom gender (expect success on user 2) */
	apiErr := lib.ErrorRecordNotFound
	w, err := utils.DoHTTPReq(routers[1], http.MethodDelete, "/profile", nil)
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Delete profile with missing profile (expect failure) */
	apiErr = lib.ErrorRecordNotFound
	assert.NoError(err)
	w, err = utils.DoHTTPReq(routers[2], http.MethodDelete, "/profile", nil)
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Update profile with good custom gender (expect success on user 1) */
	w, err = utils.DoHTTPReq(routers[0], http.MethodDelete, "/profile", nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)
	// Get from DB
	count := -1
	assert.NoError(db.Model(&models.EphmatchProfile{}).Where(models.EphmatchProfile{UserID: s[0].ID}).Count(&count).Error)
	assert.Equal(0, count)
}
