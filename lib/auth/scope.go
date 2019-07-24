package auth

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

type Scope string

// The current scopes
const (
	ScopeAdminAll      Scope = "admin:all"
	ScopeReadAll       Scope = "read:all"
	ScopeWriteSelf     Scope = "write:self"
	ScopeReadEphcatch  Scope = "read:ephcatch"
	ScopeWriteEphcatch Scope = "write:ephcatch"
	ScopeAdminFactrak  Scope = "admin:factrak"
)

// Require this endpoint to have a scope; multiple scopes mean an OR. For an AND, call this function multiple times
func RequireScopes(scopes ...Scope) func(c *gin.Context) {
	return func(c *gin.Context) {
		// If scope isn't valid, abort with error
		if !HasScope(c, scopes...) {
			services.Base.RespondError(c, lib.ErrorNoScopeAuthorization)
			return
		}

		c.Next()
	}
}

func HasScope(c *gin.Context, scopes ...Scope) bool {
	jwtScopes := c.GetStringSlice("jwtScopes")

	// Check if the scope is valid
	authed := false
	for _, scope := range scopes {
		if authed = containsString(jwtScopes, string(scope)); authed {
			break
		}
	}

	return authed
}

// ContainsString returns true if a string is present.
func containsString(s []string, v string) bool {
	for _, vv := range s {
		if vv == v {
			return true
		}
	}
	return false
}
