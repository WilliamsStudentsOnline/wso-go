package ephmatch_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	. "github.com/WilliamsStudentsOnline/wso-go/services/ephmatch"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestSetupRouter_Gone(t *testing.T) {
	assert := testify.New(t)
	router := utils.SetupRouter()
	SetupRouter(router.Group("/ephmatch"), nil, utils.SetupConfig(), zaptest.NewLogger(t).Sugar())

	for _, req := range []struct{ method, path string }{
		{http.MethodGet, "/ephmatch/availability"},
		{http.MethodGet, "/ephmatch/profile"},
		{http.MethodPut, "/ephmatch/profile/photo"},
		{http.MethodGet, "/ephmatch/matches"},
		{http.MethodDelete, "/ephmatch/matches/1"},
		{http.MethodPut, "/ephmatch/profiles/1/relation"},
		{http.MethodPost, "/ephmatch/anything/else"},
	} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(req.method, req.path, nil))
		assert.Equal(http.StatusGone, w.Code, "%s %s", req.method, req.path)
	}
}
