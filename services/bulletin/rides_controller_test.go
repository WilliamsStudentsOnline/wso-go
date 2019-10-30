package bulletin_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	. "github.com/WilliamsStudentsOnline/wso-go/services/bulletin"
	testify "github.com/stretchr/testify/assert"
)

func TestController_ListRides(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeBulletin, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()

	rides := []*models.BulletinRide{
		// 0
		{},
		// 1
		{},
		// 2
		{
			Offer: lib.BoolToPtr(true),
		},
		// 3
		{
			Offer: lib.BoolToPtr(true),
		},
		// 4: Ride was a week ago
		{
			Date: time.Now().Add(-1 * time.Hour * 24 * 7),
		},
		// 5: Ride was at the start of the day today
		{
			Date: time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 2, 0, time.Now().Location()),
		},
	}
	for i, val := range rides {
		val.Source = fmt.Sprintf("Source %d", i)
		val.Destination = fmt.Sprintf("Destination %d", i)
		val.Body = generateBulletinTestBody()
		if val.Offer == nil {
			val.Offer = lib.BoolToPtr(false)
		}
		if val.Date.IsZero() {
			val.Date = time.Now().Add(time.Hour * time.Duration(i))
		}
		assert.NoError(db.Create(val).Error)
	}

	testCases := []struct {
		name     string
		query    string
		expected []int
	}{
		{
			"default",
			"",
			[]int{3, 2, 1, 0, 5},
		},
		{
			"type request",
			"type=request",
			[]int{1, 0, 5},
		},
		{
			"type offer",
			"type=offer",
			[]int{3, 2},
		},
		{
			"limit",
			"limit=2",
			[]int{3, 2},
		},
		{
			"limit and start",
			"limit=2&start=" + rides[2].Date.Format(time.RFC3339),
			[]int{1, 0},
		},
		{
			"all rides",
			"all=true",
			[]int{3, 2, 1, 0, 5, 4},
		},
	}

	SetupRouter(router, db, cfg)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := testify.New(t)
			// Get test user
			w, err := utils.DoHTTPReq(router, http.MethodGet, "/rides?"+tc.query, nil)
			a.NoError(err)

			// Status is okay
			a.Equal(http.StatusOK, w.Code)

			// Decode response
			respData := utils.GetHTTPDataResp(a, w.Body.Bytes())
			a.Nil(respData.Error)
			var resp []*models.BulletinRide
			a.NoError(json.Unmarshal(respData.Data, &resp))

			// Check if correct result
			a.Len(resp, len(tc.expected))
			for i := range tc.expected {
				a.Equal(rides[tc.expected[i]].ID, resp[i].ID)
				a.Equal(rides[tc.expected[i]].Body, resp[i].Body)
			}
		})
	}
}

func TestController_GetRide(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeBulletin, auth.ScopeWriteSelf, auth.ScopeUsers)
	cfg := utils.SetupConfig()

	b1 := models.BulletinRide{
		Source:      "Source 1",
		Destination: "Destination 1",
		Body:        generateBulletinTestBody(),
		Offer:       lib.BoolToPtr(false),
		Date:        time.Now().Add(24 * time.Hour),
		User: &models.User{
			Type:   models.UserTypeStudent,
			UnixID: "u1",
			Name:   "User 1",
		},
	}

	assert.NoError(db.Create(&b1).Error)

	utils.AddUserContexts(router, b1.User.ID)
	SetupRouter(router, db, cfg)

	/* Get test ride (signed in) */
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/rides/%d", b1.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.BulletinRide
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct ride
	assert.Equal(b1.ID, resp.ID)
	assert.Equal(b1.Body, resp.Body)
	assert.False(*resp.Offer)
	// User should not be nil, as signed in
	assert.NotNil(resp.User)

	/* Get test ride (signed out) */
	r1 := utils.SetupRouter(auth.ScopeBulletin, auth.ScopeWriteSelf)
	SetupRouter(r1, db, cfg)
	w, err = utils.DoHTTPReq(r1, http.MethodGet, fmt.Sprintf("/rides/%d", b1.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	resp = models.BulletinRide{}
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct ride
	assert.Equal(b1.ID, resp.ID)
	// Ensure that user is missing
	assert.Nil(resp.User)

	/* Get test ride bad id (expect failure) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/rides/%d", 42), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusNotFound, w.Code)
}

func TestController_CreateRide(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeBulletin, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()

	u1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "User 1",
		UnixID: "u1",
	}
	assert.NoError(db.Create(&u1).Error)

	utils.AddUserContexts(router, u1.ID)
	SetupRouter(router, db, cfg)

	/* Create ride with missing data (expect failure) */
	apiErr := lib.ErrorRequestDataValidationFailed
	w, err := utils.DoHTTPReq(router, http.MethodPost, "/rides", bytes.NewBufferString(`{"source": "hi"}`))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Create ride with bad date (expect failure) */
	params := CreateRideParams{
		Offer:       lib.BoolToPtr(true),
		Source:      "Source 1",
		Destination: "Destination 1",
		Body:        generateBulletinTestBody(),
		Date:        time.Now().Add(-time.Hour * 24),
	}
	apiErr = lib.ErrorBulletinRideDateInPast
	paramsData, err := json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(router, http.MethodPost, "/rides", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Create ride (expect success) */
	params = CreateRideParams{
		Offer:       lib.BoolToPtr(false),
		Source:      "Source 1",
		Destination: "Destination 1",
		Body:        generateBulletinTestBody(),
		Date:        time.Now().Add(time.Hour),
	}
	paramsData, err = json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(router, http.MethodPost, "/rides", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(http.StatusCreated, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.BulletinRide
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Assert good response
	assert.NotZero(resp.ID)
	assert.Equal(params.Body, resp.Body)
	assert.Equal(params.Offer, resp.Offer)

	// Assert found in db
	var count int
	assert.NoError(db.Model(&models.BulletinRide{}).Where("id = ?", resp.ID).Count(&count).Error)
	assert.Equal(1, count)
}

func TestController_UpdateRide(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeBulletin, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()

	u1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "User 1",
		UnixID: "u1",
	}
	b1 := models.BulletinRide{
		Source:      "Source 1",
		Destination: "Destination 1",
		Body:        generateBulletinTestBody(),
		Offer:       lib.BoolToPtr(true),
		Date:        time.Now().Add(time.Hour),
		User:        &u1,
	}
	b2 := models.BulletinRide{
		Source:      "Source 2",
		Destination: "Destination 2",
		Body:        generateBulletinTestBody(),
		Offer:       lib.BoolToPtr(false),
		Date:        time.Now().Add(time.Hour),
		User: &models.User{
			Type:   models.UserTypeStudent,
			Name:   "User 2",
			UnixID: "u2",
		},
	}
	assert.NoError(db.Create(&u1).Create(&b1).Create(&b2).Error)

	utils.AddUserContexts(router, u1.ID)
	SetupRouter(router, db, cfg)

	/* Update ride with bad ride (expect failure) */
	params := UpdateRideParams{
		Body: lib.StrToPtr(generateBulletinTestBody()),
	}
	apiErr := lib.ErrorRecordNotFound
	paramsData, err := json.Marshal(&params)
	assert.NoError(err)
	w, err := utils.DoHTTPReq(router, http.MethodPatch, fmt.Sprintf("/rides/%d", 42), bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Update ride with bad owner (expect failure) */
	params = UpdateRideParams{
		Body: lib.StrToPtr(generateBulletinTestBody()),
	}
	apiErr = lib.ErrorMustBeSelf
	paramsData, err = json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(router, http.MethodPatch, fmt.Sprintf("/rides/%d", b2.ID), bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Update ride with bad dates (expect failure) */
	params = UpdateRideParams{
		Body: lib.StrToPtr(generateBulletinTestBody()),
		Date: lib.TimeToPtr(time.Now().Add(-time.Hour)),
	}
	apiErr = lib.ErrorBulletinRideDateInPast
	paramsData, err = json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(router, http.MethodPatch, fmt.Sprintf("/rides/%d", b1.ID), bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Update ride (expect success) */
	params = UpdateRideParams{
		Body: lib.StrToPtr(generateBulletinTestBody()),
	}
	paramsData, err = json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(router, http.MethodPatch, fmt.Sprintf("/rides/%d", b1.ID), bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.BulletinRide
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Assert good response
	assert.Equal(b1.ID, resp.ID)
	assert.Equal(*params.Body, resp.Body)

	// Assert updated in db
	var respDB models.BulletinRide
	assert.NoError(db.First(&respDB, b1.ID).Error)
	assert.Equal(*params.Body, respDB.Body)
}

func TestController_DeleteRide(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeBulletin, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()

	u1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "User 1",
		UnixID: "u1",
	}
	b1 := models.BulletinRide{
		Source:      "Source 1",
		Destination: "Destination 1",
		Body:        generateBulletinTestBody(),
		Offer:       lib.BoolToPtr(true),
		Date:        time.Now().Add(time.Hour),
		User:        &u1,
	}
	b2 := models.BulletinRide{
		Source:      "Source 2",
		Destination: "Destination 2",
		Body:        generateBulletinTestBody(),
		Offer:       lib.BoolToPtr(false),
		Date:        time.Now().Add(time.Hour),
		User: &models.User{
			Type:   models.UserTypeStudent,
			Name:   "User 2",
			UnixID: "u2",
		},
	}
	assert.NoError(db.Create(&u1).Create(&b1).Create(&b2).Error)

	utils.AddUserContexts(router, u1.ID)
	SetupRouter(router, db, cfg)

	/* Delete ride with bad bulletin (expect failure) */
	apiErr := lib.ErrorRecordNotFound
	w, err := utils.DoHTTPReq(router, http.MethodDelete, fmt.Sprintf("/rides/%d", 42), nil)
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Delete ride with bad owner (expect failure) */
	apiErr = lib.ErrorMustBeSelf
	w, err = utils.DoHTTPReq(router, http.MethodDelete, fmt.Sprintf("/rides/%d", b2.ID), nil)
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Delete ride (expect success) */
	w, err = utils.DoHTTPReq(router, http.MethodDelete, fmt.Sprintf("/rides/%d", b1.ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.BulletinRide
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Assert good response
	assert.Equal(b1.ID, resp.ID)
	assert.Equal(b1.Body, resp.Body)

	// Assert not in db
	var count int
	assert.NoError(db.Model(&models.BulletinRide{}).Where("id = ?", resp.ID).Count(&count).Error)
	assert.Equal(0, count)
}
