package auth

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

type Scope string

// The current scopes
const (
	// General scopes:
	ScopeAdminAll     Scope = "admin:all"
	ScopeAdminFactrak Scope = "admin:factrak"
	// Allows client to do write-level requests as long as it is scoped to models involving self, not all models
	ScopeWriteSelf Scope = "write:self"

	// Service scopes. Permits clients to access services read only
	// Limited access to factrak for people with outstanding survey deficit
	ScopeFactrakLimited Scope = "service:factrak:limited"
	// Full access to factrak for people with no survey deficit
	ScopeFactrakFull Scope = "service:factrak:full"
	ScopeEphcatch    Scope = "service:ephcatch"
	ScopeBulletins   Scope = "service:bulletins"
	// This is for factrak & users
	ScopeUsers    Scope = "service:users"
	ScopeDormtrak Scope = "service:dormtrak"
	// Allows you to access other services not mentioned above
	ScopeAllOther Scope = "service:other"
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

func CheckIDIsSelf(c *gin.Context, checkSelf uint) bool {
	val, ok := c.Get("userID")
	if !ok {
		return false
	}
	return val.(uint) == checkSelf
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
