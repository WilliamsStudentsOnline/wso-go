package banner_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/services/banner"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

// Skip locally if redis isn't running (local dev only)
func configureRedis(t *testing.T) {
	t.Helper()
	err := banner.ConfigureControllerForTest()
	if err != nil && os.Getenv("GITHUB_ACTIONS") == "" {
		t.Skipf("redis not reachable: %v", err)
	}
	testify.NoError(t, err)
}

func TestBanner_PostRequiresAdmin(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeUsers)
	utils.AddUserContexts(router, 1)

	cfg := utils.SetupConfig()
	log := zaptest.NewLogger(t).Sugar()
	banner.SetupPublicRouter(router, db, cfg, log)
	banner.SetupAdminRouter(router, db, cfg, log)

	params := banner.Banner{
		Visible: true,
		Type:    banner.BannerTypePinned,
		Text:    "banner title",
		Link:    "https://example.com/banner",
		Sub:     "banner subtitle",
	}
	payload, err := json.Marshal(params)
	assert.NoError(err)

	w, err := utils.DoHTTPReq(router, http.MethodPost, "/", bytes.NewBuffer(payload))
	assert.NoError(err)
	utils.CheckRespError(assert, w, lib.ErrorNoScopeAuthorization)
}

func TestBanner_GetIsPublic(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter()
	cfg := utils.SetupConfig()
	log := zaptest.NewLogger(t).Sugar()
	banner.SetupPublicRouter(router, db, cfg, log)
	banner.SetupAdminRouter(router, db, cfg, log)
	configureRedis(t)

	w, err := utils.DoHTTPReq(router, http.MethodGet, "/", nil)
	assert.NoError(err)
	respData := utils.GetGoodResp(assert, w)

	var getResp banner.Banner
	assert.NoError(json.Unmarshal(respData.Data, &getResp))
	assert.False(getResp.Visible)
}

func TestBanner_RedisSetAndGet(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeAdminAll)
	utils.AddUserContexts(router, 1)

	cfg := utils.SetupConfig()
	log := zaptest.NewLogger(t).Sugar()
	banner.SetupPublicRouter(router, db, cfg, log)
	banner.SetupAdminRouter(router, db, cfg, log)
	configureRedis(t)

	params := banner.Banner{
		Visible: true,
		Type:    banner.BannerTypePinned,
		Text:    "banner title",
		Link:    "https://example.com/banner",
		Sub:     "banner subtitle",
	}
	payload, err := json.Marshal(params)
	assert.NoError(err)

	w, err := utils.DoHTTPReq(router, http.MethodPost, "/", bytes.NewBuffer(payload))
	assert.NoError(err)
	respData := utils.GetGoodResp(assert, w)

	var setResp banner.Banner
	assert.NoError(json.Unmarshal(respData.Data, &setResp))
	assert.Equal(params, setResp)

	w, err = utils.DoHTTPReq(router, http.MethodGet, "/", nil)
	assert.NoError(err)
	respData = utils.GetGoodResp(assert, w)

	var getResp banner.Banner
	assert.NoError(json.Unmarshal(respData.Data, &getResp))
	assert.Equal(params, getResp)
}

func TestBanner_HideWithVisibleFalse(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeAdminAll)
	utils.AddUserContexts(router, 1)

	cfg := utils.SetupConfig()
	log := zaptest.NewLogger(t).Sugar()
	banner.SetupPublicRouter(router, db, cfg, log)
	banner.SetupAdminRouter(router, db, cfg, log)
	configureRedis(t)

	seed := banner.Banner{
		Visible: true,
		Type:    banner.BannerTypeDismissible,
		Text:    "Visible",
		Link:    "https://example.com",
		Sub:     "sub",
	}
	payload, err := json.Marshal(seed)
	assert.NoError(err)
	w, err := utils.DoHTTPReq(router, http.MethodPost, "/", bytes.NewBuffer(payload))
	assert.NoError(err)
	assert.Equal(http.StatusOK, w.Code)

	// Hide while keeping copy for later re-enable
	hidden := banner.Banner{
		Visible: false,
		Type:    seed.Type,
		Text:    seed.Text,
		Link:    seed.Link,
		Sub:     seed.Sub,
	}
	payload, err = json.Marshal(hidden)
	assert.NoError(err)
	w, err = utils.DoHTTPReq(router, http.MethodPost, "/", bytes.NewBuffer(payload))
	assert.NoError(err)
	respData := utils.GetGoodResp(assert, w)

	var setResp banner.Banner
	assert.NoError(json.Unmarshal(respData.Data, &setResp))
	assert.Equal(hidden, setResp)

	w, err = utils.DoHTTPReq(router, http.MethodGet, "/", nil)
	assert.NoError(err)
	respData = utils.GetGoodResp(assert, w)

	var getResp banner.Banner
	assert.NoError(json.Unmarshal(respData.Data, &getResp))
	assert.Equal(hidden, getResp)
	assert.False(getResp.Visible)
}
