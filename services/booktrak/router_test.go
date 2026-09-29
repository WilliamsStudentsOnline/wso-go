package booktrak_test

import (
	"net/http"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	. "github.com/WilliamsStudentsOnline/wso-go/services/booktrak"
	testify "github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestSetupRouter_ReturnsGone(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	router := utils.SetupRouter(auth.ScopeBooktrak, auth.ScopeBooktrakWrite)
	cfg := utils.SetupConfig()

	SetupRouter(router, db, cfg, zaptest.NewLogger(t).Sugar())

	paths := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/health-check"},
		{http.MethodGet, "/books"},
		{http.MethodGet, "/books/web"},
		{http.MethodGet, "/books/1"},
		{http.MethodPost, "/books"},
		{http.MethodPatch, "/books/1"},
		{http.MethodGet, "/listings"},
		{http.MethodPost, "/listings"},
		{http.MethodGet, "/listings/1"},
		{http.MethodPut, "/listings/1"},
		{http.MethodDelete, "/listings/1"},
	}

	apiErr := lib.ErrorBooktrakDeprecated
	for _, tc := range paths {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			assert := testify.New(t)
			w, err := utils.DoHTTPReq(router, tc.method, tc.path, nil)
			assert.NoError(err)
			assert.Equal(apiErr.HTTPCode, w.Code)
			assert.Equal(apiErr.Code, utils.GetHTTPDataResp(assert, w.Body.Bytes()).Error.ErrorCode)
		})
	}
}
