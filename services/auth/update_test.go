package auth_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	libauth "github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	utils "github.com/WilliamsStudentsOnline/wso-go/lib/test_utils"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services/auth"
	"github.com/WilliamsStudentsOnline/wso-go/services/auth/api"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	testify "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func tokenScopes(t *testing.T, mw *jwt.GinJWTMiddleware, token string) interface{} {
	t.Helper()
	parsed, err := mw.ParseTokenString(token)
	require.NoError(t, err)
	return jwt.ExtractClaimsFromToken(parsed)["scope"]
}

func TestUpdateHandler_RebuildsClaims(t *testing.T) {
	assert := testify.New(t)
	db := utils.SetupServiceTest(assert)
	cfg := utils.SetupConfig()
	cfg.JWTAPITimeoutHours = 1
	log := zaptest.NewLogger(t).Sugar()

	user := models.User{
		Type:  "staff",
		Admin: lib.BoolToPtr(true),
	}
	assert.NoError(db.Create(&user).Error)

	mw, err := api.LoadMiddleware(cfg, db, log)
	require.NoError(t, err)

	token, _, err := mw.TokenGenerator(&auth.AuthenticatorPayload{
		TokenLevel: auth.TokenLevelUser,
		User:       &user,
	})
	require.NoError(t, err)
	assert.Contains(tokenScopes(t, mw, token), libauth.ScopeAdminAll)

	// Demote the user after the token was issued, refreshing should drop the admin scope
	assert.NoError(db.Model(&user).Update("admin", false).Error)

	router := gin.New()
	router.GET("/refresh", auth.UpdateHandler(mw, db))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/refresh", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	resp := utils.GetGoodResp(assert, w)
	var refreshed auth.AuthResponse
	require.NoError(t, json.Unmarshal(resp.Data, &refreshed))

	scopes := tokenScopes(t, mw, refreshed.Token)
	assert.NotContains(scopes, libauth.ScopeAdminAll)
	assert.Contains(scopes, libauth.ScopeWriteSelf)
}
