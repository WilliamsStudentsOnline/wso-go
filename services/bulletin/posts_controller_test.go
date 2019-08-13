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

func TestController_GetPost(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeBulletin, auth.ScopeWriteSelf, auth.ScopeUsers)
	cfg := utils.SetupConfig()

	u1 := models.User{
		Type:   models.UserTypeStudent,
		UnixID: "u1",
		Name:   "User 1",
	}

	p1 := models.Post{
		User: &u1,
		Discussion: &models.Discussion{
			Title:      generateDiscussionTestTitle(),
			User:       &u1,
			ExUserName: u1.Name,
		},
		Content:    generateBulletinTestBody(),
		ExUserName: u1.Name,
	}

	assert.NoError(db.Create(&u1).Create(&p1).Error)

	utils.AddUserContexts(router, u1.ID)
	SetupRouter(router, db, cfg)

	/* Get test discussion (signed in) */
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/posts/%d", p1.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.Post
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct post
	assert.Equal(p1.ID, resp.ID)
	assert.Equal(p1.Content, resp.Content)
	// User should not be nil, as signed in
	assert.NotNil(resp.User)
	assert.NotEmpty(resp.ExUserName)

	/* Get test discussion (signed out) */
	r1 := utils.SetupRouter(auth.ScopeBulletin, auth.ScopeWriteSelf)
	SetupRouter(r1, db, cfg)
	w, err = utils.DoHTTPReq(r1, http.MethodGet, fmt.Sprintf("/posts/%d", p1.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	resp = models.Post{}
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct post
	assert.Equal(p1.ID, resp.ID)
	// Ensure that user is missing
	assert.Nil(resp.User)
	assert.Empty(resp.ExUserName)

	/* Get test post bad id (expect failure) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/posts/%d", 42), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusNotFound, w.Code)
}

func TestController_CreatePost(t *testing.T) {
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
	d1 := models.Discussion{
		Title:      generateDiscussionTestTitle(),
		User:       &u1,
		ExUserName: u1.Name,
		LastActive: time.Now().Add(-time.Hour * 24),
	}
	assert.NoError(db.Create(&u1).Create(&d1).Error)

	utils.AddUserContexts(router, u1.ID)
	SetupRouter(router, db, cfg)

	/* Create ride with missing data (expect failure) */
	apiErr := lib.ErrorRequestDataValidationFailed
	w, err := utils.DoHTTPReq(router, http.MethodPost, "/posts", bytes.NewBufferString(`{"content": "hi"}`))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Create ride with unknown discussion (expect failure) */
	apiErr = lib.ErrorDiscussionNotFound
	w, err = utils.DoHTTPReq(router, http.MethodPost, "/posts", bytes.NewBufferString(`{"content": "hi", "discussionID": 42}`))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Create discussion (expect success) */
	params := CreatePostParams{
		DiscussionID: d1.ID,
		Content:      generateBulletinTestBody(),
	}
	paramsData, err := json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(router, http.MethodPost, "/posts", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(http.StatusCreated, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.Post
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Assert good response
	assert.NotZero(resp.ID)
	assert.Equal(params.Content, resp.Content)
	assert.Equal(u1.Name, resp.ExUserName)

	// Assert found post in DB
	var count int
	assert.NoError(db.Model(&models.Post{}).Where("id = ?", resp.ID).Count(&count).Error)
	assert.Equal(1, count)

	// Assert that last active time of discussion has changed in DB
	var disc models.Discussion
	assert.NoError(db.First(&disc, d1.ID).Error)
	// NOTE: This might randomly fail (idk actually) if the seconds don't match up.
	assert.Equal(resp.CreatedTime.Unix(), disc.LastActive.Unix())
}

func TestController_UpdatePost(t *testing.T) {
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
	d1 := models.Discussion{
		Title:      generateDiscussionTestTitle(),
		User:       &u1,
		ExUserName: u1.Name,
	}
	p1 := models.Post{
		User:       &u1,
		Discussion: &d1,
		Content:    generateBulletinTestBody(),
	}
	p2 := models.Post{
		User: &models.User{
			Type:   models.UserTypeStudent,
			Name:   "User 2",
			UnixID: "u2",
		},
		Discussion: &d1,
		Content:    generateBulletinTestBody(),
	}
	assert.NoError(db.Create(&u1).Create(&d1).Create(&p1).Create(&p2).Error)

	utils.AddUserContexts(router, u1.ID)
	SetupRouter(router, db, cfg)

	/* Update post with bad user (expect failure) */
	apiErr := lib.ErrorMustBeSelf
	w, err := utils.DoHTTPReq(router, http.MethodPatch, fmt.Sprintf("/posts/%d", p2.ID),
		bytes.NewBufferString(`{"content": "hi"}`))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Update post (expect success) */
	params := UpdatePostParams{
		Content: generateBulletinTestBody(),
	}
	paramsData, err := json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(router, http.MethodPatch, fmt.Sprintf("/posts/%d", p1.ID), bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.Post
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Assert good response
	assert.Equal(p1.ID, resp.ID)
	assert.Equal(params.Content, resp.Content)

	// Assert that updated in DB
	var postDB models.Post
	assert.NoError(db.First(&postDB, p1.ID).Error)
	assert.Equal(params.Content, postDB.Content)
}

func TestController_DeletePost(t *testing.T) {
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
	badP1 := models.Post{
		User: &models.User{
			Type:   models.UserTypeStudent,
			Name:   "User 2",
			UnixID: "u2",
		},
		Discussion: &models.Discussion{
			Title:      generateDiscussionTestTitle(),
			User:       &u1,
			ExUserName: u1.Name,
		},
		Content: generateBulletinTestBody(),
	}
	d1 := models.Discussion{
		Title: generateDiscussionTestTitle(),
		User: &models.User{
			Type:   models.UserTypeStudent,
			Name:   "User 3",
			UnixID: "u3",
		},
		Posts: []*models.Post{
			{
				User:    &u1,
				Content: generateBulletinTestBody(),
			},
			{
				User:    &u1,
				Content: generateBulletinTestBody(),
			},
		},
		LastActive: time.Now().Add(-time.Hour * 6),
		ExUserName: u1.Name,
	}
	d1.Posts[0].CreatedAt = time.Now().Add(-time.Hour * 24)
	d1.Posts[1].CreatedAt = d1.LastActive
	assert.NoError(db.Create(&u1).Create(&d1).Create(&badP1).Error)

	utils.AddUserContexts(router, u1.ID)
	SetupRouter(router, db, cfg)

	/* Delete post with bad user (expect failure) */
	apiErr := lib.ErrorMustBeSelf
	w, err := utils.DoHTTPReq(router, http.MethodDelete, fmt.Sprintf("/posts/%d", badP1.ID), nil)
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Delete post with bad post (expect failure) */
	apiErr = lib.ErrorRecordNotFound
	w, err = utils.DoHTTPReq(router, http.MethodDelete, fmt.Sprintf("/posts/%d", 42), nil)
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Delete post (expect success) */
	w, err = utils.DoHTTPReq(router, http.MethodDelete, fmt.Sprintf("/posts/%d", d1.Posts[1].ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.Post
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Assert good response
	assert.Equal(d1.Posts[1].ID, resp.ID)
	assert.Equal(d1.Posts[1].Content, resp.Content)

	// Assert that deleted in DB
	var count int
	assert.NoError(db.Model(&models.Post{}).Where("id = ?", d1.Posts[1].ID).Count(&count).Error)
	assert.Equal(0, count)

	// Assert that last active time of discussion has changed in DB to old one
	var disc models.Discussion
	assert.NoError(db.First(&disc, d1.ID).Error)
	// NOTE: This might randomly fail (idk actually) if the seconds don't match up.
	assert.Equal(d1.Posts[0].CreatedAt.Unix(), disc.LastActive.Unix())

	/* Delete last post of discussion (expect success and discussion delete) */
	w, err = utils.DoHTTPReq(router, http.MethodDelete, fmt.Sprintf("/posts/%d", d1.Posts[0].ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	resp = models.Post{}
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Assert good response
	assert.Equal(d1.Posts[0].ID, resp.ID)
	assert.Equal(d1.Posts[0].Content, resp.Content)

	// Assert that deleted in DB
	count = 0
	assert.NoError(db.Model(&models.Post{}).Where("id = ?", d1.Posts[0].ID).Count(&count).Error)
	assert.Equal(0, count)

	// Assert that discussion also deleted in DB
	count = 0
	assert.NoError(db.Model(&models.Discussion{}).Where("id = ?", d1.ID).Count(&count).Error)
	assert.Equal(0, count)
}
