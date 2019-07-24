package auth_test

import (
	"net/http/httptest"
	"testing"

	. "github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func BenchmarkRequireScopes(b *testing.B) {
	scopeMiddleware := RequireScopes("admin:factrak", "admin:all")
	gin.SetMode(gin.TestMode)

	for n := 0; n < b.N; n++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Set("JWT_PAYLOAD", jwt.MapClaims{
			"exp": 1563662038,
			"id": 2,
			"orig_iat": 1563658438,
			"scope": []interface{}{
				"read:all",
				"write:self",
				"admin:all",
				"admin:factrak",
				"admin:factrak",
			},
		})

		scopeMiddleware(c)
	}

}