package bulletin_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
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

func TestController_ListBulletins(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeBulletin, auth.ScopeWriteSelf)
	cfg := utils.SetupConfig()

	bulletins := []*models.Bulletin{
		// 0
		{
			Type: models.BulletinTypeAnnouncement,
		},
		// 1
		{
			Type: models.BulletinTypeAnnouncement,
		},
		// 2
		{
			Type: models.BulletinTypeLostAndFound,
		},
		// 3
		{
			Type: models.BulletinTypeJob,
		},
		// 4
		{
			Type: models.BulletinTypeExchange,
		},
		// 5: Announcement started a day ago
		{
			Type:      models.BulletinTypeAnnouncement,
			StartDate: time.Now().Add(-1 * time.Hour * 24),
		},
		// 6: Announcement will start next day
		{
			Type:      models.BulletinTypeAnnouncement,
			StartDate: time.Now().Add(time.Hour * 24),
		},
		// 7: Announcement ended yesterday
		{
			Type:      models.BulletinTypeAnnouncement,
			StartDate: time.Now().Add(-1 * time.Hour * 32),
			EndDate:   lib.TimeToPtr(time.Now().Add(-1 * time.Hour * 24)),
		},
		// 8: Announcement will end next day
		{
			Type:    models.BulletinTypeAnnouncement,
			EndDate: lib.TimeToPtr(time.Now().Add(time.Hour * 24)),
		},
	}
	for i, val := range bulletins {
		val.Title = fmt.Sprintf("Bulletin %d", i)
		val.Body = generateBulletinTestBody()
		if val.StartDate.IsZero() {
			val.StartDate = time.Now().Add(time.Second * time.Duration(i))
		}
		// Push it off by an hour so we don't have any weird errors
		val.StartDate = val.StartDate.Add(-time.Hour)
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
			// 8, 4, 3, 1, 0, 5
			[]int{8, 4, 3, 2, 1, 0, 5},
		},
		{
			"type announcement",
			"type=announcement",
			[]int{8, 1, 0, 5},
		},
		{
			"type lostAndFound",
			"type=lostAndFound",
			[]int{2},
		},
		{
			"type job",
			"type=job",
			[]int{3},
		},
		{
			"type exchange",
			"type=exchange",
			[]int{4},
		},
		{
			"limit",
			"limit=2",
			[]int{8, 4},
		},
		{
			"limit and offset",
			"limit=2&offset=" + bulletins[3].StartDate.Format(time.RFC3339),
			[]int{2, 1},
		},
		{
			"all bulletins",
			"all=true",
			[]int{6, 8, 4, 3, 2, 1, 0, 5, 7},
		},
	}

	SetupRouter(router, db, cfg)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := testify.New(t)
			// Get test user
			w, err := utils.DoHTTPReq(router, http.MethodGet, "/bulletins?"+tc.query, nil)
			a.NoError(err)

			// Status is okay
			a.Equal(http.StatusOK, w.Code)

			// Decode response
			respData := utils.GetHTTPDataResp(a, w.Body.Bytes())
			a.Nil(respData.Error)
			var resp []*models.Bulletin
			a.NoError(json.Unmarshal(respData.Data, &resp))

			// Check if correct result
			a.Len(resp, len(tc.expected))
			for i := range tc.expected {
				a.Equal(bulletins[tc.expected[i]].ID, resp[i].ID)
				a.Equal(bulletins[tc.expected[i]].Body, resp[i].Body)
				a.Equal(bulletins[tc.expected[i]].Type, resp[i].Type)
			}
		})
	}
}

func TestController_GetBulletin(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeBulletin, auth.ScopeWriteSelf, auth.ScopeUsers)
	cfg := utils.SetupConfig()

	b1 := models.Bulletin{
		Type:  models.BulletinTypeAnnouncement,
		Title: "Bulletin 1",
		Body:  generateBulletinTestBody(),
		Offer: lib.BoolToPtr(false),
		User: &models.User{
			Type:   models.UserTypeStudent,
			UnixID: "u1",
			Name:   "User 1",
		},
	}

	assert.NoError(db.Create(&b1).Error)

	utils.AddUserContexts(router, b1.User.ID)
	SetupRouter(router, db, cfg)

	/* Get test bulletin (signed in) */
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/bulletins/%d", b1.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.Bulletin
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct bulletin
	assert.Equal(b1.ID, resp.ID)
	assert.Equal(b1.Body, resp.Body)
	assert.False(*resp.Offer)
	// User should not be nil, as signed in
	assert.NotNil(resp.User)

	/* Get test bulletin (signed out) */
	r1 := utils.SetupRouter(auth.ScopeBulletin, auth.ScopeWriteSelf)
	SetupRouter(r1, db, cfg)
	w, err = utils.DoHTTPReq(r1, http.MethodGet, fmt.Sprintf("/bulletins/%d", b1.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	resp = models.Bulletin{}
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct bulletin
	assert.Equal(b1.ID, resp.ID)
	// Ensure that user is missing
	assert.Nil(resp.User)

	/* Get test bulletin bad id (expect failure) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/bulletins/%d", 42), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusNotFound, w.Code)
}

func TestController_CreateBulletin(t *testing.T) {
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

	/* Create bulletin with missing data (expect failure) */
	apiErr := lib.ErrorRequestDataValidationFailed
	w, err := utils.DoHTTPReq(router, http.MethodPost, "/bulletins", bytes.NewBufferString(`{"title": "hi"}`))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Create bulletin with bad dates (expect failure) */
	params := CreateBulletinParams{
		Type:    models.BulletinTypeAnnouncement,
		Title:   "Bulletin 1",
		Body:    generateBulletinTestBody(),
		EndDate: lib.TimeToPtr(time.Now().Add(-time.Hour)),
	}
	apiErr = lib.ErrorBulletinInvalidDates
	paramsData, err := json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(router, http.MethodPost, "/bulletins", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Create bulletin with bad type (expect failure) */
	params = CreateBulletinParams{
		Type:  "foobar",
		Title: "Bulletin 1",
		Body:  generateBulletinTestBody(),
	}
	apiErr = lib.ErrorBulletinInvalidType
	paramsData, err = json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(router, http.MethodPost, "/bulletins", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Create bulletin (expect success) */
	params = CreateBulletinParams{
		Type:      models.BulletinTypeAnnouncement,
		Title:     "Bulletin 1",
		Body:      generateBulletinTestBody(),
		Offer:     lib.BoolToPtr(false),
		StartDate: lib.TimeToPtr(time.Now().Add(24 * time.Hour)),
		EndDate:   lib.TimeToPtr(time.Now().Add(36 * time.Hour)),
	}
	paramsData, err = json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(router, http.MethodPost, "/bulletins", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(http.StatusCreated, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.Bulletin
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Assert good response
	assert.NotZero(resp.ID)
	assert.Equal(params.Body, resp.Body)
	assert.Equal(params.Offer, resp.Offer)

	// Assert found in DB
	var count int
	assert.NoError(db.Model(&models.Bulletin{}).Where("id = ?", resp.ID).Count(&count).Error)
	assert.Equal(1, count)
}

func TestController_UpdateBulletin(t *testing.T) {
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
	b1 := models.Bulletin{
		Type:      models.BulletinTypeAnnouncement,
		Title:     "Bulletin 1",
		Body:      generateBulletinTestBody(),
		StartDate: time.Now(),
		User:      &u1,
	}
	b2 := models.Bulletin{
		Type:      models.BulletinTypeExchange,
		Title:     "Bulletin 2",
		Body:      generateBulletinTestBody(),
		StartDate: time.Now(),
		User: &models.User{
			Type:   models.UserTypeStudent,
			Name:   "User 2",
			UnixID: "u2",
		},
	}
	assert.NoError(db.Create(&u1).Create(&b1).Create(&b2).Error)

	utils.AddUserContexts(router, u1.ID)
	SetupRouter(router, db, cfg)

	/* Update bulletin with bad bulletin (expect failure) */
	params := UpdateBulletinParams{
		Body: lib.StrToPtr(generateBulletinTestBody()),
	}
	apiErr := lib.ErrorRecordNotFound
	paramsData, err := json.Marshal(&params)
	assert.NoError(err)
	w, err := utils.DoHTTPReq(router, http.MethodPatch, fmt.Sprintf("/bulletins/%d", 42), bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Update bulletin with bad owner (expect failure) */
	params = UpdateBulletinParams{
		Body: lib.StrToPtr(generateBulletinTestBody()),
	}
	apiErr = lib.ErrorMustBeSelf
	paramsData, err = json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(router, http.MethodPatch, fmt.Sprintf("/bulletins/%d", b2.ID), bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Update bulletin with bad dates (expect failure) */
	params = UpdateBulletinParams{
		Body:    lib.StrToPtr(generateBulletinTestBody()),
		EndDate: lib.TimeToPtr(time.Now().Add(-time.Hour)),
	}
	apiErr = lib.ErrorBulletinInvalidDates
	paramsData, err = json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(router, http.MethodPatch, fmt.Sprintf("/bulletins/%d", b1.ID), bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Update bulletin (expect success) */
	params = UpdateBulletinParams{
		Body: lib.StrToPtr(generateBulletinTestBody()),
	}
	paramsData, err = json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(router, http.MethodPatch, fmt.Sprintf("/bulletins/%d", b1.ID), bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.Bulletin
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Assert good response
	assert.Equal(b1.ID, resp.ID)
	assert.Equal(*params.Body, resp.Body)
	assert.Equal(params.Offer, resp.Offer)

	// Assert updated in DB
	var respDB models.Bulletin
	assert.NoError(db.First(&respDB, b1.ID).Error)
	assert.Equal(*params.Body, respDB.Body)
}

func TestController_DeleteBulletin(t *testing.T) {
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
	b1 := models.Bulletin{
		Type:      models.BulletinTypeAnnouncement,
		Title:     "Bulletin 1",
		Body:      generateBulletinTestBody(),
		StartDate: time.Now(),
		User:      &u1,
	}
	b2 := models.Bulletin{
		Type:      models.BulletinTypeExchange,
		Title:     "Bulletin 2",
		Body:      generateBulletinTestBody(),
		StartDate: time.Now(),
		User: &models.User{
			Type:   models.UserTypeStudent,
			Name:   "User 2",
			UnixID: "u2",
		},
	}
	assert.NoError(db.Create(&u1).Create(&b1).Create(&b2).Error)

	utils.AddUserContexts(router, u1.ID)
	SetupRouter(router, db, cfg)

	/* Delete bulletin with bad bulletin (expect failure) */
	apiErr := lib.ErrorRecordNotFound
	w, err := utils.DoHTTPReq(router, http.MethodDelete, fmt.Sprintf("/bulletins/%d", 42), nil)
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Delete bulletin with bad owner (expect failure) */
	apiErr = lib.ErrorMustBeSelf
	w, err = utils.DoHTTPReq(router, http.MethodDelete, fmt.Sprintf("/bulletins/%d", b2.ID), nil)
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Delete bulletin (expect success) */
	w, err = utils.DoHTTPReq(router, http.MethodDelete, fmt.Sprintf("/bulletins/%d", b1.ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.Bulletin
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Assert good response
	assert.Equal(b1.ID, resp.ID)
	assert.Equal(b1.Body, resp.Body)

	// Assert not in DB
	var count int
	assert.NoError(db.Model(&models.Bulletin{}).Where("id = ?", resp.ID).Count(&count).Error)
	assert.Equal(0, count)
}

func generateBulletinTestBody() string {
	randBytes := make([]byte, 290)
	for i := 0; i < 100; i++ {
		randBytes[i] = byte(65 + rand.Intn(25)) //A=65 and Z = 65+25
	}
	str := string(randBytes)
	return str
}
