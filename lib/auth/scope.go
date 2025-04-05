package auth

import (
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
)

// The current scopes
const (
	// Global scopes:
	ScopeAdminAll = "admin:all"
	// Allows client to do write-level requests as long as it is scoped to models involving self, not all models
	ScopeWriteSelf = "write:self"

	// Service scopes. Permits clients to access services read only. If it is included with the global write-self scope,
	// allows write access to service, iff it is scoped to models owned and allowed to be edited by self.

	// Service: Factrak
	// Limited access to factrak for people with outstanding survey deficit
	ScopeFactrakLimited = "service:factrak:limited"
	// Full access to factrak for people with no survey deficit. Includes everything from ScopeFactrakLimited.
	ScopeFactrakFull = "service:factrak:full"
	// Allows admin access to factrak. This includes everything from ScopeFactrakFull, while also opening up
	// admin endpoints and allowing certain admin-level write actions (need write-self for normal actions, though).
	ScopeFactrakAdmin = "service:factrak:admin"

	// Service: Dormtrak
	// Access to dormtrak reviews, etc.
	ScopeDormtrak = "service:dormtrak"
	// Ability to create reviews, etc. (must be upperclass)
	ScopeDormtrakWrite = "service:dormtrak:write"

	ScopeEphcatch      = "service:ephcatch"
	ScopeBulletin      = "service:bulletin"
	ScopeBulletinWrite = "service:bulletin:write"
	// This is for facebook & users
	ScopeUsers = "service:users"
	// Allows you to access other services not mentioned above
	ScopeAllOther = "service:other"

	// Service: Chat
	ScopeChat = "service:chat"

	// Service: Ephmatch
	// Allows access to read/write self profile on Ephmatch. For when a user is eligible but not signed up
	ScopeEphmatch = "service:ephmatch"
	// Allows access to matches. For when a user is signed up but Ephmatch is closed
	ScopeEphmatchMatches = "service:ephmatch:matches"
	// Allows access to read profiles, write like/unlike. For when a user is signed up and Ephmatch is open
	ScopeEphmatchProfiles = "service:ephmatch:profiles"

	// Service: Goodrich
	// Allows access to read/write self goodrich orders and read goodrich menu
	ScopeGoodrich = "service:goodrich"
	// Allows access to read/write all goodrich orders and read/write goodrich menu
	ScopeGoodrichManager = "service:goodrich:manager"

	// Services: Booktrak
	ScopeBooktrak      = "service:booktrak"
	ScopeBooktrakWrite = "service:booktrak:write"
)

// Require this endpoint to have a scope; multiple scopes mean an OR. For an AND, call this function multiple times
func RequireScopes(scopes ...string) func(c *gin.Context) {
	return func(c *gin.Context) {
		// If scope isn't valid, abort with error
		if !HasScope(c, scopes...) {
			services.Base.RespondError(c, lib.ErrorNoScopeAuthorization)
			return
		}

		c.Next()
	}
}

func HasScope(c *gin.Context, scopes ...string) bool {
	jwtScopes := c.GetStringSlice("scopes")

	// Check if the scope is valid
	authed := false
	for _, scope := range scopes {
		if authed = containsString(jwtScopes, scope); authed {
			break
		}
	}

	return authed
}

func CheckIDIsSelf(c *gin.Context, checkSelf uint) bool {
	val, ok := c.Get("id")
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
