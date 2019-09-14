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

func TestController_ListDiscussions(t *testing.T) {
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

	discussions := []*models.Discussion{
		// 0
		{},
		// 1
		{
			Posts: []*models.Post{
				{
					User:    &u1,
					Content: generateBulletinTestBody(),
				},
				{
					User:    &u1,
					Content: generateBulletinTestBody(),
				},
				{
					User: &models.User{
						Type:   models.UserTypeStudent,
						Name:   "User 2",
						UnixID: "u2",
					},
					Content: generateBulletinTestBody(),
				},
			},
		},
		// 2
		{
			Posts: []*models.Post{
				{
					User:    &u1,
					Content: generateBulletinTestBody(),
				},
			},
		},
		// 3
		{},
	}
	for i, val := range discussions {
		val.Title = generateDiscussionTestTitle()
		val.User = &u1
		val.LastActive = time.Now().Add(time.Minute * time.Duration(i))
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
			[]int{3, 2, 1, 0},
		},
		{
			"limit",
			"limit=2",
			[]int{3, 2},
		},
		{
			"limit and start",
			"limit=2&start=" + discussions[2].LastActive.Format(time.RFC3339),
			[]int{1, 0},
		},
	}

	SetupRouter(router, db, cfg)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := testify.New(t)
			// Get test discussion
			w, err := utils.DoHTTPReq(router, http.MethodGet, "/discussions?"+tc.query, nil)
			a.NoError(err)

			// Status is okay
			a.Equal(http.StatusOK, w.Code)

			// Decode response
			respData := utils.GetHTTPDataResp(a, w.Body.Bytes())
			a.Nil(respData.Error)
			var resp []*models.Discussion
			a.NoError(json.Unmarshal(respData.Data, &resp))

			// Check if correct result
			a.Len(resp, len(tc.expected))
			for i := range tc.expected {
				a.Equal(discussions[tc.expected[i]].ID, resp[i].ID)
				a.Equal(discussions[tc.expected[i]].Title, resp[i].Title)
			}
		})
	}

	t.Run("getLastPost", func(t *testing.T) {
		a := testify.New(t)
		// Get test user
		w, err := utils.DoHTTPReq(router, http.MethodGet, "/discussions?getLastPost=true", nil)
		a.NoError(err)

		// Status is okay
		a.Equal(http.StatusOK, w.Code)

		// Decode response
		respData := utils.GetHTTPDataResp(a, w.Body.Bytes())
		a.Nil(respData.Error)
		var resp []*models.Discussion
		a.NoError(json.Unmarshal(respData.Data, &resp))

		// Check if correct result
		a.Len(resp, 4)
		// Discussion 1 (diff b/c of ordering)
		a.Equal(discussions[1].ID, resp[2].ID)
		a.Equal(discussions[1].Title, resp[2].Title)
		a.Len(resp[1].Posts, 1)
		a.Equal(discussions[1].Posts[2].Content, resp[2].Posts[0].Content)
		// Discussion 2 (diff idx b/c of order)
		a.Equal(discussions[2].ID, resp[1].ID)
		a.Equal(discussions[2].Title, resp[1].Title)
		a.Len(resp[2].Posts, 1)
		a.Equal(discussions[2].Posts[0].Content, resp[1].Posts[0].Content)
	})
}

func TestController_GetDiscussion(t *testing.T) {
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

	d1 := models.Discussion{
		Title:      generateDiscussionTestTitle(),
		User:       &u1,
		ExUserName: u1.Name,
		Posts: []*models.Post{
			{
				User:       &u1,
				Content:    generateBulletinTestBody(),
				ExUserName: u1.Name,
			},
		},
	}

	assert.NoError(db.Create(&u1).Create(&d1).Error)

	utils.AddUserContexts(router, u1.ID)
	SetupRouter(router, db, cfg)

	/* Get test discussion (signed in) */
	w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/discussions/%d?preload[]=user&preload[]=posts&preload[]=postsUsers", d1.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.Discussion
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct ride
	assert.Equal(d1.ID, resp.ID)
	assert.Equal(d1.Title, resp.Title)
	assert.Len(resp.Posts, len(d1.Posts))
	assert.Equal(d1.Posts[0].Content, resp.Posts[0].Content)
	// User should not be nil, as signed in
	assert.NotNil(resp.User)
	assert.NotEmpty(resp.ExUserName)
	assert.NotNil(resp.Posts[0].User)
	assert.NotEmpty(resp.Posts[0].ExUserName)

	/* Get test discussion (signed out) */
	r1 := utils.SetupRouter(auth.ScopeBulletin, auth.ScopeWriteSelf)
	SetupRouter(r1, db, cfg)
	w, err = utils.DoHTTPReq(r1, http.MethodGet, fmt.Sprintf("/discussions/%d?preload[]=user&preload[]=posts&preload[]=postsUsers", d1.ID), nil)
	assert.NoError(err)

	// Status is okay
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData = utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	resp = models.Discussion{}
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Check if correct ride
	assert.Equal(d1.ID, resp.ID)
	// Ensure that user is missing
	assert.Nil(resp.User)
	assert.Empty(resp.ExUserName)
	assert.Nil(resp.Posts[0].User)
	assert.Empty(resp.Posts[0].ExUserName)

	/* Get test ride bad id (expect failure) */
	w, err = utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/discussions/%d?preload[]=user&preload[]=posts&preload[]=postsUsers", 42), nil)
	assert.NoError(err)

	// Status is not found
	assert.Equal(http.StatusNotFound, w.Code)
}

func TestController_GetDiscussionPosts(t *testing.T) {
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

	d1 := models.Discussion{
		Title: generateDiscussionTestTitle(),
		User:  &u1,
	}
	assert.NoError(db.Create(&d1).Error)

	posts := []*models.Post{
		// 0
		{},
		// 1
		{},
		// 2
		{},
		// 3
		{},
	}
	for i, val := range posts {
		val.Content = generateBulletinTestBody()
		val.User = &u1
		val.Discussion = &d1
		val.CreatedAt = time.Now().Add(time.Minute * time.Duration(i))
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
			[]int{3, 2, 1, 0},
		},
		{
			"limit",
			"limit=2",
			[]int{3, 2},
		},
		{
			"limit and start",
			"limit=2&start=" + posts[2].CreatedAt.Format(time.RFC3339),
			[]int{1, 0},
		},
	}

	SetupRouter(router, db, cfg)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := testify.New(t)
			// Get test user
			w, err := utils.DoHTTPReq(router, http.MethodGet, fmt.Sprintf("/discussions/%d/posts?%s", d1.ID, tc.query), nil)
			a.NoError(err)

			// Status is okay
			a.Equal(http.StatusOK, w.Code)

			// Decode response
			respData := utils.GetHTTPDataResp(a, w.Body.Bytes())
			a.Nil(respData.Error)
			var resp []*models.Post
			a.NoError(json.Unmarshal(respData.Data, &resp))

			// Check if correct result
			a.Len(resp, len(tc.expected))
			for i := range tc.expected {
				a.Equal(posts[tc.expected[i]].ID, resp[i].ID)
				a.Equal(posts[tc.expected[i]].Content, resp[i].Content)
			}
		})
	}
}

func TestController_CreateDiscussion(t *testing.T) {
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
	w, err := utils.DoHTTPReq(router, http.MethodPost, "/discussions", bytes.NewBufferString(`{"title": "hi"}`))
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Create discussion (expect success) */
	params := CreateDiscussionParams{
		Title:   generateDiscussionTestTitle(),
		Content: generateBulletinTestBody(),
	}
	paramsData, err := json.Marshal(&params)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(router, http.MethodPost, "/discussions", bytes.NewBuffer(paramsData))
	assert.NoError(err)
	assert.Equal(http.StatusCreated, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.Discussion
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Assert good response
	assert.NotZero(resp.ID)
	assert.Equal(params.Title, resp.Title)
	assert.Equal(u1.Name, resp.ExUserName)
	assert.Len(resp.Posts, 1)
	assert.Equal(params.Content, resp.Posts[0].Content)
	assert.Equal(u1.Name, resp.Posts[0].ExUserName)

	// Assert found discussion in DB
	var count int
	assert.NoError(db.Model(&models.Discussion{}).Where("id = ?", resp.ID).Count(&count).Error)
	assert.Equal(1, count)

	// Assert found post in DB
	count = 0
	assert.NoError(db.Model(&models.Post{}).Where("id = ?", resp.Posts[0].ID).Count(&count).Error)
	assert.Equal(1, count)
}

func TestController_DeleteDiscussion(t *testing.T) {
	// Setup (can copy and paste this basically)
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeBulletin, auth.ScopeWriteSelf, auth.ScopeAdminAll)
	cfg := utils.SetupConfig()

	u1 := models.User{
		Type:   models.UserTypeStudent,
		Name:   "User 1",
		UnixID: "u1",
	}
	assert.NoError(db.Create(&u1).Error)

	d1 := models.Discussion{
		User:  &u1,
		Title: generateDiscussionTestTitle(),
		Posts: []*models.Post{
			{
				Content: generateBulletinTestBody(),
				User:    &u1,
			},
		},
		CreatedTime: time.Time{},
	}
	assert.NoError(db.Create(&d1).Error)

	utils.AddUserContexts(router, u1.ID)
	SetupRouter(router, db, cfg)

	/* Delete discussion with incorrect scopes (expect failure) */
	apiErr := lib.ErrorNoScopeAuthorization
	r1 := utils.SetupRouter(auth.ScopeBulletin, auth.ScopeWriteSelf)
	SetupRouter(r1, db, cfg)
	w, err := utils.DoHTTPReq(r1, http.MethodDelete, fmt.Sprintf("/discussions/%d", d1.ID), nil)
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Delete discussion that does not exist (expect failure) */
	apiErr = lib.ErrorRecordNotFound
	w, err = utils.DoHTTPReq(router, http.MethodDelete, fmt.Sprintf("/discussions/%d", 42), nil)
	assert.NoError(err)
	assert.Equal(apiErr.HTTPCode, w.Code)
	assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)

	/* Delete discussion (expect success) */
	w, err = utils.DoHTTPReq(router, http.MethodDelete, fmt.Sprintf("/discussions/%d", d1.ID), nil)
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Decode response
	respData := utils.GetHTTPDataResp(assert, w.Body.Bytes())
	assert.Nil(respData.Error)
	var resp models.Discussion
	assert.NoError(json.Unmarshal(respData.Data, &resp))

	// Assert good response
	assert.Equal(d1.ID, resp.ID)

	// Assert no discussion in DB
	var count int
	assert.NoError(db.Model(&models.Discussion{}).Where("id = ?", d1.ID).Count(&count).Error)
	assert.Equal(0, count)

	// Assert no post in DB
	count = 0
	assert.NoError(db.Model(&models.Post{}).Where("id = ?", d1.Posts[0].ID).Count(&count).Error)
	assert.Equal(0, count)
}

func generateDiscussionTestTitle() string {
	randBytes := make([]byte, 100)
	for i := 0; i < 100; i++ {
		randBytes[i] = byte(65 + rand.Intn(25)) //A=65 and Z = 65+25
	}
	str := string(randBytes)
	return str
}
