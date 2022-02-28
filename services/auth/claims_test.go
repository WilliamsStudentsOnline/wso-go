package auth

import (
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib/auth"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/stretchr/testify/assert"
)

func TestRemoveBannedScope(t *testing.T) {
	scope := []string{auth.ScopeChat, auth.ScopeUsers, auth.ScopeBulletin, auth.ScopeWriteSelf, auth.ScopeFactrakFull,
		auth.ScopeFactrakLimited, auth.ScopeBulletinWrite, auth.ScopeEphmatch, auth.ScopeDormtrakWrite,
		auth.ScopeDormtrak}
	banInfo := models.BannedUser{
		Factrak:       false,
		Dormtrak:      true,
		Ephcatch:      true,
		BulletinRead:  false,
		BulletinWrite: false,
		Ephmatch:      false,
	}

	removeBannedScope(&scope, &banInfo)

	t.Log(scope)

	assert.ElementsMatch(t, scope, []string{auth.ScopeChat, auth.ScopeUsers, auth.ScopeWriteSelf,
		auth.ScopeDormtrakWrite, auth.ScopeDormtrak})
}
