package auth_test

import (
	"net/http/httptest"
	"testing"

	. "github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	testify "github.com/stretchr/testify/assert"
)

func BenchmarkRequireScopes(b *testing.B) {
	scopeMiddleware := RequireScopes("admin:factrak", "admin:all")
	gin.SetMode(gin.TestMode)

	for n := 0; n < b.N; n++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Set("JWT_PAYLOAD", jwt.MapClaims{
			"exp":      1563662038,
			"id":       2,
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

func TestHasScope(t *testing.T) {
	assert := testify.New(t)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("scopes", []string{"foo", "bar"})

	// Test doesnt work
	assert.False(HasScope(c, "baz"))

	// Test works for one not in
	assert.True(HasScope(c, "foo"))

	// Test fails for nothing
	assert.False(HasScope(c))

	// Test works for two both in
	assert.True(HasScope(c, "foo", "baz"))

	// Test works for two one in
	assert.True(HasScope(c, "foo", "bar"))

	// Test works for two not in
	assert.False(HasScope(c, "hi", "world"))
}

func TestRequireScopes(t *testing.T) {
	assert := testify.New(t)

	// Test basic ability:
	f := RequireScopes("auth")

	// Fails with empty scopes
	c := NewGinCtx()
	c.Set("scopes", []string{})
	f(c)
	assert.True(c.IsAborted())

	// Fails with wrong scope
	c = NewGinCtx()
	c.Set("scopes", []string{"user"})
	f(c)
	assert.True(c.IsAborted())

	// Succeeds with correct scope
	c = NewGinCtx()
	c.Set("scopes", []string{"auth"})
	f(c)
	assert.False(c.IsAborted())

	// Succeeds with correct scope and wrong scope
	c = NewGinCtx()
	c.Set("scopes", []string{"user", "auth"})
	f(c)
	assert.False(c.IsAborted())

	// Test OR ability:
	f = RequireScopes("admin", "factrak_admin")

	// Fails with empty scopes
	c = NewGinCtx()
	c.Set("scopes", []string{})
	f(c)
	assert.True(c.IsAborted())

	// Fails with wrong scope
	c = NewGinCtx()
	c.Set("scopes", []string{"user"})
	f(c)
	assert.True(c.IsAborted())

	// Succeeds with correct scope
	c = NewGinCtx()
	c.Set("scopes", []string{"admin"})
	f(c)
	assert.False(c.IsAborted())

	// Succeeds with other correct scope
	c = NewGinCtx()
	c.Set("scopes", []string{"factrak_admin"})
	f(c)
	assert.False(c.IsAborted())

	// Succeeds with correct scope and wrong scope
	c = NewGinCtx()
	c.Set("scopes", []string{"user", "admin"})
	f(c)
	assert.False(c.IsAborted())

	// Succeeds with both correct scopes
	c = NewGinCtx()
	c.Set("scopes", []string{"factrak_admin", "admin"})
	f(c)
	assert.False(c.IsAborted())

	// Test AND ability:
	f = RequireScopes("user", "factrak")
	f2 := RequireScopes("write_self")

	// Fails with wrong scope
	c = NewGinCtx()
	c.Set("scopes", []string{"fake"})
	f(c)
	f2(c)
	assert.True(c.IsAborted())

	// Fails with good scope but then bad scope
	c = NewGinCtx()
	c.Set("scopes", []string{"user"})
	f(c)
	f2(c)
	assert.True(c.IsAborted())

	// Succeeds with correct scopes
	c = NewGinCtx()
	c.Set("scopes", []string{"factrak", "write_self"})
	f(c)
	f2(c)
	assert.False(c.IsAborted())
}

func TestCheckIDIsSelf(t *testing.T) {
	assert := testify.New(t)
	c := NewGinCtx()
	c.Set("id", uint(134))

	assert.False(CheckIDIsSelf(c, 425))

	assert.True(CheckIDIsSelf(c, 134))
}

func NewGinCtx() *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	return c
}
